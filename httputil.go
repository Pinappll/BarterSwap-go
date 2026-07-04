package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

// writeJSON écrit une réponse JSON avec le code HTTP donné.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// decodeJSON décode le corps de la requête, en normalisant toute erreur
// de parsing en ErrInvalidInput.
func decodeJSON(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return ErrInvalidInput
	}
	return nil
}

// writeError traduit une erreur métier (sentinelle) en réponse HTTP avec le
// code de statut approprié. Point unique de correspondance erreur -> code
// HTTP pour garder des réponses cohérentes sur tous les endpoints.
func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError

	switch {
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ErrUnauthorized):
		status = http.StatusUnauthorized
	case errors.Is(err, ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, ErrServiceUnavailable), errors.Is(err, ErrInvalidTransition):
		status = http.StatusConflict
	case errors.Is(err, ErrInvalidInput),
		errors.Is(err, ErrSkillMissing),
		errors.Is(err, ErrSelfExchange),
		errors.Is(err, ErrInsufficientCredits),
		errors.Is(err, ErrReviewNotAllowed):
		status = http.StatusBadRequest
	}

	writeJSON(w, status, map[string]string{"error": err.Error()})
}

// parseIDParam extrait et valide un identifiant numérique positif depuis un
// segment de route (ex: {id}).
func parseIDParam(r *http.Request, name string) (int, error) {
	id, err := strconv.Atoi(r.PathValue(name))
	if err != nil || id <= 0 {
		return 0, ErrInvalidInput
	}
	return id, nil
}
