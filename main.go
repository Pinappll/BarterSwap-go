package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
)


func newRouter(db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /api/users", HandleCreateUser(db))
	mux.HandleFunc("GET /api/users/{id}", HandleGetUser(db))
	mux.HandleFunc("PUT /api/users/{id}", HandleUpdateUser(db))
	mux.HandleFunc("GET /api/users/{id}/skills", HandleGetSkills(db))
	mux.HandleFunc("PUT /api/users/{id}/skills", HandlePutSkills(db))
	mux.HandleFunc("GET /api/users/{id}/stats", HandleGetUserStats(db))

	mux.HandleFunc("GET /api/services", HandleListServices(db))
	mux.HandleFunc("POST /api/services", HandleCreateService(db))
	mux.HandleFunc("GET /api/services/{id}", HandleGetService(db))
	mux.HandleFunc("PUT /api/services/{id}", HandleUpdateService(db))
	mux.HandleFunc("DELETE /api/services/{id}", HandleDeleteService(db))

	mux.HandleFunc("POST /api/exchanges", HandleCreateExchange(db))
	mux.HandleFunc("GET /api/exchanges", HandleListExchanges(db))
	mux.HandleFunc("GET /api/exchanges/{id}", HandleGetExchange(db))
	mux.HandleFunc("PUT /api/exchanges/{id}/accept", HandleAcceptExchange(db))
	mux.HandleFunc("PUT /api/exchanges/{id}/reject", HandleRejectExchange(db))
	mux.HandleFunc("PUT /api/exchanges/{id}/complete", HandleCompleteExchange(db))
	mux.HandleFunc("PUT /api/exchanges/{id}/cancel", HandleCancelExchange(db))
	mux.HandleFunc("POST /api/exchanges/{id}/review", HandleCreateReview(db))

	mux.HandleFunc("GET /api/users/{id}/reviews", HandleGetUserReviews(db))
	mux.HandleFunc("GET /api/services/{id}/reviews", HandleGetServiceReviews(db))

	var handler http.Handler = mux
	handler = withUserID(handler)
	handler = withTimeout(handler)
	handler = withCORS(handler)
	handler = withLogging(handler)
	handler = withRecovery(handler)

	return handler
}

func main() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := InitDB(connStr)
	if err != nil {
		log.Fatalf("Erreur de connexion à la base de données : %v", err)
	}
	defer db.Close()

	server := &http.Server{
		Addr:    ":8080",
		Handler: newRouter(db),
	}

	fmt.Println("Serveur démarré sur http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Erreur du serveur HTTP : %v", err)
	}
}
