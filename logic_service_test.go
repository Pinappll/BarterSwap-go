package main

import "testing"

func TestValidateServiceInput(t *testing.T) {
	base := Service{Titre: "Cours", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2}

	cases := []struct {
		name    string
		mutate  func(s Service) Service
		wantErr bool
	}{
		{"valide", func(s Service) Service { return s }, false},
		{"titre vide", func(s Service) Service { s.Titre = ""; return s }, true},
		{"titre composé uniquement d'espaces", func(s Service) Service { s.Titre = "   "; return s }, true},
		{"catégorie hors liste fermée", func(s Service) Service { s.Categorie = "Yoga"; return s }, true},
		{"durée nulle", func(s Service) Service { s.DureeMinutes = 0; return s }, true},
		{"durée négative", func(s Service) Service { s.DureeMinutes = -10; return s }, true},
		{"crédits nuls", func(s Service) Service { s.Credits = 0; return s }, true},
		{"crédits négatifs", func(s Service) Service { s.Credits = -1; return s }, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateServiceInput(tc.mutate(base))
			if (err != nil) != tc.wantErr {
				t.Errorf("erreur = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestUserHasSkill(t *testing.T) {
	skills := []Skill{{Nom: "Jardinage", Niveau: "expert"}, {Nom: "Cuisine", Niveau: "débutant"}}

	cases := []struct {
		name      string
		categorie string
		want      bool
	}{
		{"correspondance exacte", "Jardinage", true},
		{"correspondance insensible à la casse", "jardinage", true},
		{"compétence absente", "Musique", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := userHasSkill(skills, tc.categorie); got != tc.want {
				t.Errorf("userHasSkill(_, %q) = %v, attendu %v", tc.categorie, got, tc.want)
			}
		})
	}
}
