package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteError(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{ErrNotFound, http.StatusNotFound},
		{ErrUnauthorized, http.StatusUnauthorized},
		{ErrForbidden, http.StatusForbidden},
		{ErrServiceUnavailable, http.StatusConflict},
		{ErrInvalidTransition, http.StatusConflict},
		{ErrInvalidInput, http.StatusBadRequest},
		{ErrSkillMissing, http.StatusBadRequest},
		{ErrSelfExchange, http.StatusBadRequest},
		{ErrInsufficientCredits, http.StatusBadRequest},
		{ErrReviewNotAllowed, http.StatusBadRequest},
		{errors.New("erreur inconnue"), http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.err.Error(), func(t *testing.T) {
			rec := httptest.NewRecorder()
			writeError(rec, tc.err)
			if rec.Code != tc.want {
				t.Errorf("code = %d, attendu %d pour l'erreur %q", rec.Code, tc.want, tc.err)
			}
		})
	}
}

func TestParseIDParam(t *testing.T) {
	cases := []struct {
		name    string
		value   string
		wantID  int
		wantErr bool
	}{
		{"id valide", "42", 42, false},
		{"id non numérique", "abc", 0, true},
		{"id négatif", "-1", 0, true},
		{"id nul", "0", 0, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.SetPathValue("id", tc.value)

			id, err := parseIDParam(req, "id")
			if (err != nil) != tc.wantErr {
				t.Fatalf("erreur = %v, wantErr %v", err, tc.wantErr)
			}
			if !tc.wantErr && id != tc.wantID {
				t.Errorf("id = %d, attendu %d", id, tc.wantID)
			}
		})
	}
}
