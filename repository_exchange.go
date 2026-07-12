package main

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

// InsertExchange crée une demande d'échange au statut "pending". L'index unique
// partiel one_active_exchange_per_service (schema.sql) garantit un seul échange
// actif par service ; sa violation est traduite en ErrServiceUnavailable.
func InsertExchange(ctx context.Context, db *sql.DB, e *Exchange) error {
	query := `
		INSERT INTO exchanges (service_id, requester_id, owner_id, status)
		VALUES ($1, $2, $3, 'pending')
		RETURNING id, status, created_at, updated_at
	`

	err := db.QueryRowContext(ctx, query, e.ServiceID, e.RequesterID, e.OwnerID).
		Scan(&e.ID, &e.Status, &e.CreatedAt, &e.UpdatedAt)

	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return ErrServiceUnavailable
	}
	return err
}

func SelectExchangeByID(ctx context.Context, db querier, id int) (*Exchange, error) {
	query := `
		SELECT id, service_id, requester_id, owner_id, status, created_at, updated_at
		FROM exchanges
		WHERE id = $1
	`

	var e Exchange
	err := db.QueryRowContext(ctx, query, id).Scan(
		&e.ID, &e.ServiceID, &e.RequesterID, &e.OwnerID, &e.Status, &e.CreatedAt, &e.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &e, nil
}

// SelectExchangeForUpdate lit et verrouille la ligne (SELECT ... FOR UPDATE)
// pour la durée de la transaction, ce qui sérialise deux actions concurrentes
// sur le même échange au niveau de la base.
func SelectExchangeForUpdate(ctx context.Context, tx *sql.Tx, id int) (*Exchange, error) {
	query := `
		SELECT id, service_id, requester_id, owner_id, status, created_at, updated_at
		FROM exchanges
		WHERE id = $1
		FOR UPDATE
	`

	var e Exchange
	err := tx.QueryRowContext(ctx, query, id).Scan(
		&e.ID, &e.ServiceID, &e.RequesterID, &e.OwnerID, &e.Status, &e.CreatedAt, &e.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &e, nil
}

func UpdateExchangeStatus(ctx context.Context, tx *sql.Tx, id int, status string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE exchanges SET status = $1, updated_at = now() WHERE id = $2`,
		status, id,
	)
	return err
}

// SelectExchangesForUser liste les échanges où l'utilisateur est demandeur ou
// offreur, avec un filtre optionnel par statut.
func SelectExchangesForUser(ctx context.Context, db *sql.DB, userID int, status string) ([]Exchange, error) {
	query := `
		SELECT id, service_id, requester_id, owner_id, status, created_at, updated_at
		FROM exchanges
		WHERE (requester_id = $1 OR owner_id = $1)
	`
	args := []any{userID}
	if status != "" {
		args = append(args, status)
		query += " AND status = $2"
	}
	query += " ORDER BY created_at DESC"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	exchanges := []Exchange{}
	for rows.Next() {
		var e Exchange
		if err := rows.Scan(&e.ID, &e.ServiceID, &e.RequesterID, &e.OwnerID, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		exchanges = append(exchanges, e)
	}

	return exchanges, rows.Err()
}
