package main

import (
	"context"
	"database/sql"
	"net/http"
)

type createExchangeInput struct {
	ServiceID int `json:"service_id"`
}

func HandleCreateExchange(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requesterID, ok := userIDFromContext(r.Context())
		if !ok {
			writeError(w, ErrUnauthorized)
			return
		}

		var input createExchangeInput
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, err)
			return
		}

		exchange, err := CreateExchange(r.Context(), db, requesterID, input.ServiceID)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusCreated, exchange)
	}
}

func HandleListExchanges(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := userIDFromContext(r.Context())
		if !ok {
			writeError(w, ErrUnauthorized)
			return
		}

		status := r.URL.Query().Get("status")

		exchanges, err := ListExchanges(r.Context(), db, userID, status)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, exchanges)
	}
}

func HandleGetExchange(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseIDParam(r, "id")
		if err != nil {
			writeError(w, err)
			return
		}

		actingUserID, ok := userIDFromContext(r.Context())
		if !ok {
			writeError(w, ErrUnauthorized)
			return
		}

		exchange, err := GetExchangeForParticipant(r.Context(), db, id, actingUserID)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, exchange)
	}
}

func exchangeActionHandler(
	db *sql.DB,
	action func(ctx context.Context, db *sql.DB, id, actingUserID int) (*Exchange, error),
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseIDParam(r, "id")
		if err != nil {
			writeError(w, err)
			return
		}

		actingUserID, ok := userIDFromContext(r.Context())
		if !ok {
			writeError(w, ErrUnauthorized)
			return
		}

		exchange, err := action(r.Context(), db, id, actingUserID)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, http.StatusOK, exchange)
	}
}

func HandleAcceptExchange(db *sql.DB) http.HandlerFunc {
	return exchangeActionHandler(db, AcceptExchange)
}

func HandleRejectExchange(db *sql.DB) http.HandlerFunc {
	return exchangeActionHandler(db, RejectExchange)
}

func HandleCompleteExchange(db *sql.DB) http.HandlerFunc {
	return exchangeActionHandler(db, CompleteExchange)
}

func HandleCancelExchange(db *sql.DB) http.HandlerFunc {
	return exchangeActionHandler(db, CancelExchange)
}
