package main

import (
	"context"
	"database/sql"
	"errors"
)

// SelectUserStats agrège en une seule requête les données de plusieurs
// tables (services, exchanges, reviews, credit_transactions) pour un
// utilisateur donné. total_depense nette les "spend" par leurs "refund"
// correspondants : un échange annulé/refusé ne compte pas comme dépensé.
func SelectUserStats(ctx context.Context, db *sql.DB, userID int) (*UserStats, error) {
	query := `
		SELECT
			u.credit_balance,
			(SELECT COUNT(*) FROM services WHERE provider_id = u.id AND actif = true) AS services_actifs,
			(SELECT COUNT(*) FROM exchanges WHERE (requester_id = u.id OR owner_id = u.id) AND status = 'completed') AS echanges_completes,
			COALESCE((SELECT AVG(note) FROM reviews WHERE target_id = u.id), 0) AS note_moyenne,
			(SELECT COUNT(*) FROM reviews WHERE target_id = u.id) AS nb_avis,
			COALESCE((SELECT SUM(montant) FROM credit_transactions WHERE user_id = u.id AND type = 'earn'), 0) AS total_gagne,
			COALESCE((SELECT ABS(SUM(montant)) FROM credit_transactions WHERE user_id = u.id AND type IN ('spend', 'refund')), 0) AS total_depense
		FROM users u
		WHERE u.id = $1
	`

	stats := UserStats{UserID: userID}
	err := db.QueryRowContext(ctx, query, userID).Scan(
		&stats.CreditBalance,
		&stats.ServicesActifs,
		&stats.EchangesCompletes,
		&stats.NoteMoyenne,
		&stats.NbAvis,
		&stats.TotalGagne,
		&stats.TotalDepense,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
