package main

import (
	"context"
	"database/sql"
)

// canTransition définit les transitions de statut autorisées pour un échange.
//
//	pending  -> accepted | rejected | cancelled
//	accepted -> completed | cancelled
func canTransition(from, to string) bool {
	switch from {
	case "pending":
		return to == "accepted" || to == "rejected" || to == "cancelled"
	case "accepted":
		return to == "completed" || to == "cancelled"
	default:
		return false
	}
}

// canPerformTransition vérifie que actingUserID a le droit de déclencher la
// transition demandée : l'offreur accepte ou refuse, le demandeur marque comme
// terminé (pour qu'il ne s'auto-valide pas ses crédits), et les deux peuvent
// annuler.
func canPerformTransition(exchange Exchange, actingUserID int, to string) error {
	switch to {
	case "accepted", "rejected":
		if actingUserID != exchange.OwnerID {
			return ErrForbidden
		}
	case "completed":
		if actingUserID != exchange.RequesterID {
			return ErrForbidden
		}
	case "cancelled":
		if actingUserID != exchange.RequesterID && actingUserID != exchange.OwnerID {
			return ErrForbidden
		}
	}
	return nil
}

// CreateExchange enregistre une demande d'échange sur un service (pas
// d'auto-échange, crédits suffisants). Les crédits ne sont bloqués qu'à
// l'acceptation.
func CreateExchange(ctx context.Context, db *sql.DB, requesterID, serviceID int) (*Exchange, error) {
	if serviceID <= 0 {
		return nil, ErrInvalidInput
	}

	service, err := SelectServiceByID(ctx, db, serviceID)
	if err != nil {
		return nil, err
	}
	if !service.Actif {
		return nil, ErrServiceUnavailable
	}
	if service.ProviderID == requesterID {
		return nil, ErrSelfExchange
	}

	requester, err := SelectUserByID(ctx, db, requesterID)
	if err != nil {
		return nil, err
	}
	if requester.CreditBalance < service.Credits {
		return nil, ErrInsufficientCredits
	}

	exchange := &Exchange{
		ServiceID:   serviceID,
		RequesterID: requesterID,
		OwnerID:     service.ProviderID,
	}
	if err := InsertExchange(ctx, db, exchange); err != nil {
		return nil, err
	}

	return exchange, nil
}

// GetExchangeForParticipant renvoie l'échange si actingUserID en est le
// demandeur ou l'offreur, sinon ErrForbidden.
func GetExchangeForParticipant(ctx context.Context, db *sql.DB, id, actingUserID int) (*Exchange, error) {
	exchange, err := SelectExchangeByID(ctx, db, id)
	if err != nil {
		return nil, err
	}
	if actingUserID != exchange.RequesterID && actingUserID != exchange.OwnerID {
		return nil, ErrForbidden
	}
	return exchange, nil
}

func ListExchanges(ctx context.Context, db *sql.DB, userID int, status string) ([]Exchange, error) {
	return SelectExchangesForUser(ctx, db, userID, status)
}

// AcceptExchange passe l'échange en "accepted" et bloque les crédits du
// demandeur : ils sont débités et journalisés en transaction "spend", sans
// encore être crédités à l'offreur. Le tout dans une transaction verrouillée
// sur la ligne exchange, pour rester correct en accès concurrent.
func AcceptExchange(ctx context.Context, db *sql.DB, id, actingUserID int) (*Exchange, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	exchange, err := SelectExchangeForUpdate(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if err := canPerformTransition(*exchange, actingUserID, "accepted"); err != nil {
		return nil, err
	}
	if !canTransition(exchange.Status, "accepted") {
		return nil, ErrInvalidTransition
	}

	service, err := SelectServiceByID(ctx, tx, exchange.ServiceID)
	if err != nil {
		return nil, err
	}

	if err := AdjustUserBalance(ctx, tx, exchange.RequesterID, -service.Credits); err != nil {
		return nil, err
	}
	if err := InsertCreditTransaction(ctx, tx, exchange.RequesterID, exchange.ID, -service.Credits, "spend"); err != nil {
		return nil, err
	}
	if err := UpdateExchangeStatus(ctx, tx, exchange.ID, "accepted"); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	exchange.Status = "accepted"
	return exchange, nil
}

// RejectExchange refuse un échange "pending". Aucun crédit n'ayant été bloqué
// à ce stade, il n'y a rien à rembourser.
func RejectExchange(ctx context.Context, db *sql.DB, id, actingUserID int) (*Exchange, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	exchange, err := SelectExchangeForUpdate(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if err := canPerformTransition(*exchange, actingUserID, "rejected"); err != nil {
		return nil, err
	}
	if !canTransition(exchange.Status, "rejected") {
		return nil, ErrInvalidTransition
	}

	if err := UpdateExchangeStatus(ctx, tx, exchange.ID, "rejected"); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	exchange.Status = "rejected"
	return exchange, nil
}

// CompleteExchange crédite définitivement l'offreur des crédits bloqués. Le
// montant transféré est celui journalisé à l'acceptation, pas le prix courant
// du service, qui a pu changer depuis.
func CompleteExchange(ctx context.Context, db *sql.DB, id, actingUserID int) (*Exchange, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	exchange, err := SelectExchangeForUpdate(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if err := canPerformTransition(*exchange, actingUserID, "completed"); err != nil {
		return nil, err
	}
	if !canTransition(exchange.Status, "completed") {
		return nil, ErrInvalidTransition
	}

	spent, err := SelectSpendAmountForExchange(ctx, tx, exchange.ID)
	if err != nil {
		return nil, err
	}
	earned := -spent

	if err := AdjustUserBalance(ctx, tx, exchange.OwnerID, earned); err != nil {
		return nil, err
	}
	if err := InsertCreditTransaction(ctx, tx, exchange.OwnerID, exchange.ID, earned, "earn"); err != nil {
		return nil, err
	}
	if err := UpdateExchangeStatus(ctx, tx, exchange.ID, "completed"); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	exchange.Status = "completed"
	return exchange, nil
}

// CancelExchange annule un échange et rembourse le demandeur si l'échange
// avait déjà été accepté. Sur un échange "pending", rien n'était bloqué.
func CancelExchange(ctx context.Context, db *sql.DB, id, actingUserID int) (*Exchange, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	exchange, err := SelectExchangeForUpdate(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if err := canPerformTransition(*exchange, actingUserID, "cancelled"); err != nil {
		return nil, err
	}
	if !canTransition(exchange.Status, "cancelled") {
		return nil, ErrInvalidTransition
	}

	if exchange.Status == "accepted" {
		spent, err := SelectSpendAmountForExchange(ctx, tx, exchange.ID)
		if err != nil {
			return nil, err
		}
		refund := -spent

		if err := AdjustUserBalance(ctx, tx, exchange.RequesterID, refund); err != nil {
			return nil, err
		}
		if err := InsertCreditTransaction(ctx, tx, exchange.RequesterID, exchange.ID, refund, "refund"); err != nil {
			return nil, err
		}
	}

	if err := UpdateExchangeStatus(ctx, tx, exchange.ID, "cancelled"); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	exchange.Status = "cancelled"
	return exchange, nil
}
