package main

import (
	"context"
	"database/sql"
	"strings"
)

const creditBienvenue = 10

func validatePseudo(pseudo string) error {
	if strings.TrimSpace(pseudo) == "" {
		return ErrInvalidInput
	}
	return nil
}

func RegisterUser(ctx context.Context, db *sql.DB, user *User) error {
	if err := validatePseudo(user.Pseudo); err != nil {
		return err
	}

	user.CreditBalance = creditBienvenue

	return InsertUser(ctx, db, user)
}

func GetUserProfile(ctx context.Context, db *sql.DB, id int) (*User, error) {
	user, err := SelectUserByID(ctx, db, id)
	if err != nil {
		return nil, err
	}

	skills, err := SelectSkillsByUserID(ctx, db, id)
	if err != nil {
		return nil, err
	}
	user.Skills = skills

	return user, nil
}

func UpdateUserProfile(ctx context.Context, db *sql.DB, targetID, requesterID int, update User) (*User, error) {
	if err := requireOwner(targetID, requesterID); err != nil {
		return nil, err
	}
	if err := validatePseudo(update.Pseudo); err != nil {
		return nil, err
	}

	user := &User{ID: targetID, Pseudo: update.Pseudo, Bio: update.Bio, Ville: update.Ville}
	if err := UpdateUser(ctx, db, user); err != nil {
		return nil, err
	}

	return user, nil
}
