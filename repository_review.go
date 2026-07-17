package main

import (
	"context"
	"database/sql"
)

func InsertReview(ctx context.Context, db *sql.DB, r *Review) error {
	query := `
		INSERT INTO reviews (exchange_id, author_id, target_id, note, commentaire)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	return db.QueryRowContext(ctx, query, r.ExchangeID, r.AuthorID, r.TargetID, r.Note, r.Commentaire).
		Scan(&r.ID, &r.CreatedAt)
}

func ExistsReviewByAuthorForExchange(ctx context.Context, db querier, exchangeID, authorID int) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM reviews WHERE exchange_id = $1 AND author_id = $2)`,
		exchangeID, authorID,
	).Scan(&exists)
	return exists, err
}

func SelectReviewsByTargetID(ctx context.Context, db *sql.DB, targetID int) ([]Review, error) {
	query := `
		SELECT id, exchange_id, author_id, target_id, note, COALESCE(commentaire, ''), created_at
		FROM reviews
		WHERE target_id = $1
		ORDER BY created_at DESC
	`

	rows, err := db.QueryContext(ctx, query, targetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := []Review{}
	for rows.Next() {
		var r Review
		if err := rows.Scan(&r.ID, &r.ExchangeID, &r.AuthorID, &r.TargetID, &r.Note, &r.Commentaire, &r.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}

	return reviews, rows.Err()
}

func SelectReviewsByServiceID(ctx context.Context, db *sql.DB, serviceID int) ([]Review, error) {
	query := `
		SELECT r.id, r.exchange_id, r.author_id, r.target_id, r.note, COALESCE(r.commentaire, ''), r.created_at
		FROM reviews r
		JOIN exchanges e ON e.id = r.exchange_id
		WHERE e.service_id = $1 AND r.target_id = e.owner_id
		ORDER BY r.created_at DESC
	`

	rows, err := db.QueryContext(ctx, query, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := []Review{}
	for rows.Next() {
		var r Review
		if err := rows.Scan(&r.ID, &r.ExchangeID, &r.AuthorID, &r.TargetID, &r.Note, &r.Commentaire, &r.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}

	return reviews, rows.Err()
}
