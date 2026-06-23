package main

import (
	"database/sql"
)

func CreateUser(db *sql.DB, user *User) error {
	query := `
		INSERT INTO users (pseudo, bio, ville, credit_balance)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := db.QueryRow(
		query,
		user.Pseudo,
		user.Bio,
		user.Ville,
		user.CreditBalance,
	).Scan(&user.ID, &user.CreatedAt)

	return err
}
