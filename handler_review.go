package main

import (
	"database/sql"
	"net/http"
)

type createReviewInput struct {
	Note        int    `json:"note"`
	Commentaire string `json:"commentaire"`
}

func HandleCreateReview(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exchangeID, err := parseIDParam(r, "id")
		if err != nil {
			writeError(w, err)
			return
		}

		authorID, ok := userIDFromContext(r.Context())
		if !ok {
			writeError(w, ErrUnauthorized)
			return
		}

		var input createReviewInput
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, err)
			return
		}

		review, err := CreateReview(r.Context(), db, exchangeID, authorID, input.Note, input.Commentaire)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusCreated, review)
	}
}

func HandleGetUserReviews(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseIDParam(r, "id")
		if err != nil {
			writeError(w, err)
			return
		}

		reviews, err := GetUserReviews(r.Context(), db, id)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, reviews)
	}
}

func HandleGetServiceReviews(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseIDParam(r, "id")
		if err != nil {
			writeError(w, err)
			return
		}

		reviews, err := GetServiceReviews(r.Context(), db, id)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, reviews)
	}
}
