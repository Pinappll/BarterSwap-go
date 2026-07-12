package main

import (
	"context"
	"database/sql"
	"errors"
)

func InsertCreditTransaction(ctx context.Context, db querier, userID, exchangeID, montant int, txType string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO credit_transactions (user_id, exchange_id, montant, type) VALUES ($1, $2, $3, $4)`,
		userID, exchangeID, montant, txType,
	)
	return err
}

// SelectSpendAmountForExchange renvoie le montant (négatif) bloqué en
// transaction "spend" à l'acceptation. Il sert de référence pour le
// remboursement ou le transfert, plutôt que le prix courant du service.
func SelectSpendAmountForExchange(ctx context.Context, db querier, exchangeID int) (int, error) {
	var montant int
	err := db.QueryRowContext(ctx,
		`SELECT montant FROM credit_transactions WHERE exchange_id = $1 AND type = 'spend'`,
		exchangeID,
	).Scan(&montant)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrInvalidTransition
	}
	return montant, err
}

// AdjustUserBalance applique un delta au solde d'un utilisateur. Le WHERE
// empêche tout passage en négatif et sérialise les dépenses concurrentes via
// le verrou de ligne Postgres.
func AdjustUserBalance(ctx context.Context, db querier, userID, delta int) error {
	result, err := db.ExecContext(ctx,
		`UPDATE users SET credit_balance = credit_balance + $1 WHERE id = $2 AND credit_balance + $1 >= 0`,
		delta, userID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrInsufficientCredits
	}

	return nil
}
