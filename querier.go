package main

import (
	"context"
	"database/sql"
)

// querier regroupe les méthodes communes à *sql.DB et *sql.Tx, pour que les
// fonctions du repository puissent tourner seules ou dans une transaction plus
// large (ex: lire un service pendant l'acceptation d'un échange).
type querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}
