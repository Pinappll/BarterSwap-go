package main

import "testing"

func TestValidateSkills(t *testing.T) {
	cases := []struct {
		name    string
		skills  []Skill
		wantErr bool
	}{
		{"liste vide (efface les compétences)", []Skill{}, false},
		{"niveau valide", []Skill{{Nom: "Jardinage", Niveau: "expert"}}, false},
		{"plusieurs compétences valides", []Skill{
			{Nom: "Jardinage", Niveau: "expert"},
			{Nom: "Cuisine", Niveau: "débutant"},
		}, false},
		{"niveau invalide", []Skill{{Nom: "Jardinage", Niveau: "pro"}}, true},
		{"nom vide", []Skill{{Nom: "", Niveau: "expert"}}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateSkills(tc.skills)
			if (err != nil) != tc.wantErr {
				t.Errorf("validateSkills(%v) erreur = %v, wantErr %v", tc.skills, err, tc.wantErr)
			}
		})
	}
}
