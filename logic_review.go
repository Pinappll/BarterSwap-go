package main

import (
	"context"
	"database/sql"
)

func validateNote(note int) error {
	if note < 1 || note > 5 {
		return ErrInvalidInput
	}
	return nil
}

func CreateReview(ctx context.Context, db *sql.DB, exchangeID, authorID, note int, commentaire string) (*Review, error) {
	if err := validateNote(note); err != nil {
		return nil, err
	}

	exchange, err := SelectExchangeByID(ctx, db, exchangeID)
	if err != nil {
		return nil, err
	}
	if exchange.Status != "completed" {
		return nil, ErrReviewNotAllowed
	}

	var targetID int
	switch authorID {
	case exchange.RequesterID:
		targetID = exchange.OwnerID
	case exchange.OwnerID:
		targetID = exchange.RequesterID
	default:
		return nil, ErrForbidden
	}

	alreadyReviewed, err := ExistsReviewByAuthorForExchange(ctx, db, exchangeID, authorID)
	if err != nil {
		return nil, err
	}
	if alreadyReviewed {
		return nil, ErrReviewNotAllowed
	}

	review := &Review{
		ExchangeID:  exchangeID,
		AuthorID:    authorID,
		TargetID:    targetID,
		Note:        note,
		Commentaire: commentaire,
	}
	if err := InsertReview(ctx, db, review); err != nil {
		return nil, err
	}

	return review, nil
}

func GetUserReviews(ctx context.Context, db *sql.DB, userID int) ([]Review, error) {
	if _, err := SelectUserByID(ctx, db, userID); err != nil {
		return nil, err
	}
	return SelectReviewsByTargetID(ctx, db, userID)
}

func GetServiceReviews(ctx context.Context, db *sql.DB, serviceID int) ([]Review, error) {
	if _, err := SelectServiceByID(ctx, db, serviceID); err != nil {
		return nil, err
	}
	return SelectReviewsByServiceID(ctx, db, serviceID)
}
