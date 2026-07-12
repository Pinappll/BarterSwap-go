package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestAPI_CreateReview(t *testing.T) {
	if !dbAvailable {
		t.Skip("base de données indisponible")
	}
	router := newRouter(testDB)
	sc := setupExchangeScenario(t, router, "Photographie", 3)
	exchange := createTestExchange(t, router, sc.Requester.ID, sc.Service.ID)

	t.Run("avis sur échange non terminé -> 400", func(t *testing.T) {
		rec := doRequest(router, http.MethodPost, fmt.Sprintf("/api/exchanges/%d/review", exchange.ID), `{"note":5}`, sc.Requester.ID)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("code = %d, attendu 400", rec.Code)
		}
	})

	doRequest(router, http.MethodPut, fmt.Sprintf("/api/exchanges/%d/accept", exchange.ID), "", sc.Provider.ID)
	doRequest(router, http.MethodPut, fmt.Sprintf("/api/exchanges/%d/complete", exchange.ID), "", sc.Requester.ID)

	t.Run("note hors 1-5 -> 400", func(t *testing.T) {
		rec := doRequest(router, http.MethodPost, fmt.Sprintf("/api/exchanges/%d/review", exchange.ID), `{"note":0}`, sc.Requester.ID)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("code = %d, attendu 400", rec.Code)
		}
	})

	t.Run("tiers non participant -> 403", func(t *testing.T) {
		stranger := createTestUser(t, router, "stranger")
		rec := doRequest(router, http.MethodPost, fmt.Sprintf("/api/exchanges/%d/review", exchange.ID), `{"note":5}`, stranger.ID)
		if rec.Code != http.StatusForbidden {
			t.Errorf("code = %d, attendu 403", rec.Code)
		}
	})

	t.Run("succès puis doublon refusé -> 400", func(t *testing.T) {
		rec := doRequest(router, http.MethodPost, fmt.Sprintf("/api/exchanges/%d/review", exchange.ID), `{"note":5,"commentaire":"Top"}`, sc.Requester.ID)
		if rec.Code != http.StatusCreated {
			t.Fatalf("code = %d, attendu 201 (%s)", rec.Code, rec.Body.String())
		}

		var review Review
		json.Unmarshal(rec.Body.Bytes(), &review)
		if review.TargetID != sc.Provider.ID {
			t.Errorf("target_id = %d, attendu %d (l'offreur)", review.TargetID, sc.Provider.ID)
		}

		dupRec := doRequest(router, http.MethodPost, fmt.Sprintf("/api/exchanges/%d/review", exchange.ID), `{"note":3}`, sc.Requester.ID)
		if dupRec.Code != http.StatusBadRequest {
			t.Errorf("code = %d, attendu 400 (avis en double)", dupRec.Code)
		}
	})

	t.Run("l'autre partie peut aussi noter", func(t *testing.T) {
		rec := doRequest(router, http.MethodPost, fmt.Sprintf("/api/exchanges/%d/review", exchange.ID), `{"note":4}`, sc.Provider.ID)
		if rec.Code != http.StatusCreated {
			t.Errorf("code = %d, attendu 201 (%s)", rec.Code, rec.Body.String())
		}
	})
}

func TestAPI_ListReviews(t *testing.T) {
	if !dbAvailable {
		t.Skip("base de données indisponible")
	}
	router := newRouter(testDB)
	sc := setupExchangeScenario(t, router, "Déménagement", 4)
	exchange := createTestExchange(t, router, sc.Requester.ID, sc.Service.ID)
	doRequest(router, http.MethodPut, fmt.Sprintf("/api/exchanges/%d/accept", exchange.ID), "", sc.Provider.ID)
	doRequest(router, http.MethodPut, fmt.Sprintf("/api/exchanges/%d/complete", exchange.ID), "", sc.Requester.ID)
	doRequest(router, http.MethodPost, fmt.Sprintf("/api/exchanges/%d/review", exchange.ID), `{"note":5}`, sc.Requester.ID)
	doRequest(router, http.MethodPost, fmt.Sprintf("/api/exchanges/%d/review", exchange.ID), `{"note":2}`, sc.Provider.ID)

	t.Run("avis reçus par l'offreur ne contiennent que ceux le ciblant", func(t *testing.T) {
		rec := doRequest(router, http.MethodGet, fmt.Sprintf("/api/users/%d/reviews", sc.Provider.ID), "", 0)
		var reviews []Review
		json.Unmarshal(rec.Body.Bytes(), &reviews)
		for _, r := range reviews {
			if r.ExchangeID == exchange.ID && r.TargetID != sc.Provider.ID {
				t.Errorf("un avis ciblant %d apparaît dans les avis reçus par %d", r.TargetID, sc.Provider.ID)
			}
		}
	})

	t.Run("avis sur le service ne contiennent que ceux visant le prestataire", func(t *testing.T) {
		rec := doRequest(router, http.MethodGet, fmt.Sprintf("/api/services/%d/reviews", sc.Service.ID), "", 0)
		var reviews []Review
		json.Unmarshal(rec.Body.Bytes(), &reviews)

		if len(reviews) != 1 {
			t.Fatalf("nb avis sur le service = %d, attendu 1", len(reviews))
		}
		if reviews[0].TargetID != sc.Provider.ID {
			t.Errorf("target_id = %d, attendu %d (l'avis du client sur l'offreur, pas l'inverse)", reviews[0].TargetID, sc.Provider.ID)
		}
	})
}
