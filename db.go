package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func InitDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	fmt.Println("✅ Connexion à PostgreSQL réussie !")
	return db, nil
}
