package main

import (
	"context"
	"database/sql"
	"errors"
)

func InsertUser(ctx context.Context, db *sql.DB, user *User) error {
	query := `
		INSERT INTO users (pseudo, bio, ville, credit_balance)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := db.QueryRowContext(
		ctx,
		query,
		user.Pseudo,
		user.Bio,
		user.Ville,
		user.CreditBalance,
	).Scan(&user.ID, &user.CreatedAt)

	return err
}

func SelectUserByID(ctx context.Context, db *sql.DB, id int) (*User, error) {
	query := `
		SELECT id, pseudo, COALESCE(bio, ''), COALESCE(ville, ''), credit_balance, created_at
		FROM users
		WHERE id = $1
	`

	var user User
	err := db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Pseudo, &user.Bio, &user.Ville, &user.CreditBalance, &user.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func UpdateUser(ctx context.Context, db *sql.DB, user *User) error {
	query := `
		UPDATE users
		SET pseudo = $1, bio = $2, ville = $3
		WHERE id = $4
		RETURNING credit_balance, created_at
	`

	err := db.QueryRowContext(ctx, query, user.Pseudo, user.Bio, user.Ville, user.ID).
		Scan(&user.CreditBalance, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
