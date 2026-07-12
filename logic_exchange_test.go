package main

import "testing"

func TestCanTransition(t *testing.T) {
	cases := []struct {
		from string
		to   string
		want bool
	}{
		{"pending", "accepted", true},
		{"pending", "rejected", true},
		{"pending", "cancelled", true},
		{"pending", "completed", false},
		{"accepted", "completed", true},
		{"accepted", "cancelled", true},
		{"accepted", "accepted", false},
		{"accepted", "rejected", false},
		{"completed", "cancelled", false},
		{"rejected", "accepted", false},
		{"cancelled", "accepted", false},
	}

	for _, tc := range cases {
		t.Run(tc.from+"->"+tc.to, func(t *testing.T) {
			if got := canTransition(tc.from, tc.to); got != tc.want {
				t.Errorf("canTransition(%q, %q) = %v, attendu %v", tc.from, tc.to, got, tc.want)
			}
		})
	}
}

func TestCanPerformTransition(t *testing.T) {
	exchange := Exchange{RequesterID: 1, OwnerID: 2}

	cases := []struct {
		name    string
		userID  int
		to      string
		wantErr bool
	}{
		{"owner accepte", 2, "accepted", false},
		{"requester accepte -> interdit", 1, "accepted", true},
		{"owner refuse", 2, "rejected", false},
		{"requester refuse -> interdit", 1, "rejected", true},
		{"requester complète", 1, "completed", false},
		{"owner complète -> interdit (auto-validation)", 2, "completed", true},
		{"requester annule", 1, "cancelled", false},
		{"owner annule", 2, "cancelled", false},
		{"tiers annule -> interdit", 999, "cancelled", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := canPerformTransition(exchange, tc.userID, tc.to)
			if (err != nil) != tc.wantErr {
				t.Errorf("canPerformTransition(_, %d, %q) erreur = %v, wantErr %v", tc.userID, tc.to, err, tc.wantErr)
			}
		})
	}
}
