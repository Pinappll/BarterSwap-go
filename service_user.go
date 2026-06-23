package main

import (
	"database/sql"
)


func CreateUserService(db *sql.DB, user *User) error {

	if user.Pseudo == "" {
		return ErrInvalidInput
	}

	user.CreditBalance = 10

	return CreateUser(db, user)
}
