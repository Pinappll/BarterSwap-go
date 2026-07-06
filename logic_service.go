package main

import (
	"context"
	"database/sql"
	"strings"
)

var categoriesValides = map[string]bool{
	"Informatique": true,
	"Jardinage":    true,
	"Bricolage":    true,
	"Cuisine":      true,
	"Musique":      true,
	"Langues":      true,
	"Sport":        true,
	"Tutorat":      true,
	"Déménagement": true,
	"Photographie": true,
	"Animalier":    true,
	"Couture":      true,
	"Autre":        true,
}

func validateServiceInput(s Service) error {
	if strings.TrimSpace(s.Titre) == "" {
		return ErrInvalidInput
	}
	if !categoriesValides[s.Categorie] {
		return ErrInvalidInput
	}
	if s.DureeMinutes <= 0 || s.Credits <= 0 {
		return ErrInvalidInput
	}
	return nil
}

// userHasSkill vérifie que la liste de compétences contient une compétence
// dont le nom correspond à la catégorie de service visée (comparaison
// insensible à la casse).
func userHasSkill(skills []Skill, categorie string) bool {
	for _, sk := range skills {
		if strings.EqualFold(sk.Nom, categorie) {
			return true
		}
	}
	return false
}

// CreateService publie une annonce pour providerID, à condition qu'il
// possède une compétence correspondant à la catégorie visée.
func CreateService(ctx context.Context, db *sql.DB, providerID int, input Service) (*Service, error) {
	if err := validateServiceInput(input); err != nil {
		return nil, err
	}

	skills, err := SelectSkillsByUserID(ctx, db, providerID)
	if err != nil {
		return nil, err
	}
	if !userHasSkill(skills, input.Categorie) {
		return nil, ErrSkillMissing
	}

	service := &Service{
		ProviderID:   providerID,
		Titre:        input.Titre,
		Description:  input.Description,
		Categorie:    input.Categorie,
		DureeMinutes: input.DureeMinutes,
		Credits:      input.Credits,
		Ville:        input.Ville,
		Actif:        true,
	}
	if err := InsertService(ctx, db, service); err != nil {
		return nil, err
	}

	return service, nil
}

func GetService(ctx context.Context, db *sql.DB, id int) (*Service, error) {
	return SelectServiceByID(ctx, db, id)
}

func ListServices(ctx context.Context, db *sql.DB, filters ServiceFilters) ([]Service, error) {
	return SelectServices(ctx, db, filters)
}

// UpdateServiceListing modifie une annonce existante, à condition que
// requesterID en soit bien le propriétaire et que la nouvelle catégorie
// corresponde toujours à une compétence qu'il possède.
func UpdateServiceListing(ctx context.Context, db *sql.DB, id, requesterID int, update Service) (*Service, error) {
	existing, err := SelectServiceByID(ctx, db, id)
	if err != nil {
		return nil, err
	}
	if err := requireOwner(existing.ProviderID, requesterID); err != nil {
		return nil, err
	}
	if err := validateServiceInput(update); err != nil {
		return nil, err
	}

	skills, err := SelectSkillsByUserID(ctx, db, existing.ProviderID)
	if err != nil {
		return nil, err
	}
	if !userHasSkill(skills, update.Categorie) {
		return nil, ErrSkillMissing
	}

	existing.Titre = update.Titre
	existing.Description = update.Description
	existing.Categorie = update.Categorie
	existing.DureeMinutes = update.DureeMinutes
	existing.Credits = update.Credits
	existing.Ville = update.Ville

	if err := UpdateService(ctx, db, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

// DeleteServiceListing supprime une annonce, à condition que requesterID en
// soit bien le propriétaire.
func DeleteServiceListing(ctx context.Context, db *sql.DB, id, requesterID int) error {
	existing, err := SelectServiceByID(ctx, db, id)
	if err != nil {
		return err
	}
	if err := requireOwner(existing.ProviderID, requesterID); err != nil {
		return err
	}

	return DeleteService(ctx, db, id)
}
