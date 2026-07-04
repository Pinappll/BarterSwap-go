package main

import "testing"

func TestValidatePseudo(t *testing.T) {
	cases := []struct {
		name    string
		pseudo  string
		wantErr bool
	}{
		{"pseudo valide", "joshua", false},
		{"pseudo vide", "", true},
		{"pseudo composé uniquement d'espaces", "   ", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePseudo(tc.pseudo)
			if (err != nil) != tc.wantErr {
				t.Errorf("validatePseudo(%q) erreur = %v, wantErr %v", tc.pseudo, err, tc.wantErr)
			}
		})
	}
}
