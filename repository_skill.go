package main

import (
	"context"
	"database/sql"
)

func SelectSkillsByUserID(ctx context.Context, db *sql.DB, userID int) ([]Skill, error) {
	query := `SELECT nom, niveau FROM skills WHERE user_id = $1 ORDER BY nom`

	rows, err := db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skills := []Skill{}
	for rows.Next() {
		var s Skill
		if err := rows.Scan(&s.Nom, &s.Niveau); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}

	return skills, rows.Err()
}

// ReplaceSkills écrase l'intégralité des compétences d'un utilisateur par la
// liste donnée (sémantique PUT : pas d'ajout/fusion individuelle).
func ReplaceSkills(ctx context.Context, db *sql.DB, userID int, skills []Skill) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM skills WHERE user_id = $1`, userID); err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO skills (user_id, nom, niveau) VALUES ($1, $2, $3)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range skills {
		if _, err := stmt.ExecContext(ctx, userID, s.Nom, s.Niveau); err != nil {
			return err
		}
	}

	return tx.Commit()
}
