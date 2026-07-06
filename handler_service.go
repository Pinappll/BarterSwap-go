package main

import (
	"database/sql"
	"net/http"
)

func HandleListServices(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filters := ServiceFilters{
			Categorie: r.URL.Query().Get("categorie"),
			Ville:     r.URL.Query().Get("ville"),
			Search:    r.URL.Query().Get("search"),
		}

		services, err := ListServices(r.Context(), db, filters)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, services)
	}
}

func HandleCreateService(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requesterID, ok := userIDFromContext(r.Context())
		if !ok {
			writeError(w, ErrUnauthorized)
			return
		}

		var input Service
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, err)
			return
		}

		service, err := CreateService(r.Context(), db, requesterID, input)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusCreated, service)
	}
}

func HandleGetService(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseIDParam(r, "id")
		if err != nil {
			writeError(w, err)
			return
		}

		service, err := GetService(r.Context(), db, id)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, service)
	}
}

func HandleUpdateService(db *sql.DB) http.HandlerFunc {
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

		var update Service
		if err := decodeJSON(r, &update); err != nil {
			writeError(w, err)
			return
		}

		service, err := UpdateServiceListing(r.Context(), db, id, requesterID, update)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, service)
	}
}

func HandleDeleteService(db *sql.DB) http.HandlerFunc {
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

		if err := DeleteServiceListing(r.Context(), db, id, requesterID); err != nil {
			writeError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
