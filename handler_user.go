package main

import (
	"database/sql"
	"net/http"
)

func HandleCreateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user User

		if err := decodeJSON(r, &user); err != nil {
			writeError(w, err)
			return
		}

		if err := RegisterUser(r.Context(), db, &user); err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusCreated, user)
	}
}

func HandleGetUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseIDParam(r, "id")
		if err != nil {
			writeError(w, err)
			return
		}

		user, err := GetUserProfile(r.Context(), db, id)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, user)
	}
}

func HandleUpdateUser(db *sql.DB) http.HandlerFunc {
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

		var update User
		if err := decodeJSON(r, &update); err != nil {
			writeError(w, err)
			return
		}

		user, err := UpdateUserProfile(r.Context(), db, id, requesterID, update)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, user)
	}
}
