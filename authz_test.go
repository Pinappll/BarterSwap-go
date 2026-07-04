package main

import "testing"

func TestRequireOwner(t *testing.T) {
	cases := []struct {
		name        string
		targetID    int
		requesterID int
		wantErr     bool
	}{
		{"même utilisateur", 1, 1, false},
		{"utilisateur différent", 1, 2, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := requireOwner(tc.targetID, tc.requesterID)
			if (err != nil) != tc.wantErr {
				t.Errorf("requireOwner(%d, %d) erreur = %v, wantErr %v", tc.targetID, tc.requesterID, err, tc.wantErr)
			}
		})
	}
}
