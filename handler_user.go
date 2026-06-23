package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
)

func HandleCreateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user User

		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, `{"error": "JSON invalide"}`, http.StatusBadRequest)
			return
		}

		err := CreateUserService(db, &user)
		if err != nil {
			if errors.Is(err, ErrInvalidInput) {
				http.Error(w, `{"error": "Le pseudo est obligatoire"}`, http.StatusBadRequest)
				return
			}
			http.Error(w, `{"error": "Erreur interne du serveur"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	}
}
