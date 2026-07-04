package main

import (
	"database/sql"
	"net/http"
)

func HandleGetSkills(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseIDParam(r, "id")
		if err != nil {
			writeError(w, err)
			return
		}

		skills, err := GetUserSkills(r.Context(), db, id)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, skills)
	}
}

func HandlePutSkills(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseIDParam(r, "id")
		if err != nil {
			writeError(w, err)
			return
		}

		requesterID, ok := userIDFromContext(r.Context())
		if !ok {
			writeError(w, ErrUnauthorized)
			return
		}

		var skills []Skill
		if err := decodeJSON(r, &skills); err != nil {
			writeError(w, err)
			return
		}

		updated, err := SetUserSkills(r.Context(), db, id, requesterID, skills)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, updated)
	}
}
