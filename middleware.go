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

// requestTimeout borne la durée de toute requête (et des appels DB qui en
// découlent) via le context.Context propagé à travers les couches.
const requestTimeout = 5 * time.Second

// withRecovery récupère les panics pour éviter qu'un crash dans un handler
// ne fasse tomber tout le serveur.
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

// withLogging journalise chaque requête traitée avec sa durée.
func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// withCORS autorise les appels cross-origin (utile pour tester depuis un
// client web ou un outil comme Postman/Insomnia en mode navigateur).
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

// withTimeout borne chaque requête dans le temps via context.WithTimeout ;
// les appels database/sql en aval doivent utiliser les variantes *Context
// pour hériter de cette annulation.
func withTimeout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// withUserID lit le header X-User-ID (l'authentification "simple" imposée
// par le sujet) et, s'il est présent et valide, place l'identifiant dans le
// context de la requête. L'absence du header n'est pas bloquée ici : ce
// sont les endpoints qui en ont besoin (modification, réservation...) qui
// exigent sa présence via userIDFromContext.
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

// userIDFromContext récupère l'identifiant posé par withUserID.
func userIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(userIDContextKey).(int)
	return id, ok
}
