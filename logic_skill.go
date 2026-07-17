package main

import (
	"context"
	"database/sql"
	"strings"
)

var niveauxValides = map[string]bool{
	"débutant":      true,
	"intermédiaire": true,
	"expert":        true,
}

func validateSkills(skills []Skill) error {
	for _, s := range skills {
		if strings.TrimSpace(s.Nom) == "" || !niveauxValides[s.Niveau] {
			return ErrInvalidInput
		}
	}
	return nil
}

func GetUserSkills(ctx context.Context, db *sql.DB, userID int) ([]Skill, error) {
	if _, err := SelectUserByID(ctx, db, userID); err != nil {
		return nil, err
	}
	return SelectSkillsByUserID(ctx, db, userID)
}

func SetUserSkills(ctx context.Context, db *sql.DB, targetID, requesterID int, skills []Skill) ([]Skill, error) {
	if err := requireOwner(targetID, requesterID); err != nil {
		return nil, err
	}
	if _, err := SelectUserByID(ctx, db, targetID); err != nil {
		return nil, err
	}
	if err := validateSkills(skills); err != nil {
		return nil, err
	}

	if err := ReplaceSkills(ctx, db, targetID, skills); err != nil {
		return nil, err
	}

	return skills, nil
}
