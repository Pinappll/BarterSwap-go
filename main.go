package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

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
		log.Fatalf("❌ Erreur de connexion à la base de données : %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	// --- NOUVELLE ROUTE ICI ---
	mux.HandleFunc("POST /api/users", HandleCreateUser(db))
	// --------------------------

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	fmt.Println("🚀 Serveur démarré sur http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ Erreur du serveur HTTP : %v", err)
	}
}
