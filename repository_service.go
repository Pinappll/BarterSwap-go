package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ServiceFilters porte les filtres optionnels de GET /api/services. Un champ
// vide signifie "pas de filtre sur ce critère".
type ServiceFilters struct {
	Categorie string
	Ville     string
	Search    string
}

func InsertService(ctx context.Context, db *sql.DB, s *Service) error {
	query := `
		INSERT INTO services (provider_id, titre, description, categorie, duree_minutes, credits, ville, actif)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	return db.QueryRowContext(ctx, query,
		s.ProviderID, s.Titre, s.Description, s.Categorie, s.DureeMinutes, s.Credits, s.Ville, s.Actif,
	).Scan(&s.ID, &s.CreatedAt)
}

func SelectServiceByID(ctx context.Context, db querier, id int) (*Service, error) {
	query := `
		SELECT id, provider_id, titre, COALESCE(description, ''), categorie,
		       duree_minutes, credits, COALESCE(ville, ''), actif, created_at
		FROM services
		WHERE id = $1
	`

	var s Service
	err := db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.ProviderID, &s.Titre, &s.Description, &s.Categorie,
		&s.DureeMinutes, &s.Credits, &s.Ville, &s.Actif, &s.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// UpdateService remplace les champs descriptifs/commerciaux d'une annonce.
// "actif" n'est volontairement pas modifiable ici : aucune fonctionnalité de
// mise en pause n'est demandée par le sujet, seule la suppression existe.
func UpdateService(ctx context.Context, db *sql.DB, s *Service) error {
	query := `
		UPDATE services
		SET titre = $1, description = $2, categorie = $3, duree_minutes = $4, credits = $5, ville = $6
		WHERE id = $7
		RETURNING created_at
	`

	err := db.QueryRowContext(ctx, query, s.Titre, s.Description, s.Categorie, s.DureeMinutes, s.Credits, s.Ville, s.ID).
		Scan(&s.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func DeleteService(ctx context.Context, db *sql.DB, id int) error {
	result, err := db.ExecContext(ctx, `DELETE FROM services WHERE id = $1`, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// SelectServices liste les annonces actives, avec filtrage serveur optionnel
// par catégorie (égalité), ville (insensible à la casse) et recherche
// textuelle (titre/description). Les conditions sont ajoutées
// dynamiquement mais toujours paramétrées, jamais concaténées dans la
// requête, pour éviter toute injection SQL.
func SelectServices(ctx context.Context, db *sql.DB, f ServiceFilters) ([]Service, error) {
	query := `
		SELECT id, provider_id, titre, COALESCE(description, ''), categorie,
		       duree_minutes, credits, COALESCE(ville, ''), actif, created_at
		FROM services
		WHERE actif = true
	`
	var args []any

	if f.Categorie != "" {
		args = append(args, f.Categorie)
		query += fmt.Sprintf(" AND categorie = $%d", len(args))
	}
	if f.Ville != "" {
		args = append(args, f.Ville)
		query += fmt.Sprintf(" AND ville ILIKE $%d", len(args))
	}
	if f.Search != "" {
		args = append(args, "%"+f.Search+"%")
		query += fmt.Sprintf(" AND (titre ILIKE $%d OR description ILIKE $%d)", len(args), len(args))
	}
	query += " ORDER BY created_at DESC"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	services := []Service{}
	for rows.Next() {
		var s Service
		if err := rows.Scan(
			&s.ID, &s.ProviderID, &s.Titre, &s.Description, &s.Categorie,
			&s.DureeMinutes, &s.Credits, &s.Ville, &s.Actif, &s.CreatedAt,
		); err != nil {
			return nil, err
		}
		services = append(services, s)
	}

	return services, rows.Err()
}
