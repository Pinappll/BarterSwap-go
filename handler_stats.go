package main

import (
	"database/sql"
	"net/http"
)

func HandleGetUserStats(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseIDParam(r, "id")
		if err != nil {
			writeError(w, err)
			return
		}

		stats, err := GetUserStats(r.Context(), db, id)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, stats)
	}
}
