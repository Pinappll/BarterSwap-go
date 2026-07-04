package main

import (
	"context"
	"database/sql"
	"math"
)

func GetUserStats(ctx context.Context, db *sql.DB, userID int) (*UserStats, error) {
	stats, err := SelectUserStats(ctx, db, userID)
	if err != nil {
		return nil, err
	}

	stats.NoteMoyenne = math.Round(stats.NoteMoyenne*100) / 100

	return stats, nil
}
