package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"
)

type contextKey string

const userIDContextKey contextKey = "userID"

const requestTimeout = 5 * time.Second

func withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic récupéré sur %s %s: %v", r.Method, r.URL.Path, rec)
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "erreur interne du serveur"})
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-ID")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withTimeout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func withUserID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := r.Header.Get("X-User-ID")
		if raw == "" {
			next.ServeHTTP(w, r)
			return
		}

		id, err := strconv.Atoi(raw)
		if err != nil || id <= 0 {
			writeError(w, ErrUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func userIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(userIDContextKey).(int)
	return id, ok
}
