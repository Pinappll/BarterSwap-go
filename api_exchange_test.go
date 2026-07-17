package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

type exchangeScenario struct {
	Provider  User
	Requester User
	Service   Service
}

func setupExchangeScenario(t *testing.T, router http.Handler, categorie string, credits int) exchangeScenario {
	t.Helper()

	provider := createTestUserWithSkill(t, router, "provider", categorie, "expert")
	requester := createTestUser(t, router, "requester")
	service := createTestService(t, router, provider.ID,
		fmt.Sprintf(`{"titre":"Service test","categorie":%q,"duree_minutes":60,"credits":%d}`, categorie, credits))

	return exchangeScenario{Provider: provider, Requester: requester, Service: service}
}

func createTestExchange(t *testing.T, router http.Handler, requesterID, serviceID int) Exchange {
	t.Helper()

	body := fmt.Sprintf(`{"service_id":%d}`, serviceID)
	rec := doRequest(router, http.MethodPost, "/api/exchanges", body, requesterID)
	if rec.Code != http.StatusCreated {
		t.Fatalf("création de l'échange de test échouée: %d %s", rec.Code, rec.Body.String())
	}

	var exchange Exchange
	if err := json.Unmarshal(rec.Body.Bytes(), &exchange); err != nil {
		t.Fatalf("réponse de création d'échange invalide: %v", err)
	}

	t.Cleanup(func() {
		testDB.Exec(`DELETE FROM exchanges WHERE id = $1`, exchange.ID)
	})

	return exchange
}

func getBalance(t *testing.T, router http.Handler, userID int) int {
	t.Helper()

	rec := doRequest(router, http.MethodGet, fmt.Sprintf("/api/users/%d", userID), "", 0)
	var u User
	if err := json.Unmarshal(rec.Body.Bytes(), &u); err != nil {
		t.Fatalf("réponse profil invalide: %v", err)
	}
	return u.CreditBalance
}

func TestAPI_CreateExchange(t *testing.T) {
	if !dbAvailable {
		t.Skip("base de données indisponible")
	}
	router := newRouter(testDB)
	sc := setupExchangeScenario(t, router, "Jardinage", 3)

	t.Run("auto-échange -> 400", func(t *testing.T) {
		rec := doRequest(router, http.MethodPost, "/api/exchanges", fmt.Sprintf(`{"service_id":%d}`, sc.Service.ID), sc.Provider.ID)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("code = %d, attendu 400", rec.Code)
		}
	})

	t.Run("sans header -> 401", func(t *testing.T) {
		rec := doRequest(router, http.MethodPost, "/api/exchanges", fmt.Sprintf(`{"service_id":%d}`, sc.Service.ID), 0)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("code = %d, attendu 401", rec.Code)
		}
	})

	t.Run("crédits insuffisants -> 400", func(t *testing.T) {
		poorUser := createTestUser(t, router, "poor")
		expensiveService := createTestService(t, router, sc.Provider.ID,
			`{"titre":"Service cher","categorie":"Jardinage","duree_minutes":60,"credits":10}`)
		ex := createTestExchange(t, router, poorUser.ID, expensiveService.ID)

		acceptRec := doRequest(router, http.MethodPut, fmt.Sprintf("/api/exchanges/%d/accept", ex.ID), "", sc.Provider.ID)
		if acceptRec.Code != http.StatusOK {
			t.Fatalf("préparation (acceptation) échouée: %d %s", acceptRec.Code, acceptRec.Body.String())
		}

		rec := doRequest(router, http.MethodPost, "/api/exchanges", fmt.Sprintf(`{"service_id":%d}`, sc.Service.ID), poorUser.ID)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("code = %d, attendu 400", rec.Code)
		}
	})

	t.Run("succès puis conflit sur le même service -> 409", func(t *testing.T) {
		exchange := createTestExchange(t, router, sc.Requester.ID, sc.Service.ID)
		if exchange.Status != "pending" {
			t.Errorf("status = %q, attendu pending", exchange.Status)
		}

		other := createTestUser(t, router, "other")
		rec := doRequest(router, http.MethodPost, "/api/exchanges", fmt.Sprintf(`{"service_id":%d}`, sc.Service.ID), other.ID)
		if rec.Code != http.StatusConflict {
			t.Errorf("code = %d, attendu 409 (%s)", rec.Code, rec.Body.String())
		}
	})
}

func TestAPI_AcceptExchange(t *testing.T) {
	if !dbAvailable {
		t.Skip("base de données indisponible")
	}
	router := newRouter(testDB)
	sc := setupExchangeScenario(t, router, "Bricolage", 4)
	exchange := createTestExchange(t, router, sc.Requester.ID, sc.Service.ID)

	t.Run("mauvais utilisateur (requester) -> 403", func(t *testing.T) {
		rec := doRequest(router, http.MethodPut, fmt.Sprintf("/api/exchanges/%d/accept", exchange.ID), "", sc.Requester.ID)
		if rec.Code != http.StatusForbidden {
			t.Errorf("code = %d, attendu 403", rec.Code)
		}
	})

	t.Run("succès -> crédits bloqués côté demandeur, pas encore côté offreur", func(t *testing.T) {
		balanceBefore := getBalance(t, router, sc.Requester.ID)

		rec := doRequest(router, http.MethodPut, fmt.Sprintf("/api/exchanges/%d/accept", exchange.ID), "", sc.Provider.ID)
		if rec.Code != http.StatusOK {
			t.Fatalf("code = %d, attendu 200 (%s)", rec.Code, rec.Body.String())
		}

		if got := getBalance(t, router, sc.Requester.ID); got != balanceBefore-sc.Service.Credits {
			t.Errorf("solde demandeur = %d, attendu %d", got, balanceBefore-sc.Service.Credits)
		}
		if got := getBalance(t, router, sc.Provider.ID); got != 10 {
			t.Errorf("solde offreur = %d, attendu 10 (pas encore crédité)", got)
		}
	})

	t.Run("accepter un échange déjà accepté -> 409", func(t *testing.T) {
		rec := doRequest(router, http.MethodPut, fmt.Sprintf("/api/exchanges/%d/accept", exchange.ID), "", sc.Provider.ID)
		if rec.Code != http.StatusConflict {
			t.Errorf("code = %d, attendu 409", rec.Code)
		}
	})
}

func TestAPI_CompleteExchange(t *testing.T) {
	if !dbAvailable {
		t.Skip("base de données indisponible")
	}
	router := newRouter(testDB)
	sc := setupExchangeScenario(t, router, "Musique", 5)
	exchange := createTestExchange(t, router, sc.Requester.ID, sc.Service.ID)
	doRequest(router, http.MethodPut, fmt.Sprintf("/api/exchanges/%d/accept", exchange.ID), "", sc.Provider.ID)

	t.Run("l'offreur ne peut pas s'auto-valider -> 403", func(t *testing.T) {
		rec := doRequest(router, http.MethodPut, fmt.Sprintf("/api/exchanges/%d/complete", exchange.ID), "", sc.Provider.ID)
		if rec.Code != http.StatusForbidden {
			t.Errorf("code = %d, attendu 403", rec.Code)
		}
	})

	t.Run("le demandeur complète -> crédits transférés définitivement", func(t *testing.T) {
		rec := doRequest(router, http.MethodPut, fmt.Sprintf("/api/exchanges/%d/complete", exchange.ID), "", sc.Requester.ID)
		if rec.Code != http.StatusOK {
			t.Fatalf("code = %d, attendu 200 (%s)", rec.Code, rec.Body.String())
		}

		if got := getBalance(t, router, sc.Provider.ID); got != 10+sc.Service.Credits {
			t.Errorf("solde offreur = %d, attendu %d", got, 10+sc.Service.Credits)
		}
	})
}

func TestAPI_RejectExchange(t *testing.T) {
	if !dbAvailable {
		t.Skip("base de données indisponible")
	}
	router := newRouter(testDB)
	sc := setupExchangeScenario(t, router, "Sport", 2)
	exchange := createTestExchange(t, router, sc.Requester.ID, sc.Service.ID)
	balanceBefore := getBalance(t, router, sc.Requester.ID)

	rec := doRequest(router, http.MethodPut, fmt.Sprintf("/api/exchanges/%d/reject", exchange.ID), "", sc.Provider.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("code = %d, attendu 200 (%s)", rec.Code, rec.Body.String())
	}

	if got := getBalance(t, router, sc.Requester.ID); got != balanceBefore {
		t.Errorf("solde demandeur = %d, ne devrait pas avoir changé (rien n'était bloqué depuis pending)", got)
	}
}

func TestAPI_CancelExchange(t *testing.T) {
	if !dbAvailable {
		t.Skip("base de données indisponible")
	}
	router := newRouter(testDB)

	t.Run("annulation depuis pending: rien à rembourser", func(t *testing.T) {
		sc := setupExchangeScenario(t, router, "Couture", 3)
		exchange := createTestExchange(t, router, sc.Requester.ID, sc.Service.ID)
		balanceBefore := getBalance(t, router, sc.Requester.ID)

		rec := doRequest(router, http.MethodPut, fmt.Sprintf("/api/exchanges/%d/cancel", exchange.ID), "", sc.Requester.ID)
		if rec.Code != http.StatusOK {
			t.Fatalf("code = %d, attendu 200 (%s)", rec.Code, rec.Body.String())
		}
		if got := getBalance(t, router, sc.Requester.ID); got != balanceBefore {
			t.Errorf("solde = %d, ne devrait pas avoir changé", got)
		}
	})

	t.Run("annulation depuis accepted: remboursement", func(t *testing.T) {
		sc := setupExchangeScenario(t, router, "Animalier", 6)
		exchange := createTestExchange(t, router, sc.Requester.ID, sc.Service.ID)
		doRequest(router, http.MethodPut, fmt.Sprintf("/api/exchanges/%d/accept", exchange.ID), "", sc.Provider.ID)

		rec := doRequest(router, http.MethodPut, fmt.Sprintf("/api/exchanges/%d/cancel", exchange.ID), "", sc.Provider.ID)
		if rec.Code != http.StatusOK {
			t.Fatalf("code = %d, attendu 200 (%s)", rec.Code, rec.Body.String())
		}
		if got := getBalance(t, router, sc.Requester.ID); got != 10 {
			t.Errorf("solde demandeur = %d, attendu 10 après remboursement", got)
		}
	})
}

func TestAPI_ListAndGetExchange(t *testing.T) {
	if !dbAvailable {
		t.Skip("base de données indisponible")
	}
	router := newRouter(testDB)
	sc := setupExchangeScenario(t, router, "Langues", 2)
	exchange := createTestExchange(t, router, sc.Requester.ID, sc.Service.ID)

	t.Run("liste des échanges du demandeur", func(t *testing.T) {
		rec := doRequest(router, http.MethodGet, "/api/exchanges", "", sc.Requester.ID)
		if rec.Code != http.StatusOK {
			t.Fatalf("code = %d, attendu 200", rec.Code)
		}

		var exchanges []Exchange
		json.Unmarshal(rec.Body.Bytes(), &exchanges)
		found := false
		for _, e := range exchanges {
			if e.ID == exchange.ID {
				found = true
			}
		}
		if !found {
			t.Errorf("l'échange créé n'apparaît pas dans la liste du demandeur")
		}
	})

	t.Run("filtre par statut exclut les autres statuts", func(t *testing.T) {
		rec := doRequest(router, http.MethodGet, "/api/exchanges?status=accepted", "", sc.Requester.ID)
		var exchanges []Exchange
		json.Unmarshal(rec.Body.Bytes(), &exchanges)
		for _, e := range exchanges {
			if e.ID == exchange.ID {
				t.Errorf("l'échange pending ne devrait pas apparaître dans le filtre status=accepted")
			}
		}
	})

	t.Run("détail accessible aux participants", func(t *testing.T) {
		rec := doRequest(router, http.MethodGet, fmt.Sprintf("/api/exchanges/%d", exchange.ID), "", sc.Provider.ID)
		if rec.Code != http.StatusOK {
			t.Errorf("code = %d, attendu 200", rec.Code)
		}
	})

	t.Run("détail refusé à un tiers -> 403", func(t *testing.T) {
		stranger := createTestUser(t, router, "stranger")
		rec := doRequest(router, http.MethodGet, fmt.Sprintf("/api/exchanges/%d", exchange.ID), "", stranger.ID)
		if rec.Code != http.StatusForbidden {
			t.Errorf("code = %d, attendu 403", rec.Code)
		}
	})
}
