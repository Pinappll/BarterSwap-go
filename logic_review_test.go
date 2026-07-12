package main

import (
	"fmt"
	"testing"
)

func TestValidateNote(t *testing.T) {
	cases := []struct {
		note    int
		wantErr bool
	}{
		{0, true},
		{1, false},
		{3, false},
		{5, false},
		{6, true},
		{-1, true},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("note=%d", tc.note), func(t *testing.T) {
			err := validateNote(tc.note)
			if (err != nil) != tc.wantErr {
				t.Errorf("validateNote(%d) erreur = %v, wantErr %v", tc.note, err, tc.wantErr)
			}
		})
	}
}
