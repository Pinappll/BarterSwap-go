package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// createTestUserWithSkill crée un utilisateur de test et lui attribue
// immédiatement une compétence, pour satisfaire la règle métier "il faut
// posséder la compétence liée à la catégorie du service".
func createTestUserWithSkill(t *testing.T, router http.Handler, pseudoPrefix, skillNom, niveau string) User {
	t.Helper()

	user := createTestUser(t, router, pseudoPrefix)

	body := fmt.Sprintf(`[{"nom":%q,"niveau":%q}]`, skillNom, niveau)
	rec := doRequest(router, http.MethodPut, fmt.Sprintf("/api/users/%d/skills", user.ID), body, user.ID)
	if rec.Code != http.StatusOK {
		t.Fatalf("attribution de compétence échouée: %d %s", rec.Code, rec.Body.String())
	}

	return user
}

func createTestService(t *testing.T, router http.Handler, providerID int, body string) Service {
	t.Helper()

	rec := doRequest(router, http.MethodPost, "/api/services", body, providerID)
	if rec.Code != http.StatusCreated {
		t.Fatalf("création du service de test échouée: %d %s", rec.Code, rec.Body.String())
	}

	var service Service
	if err := json.Unmarshal(rec.Body.Bytes(), &service); err != nil {
		t.Fatalf("réponse de création de service invalide: %v", err)
	}

	t.Cleanup(func() {
		testDB.Exec(`DELETE FROM services WHERE id = $1`, service.ID)
	})

	return service
}

func TestAPI_CreateService(t *testing.T) {
	if !dbAvailable {
		t.Skip("base de données indisponible")
	}
	router := newRouter(testDB)
	user := createTestUserWithSkill(t, router, "gardener", "Jardinage", "expert")

	t.Run("sans header -> 401", func(t *testing.T) {
		rec := doRequest(router, http.MethodPost, "/api/services",
			`{"titre":"Taille de haie","categorie":"Jardinage","duree_minutes":60,"credits":2}`, 0)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("code = %d, attendu 401", rec.Code)
		}
	})

	t.Run("compétence manquante -> 400", func(t *testing.T) {
		rec := doRequest(router, http.MethodPost, "/api/services",
			`{"titre":"Cours de guitare","categorie":"Musique","duree_minutes":60,"credits":2}`, user.ID)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("code = %d, attendu 400", rec.Code)
		}
	})

	t.Run("catégorie hors liste fermée -> 400", func(t *testing.T) {
		rec := doRequest(router, http.MethodPost, "/api/services",
			`{"titre":"Yoga","categorie":"Yoga","duree_minutes":60,"credits":2}`, user.ID)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("code = %d, attendu 400", rec.Code)
		}
	})

	t.Run("succès -> 201, provider_id forcé par le header", func(t *testing.T) {
		service := createTestService(t, router, user.ID,
			`{"titre":"Taille de haie","description":"desc","categorie":"Jardinage","duree_minutes":90,"credits":3,"ville":"Paris","provider_id":999999}`)
		if service.ProviderID != user.ID {
			t.Errorf("provider_id = %d, attendu %d (le champ envoyé par le client doit être ignoré)", service.ProviderID, user.ID)
		}
		if !service.Actif {
			t.Errorf("actif = false, attendu true à la création")
		}
	})
}

func TestAPI_GetService(t *testing.T) {
	if !dbAvailable {
		t.Skip("base de données indisponible")
	}
	router := newRouter(testDB)
	user := createTestUserWithSkill(t, router, "cook", "Cuisine", "débutant")
	service := createTestService(t, router, user.ID,
		`{"titre":"Cours de cuisine","categorie":"Cuisine","duree_minutes":45,"credits":1}`)

	t.Run("existant", func(t *testing.T) {
		rec := doRequest(router, http.MethodGet, fmt.Sprintf("/api/services/%d", service.ID), "", 0)
		if rec.Code != http.StatusOK {
			t.Errorf("code = %d, attendu 200", rec.Code)
		}
	})

	t.Run("inexistant -> 404", func(t *testing.T) {
		rec := doRequest(router, http.MethodGet, "/api/services/999999999", "", 0)
		if rec.Code != http.StatusNotFound {
			t.Errorf("code = %d, attendu 404", rec.Code)
		}
	})
}

func TestAPI_ServiceFilters(t *testing.T) {
	if !dbAvailable {
		t.Skip("base de données indisponible")
	}
	router := newRouter(testDB)
	user := createTestUserWithSkill(t, router, "musicien", "Musique", "expert")
	createTestService(t, router, user.ID,
		`{"titre":"Cours de guitare","description":"pour débutants","categorie":"Musique","duree_minutes":60,"credits":2,"ville":"Marseille"}`)

	cases := []struct {
		name      string
		query     string
		wantCount int
	}{
		{"filtre catégorie correspondante", "?categorie=Musique", 1},
		{"filtre catégorie non correspondante", "?categorie=Cuisine", 0},
		{"filtre ville insensible à la casse", "?ville=marseille", 1},
		{"recherche texte dans le titre", "?search=guitare", 1},
		{"recherche texte sans résultat", "?search=zzz_absent", 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doRequest(router, http.MethodGet, "/api/services"+tc.query, "", 0)
			if rec.Code != http.StatusOK {
				t.Fatalf("code = %d, attendu 200", rec.Code)
			}

			var services []Service
			if err := json.Unmarshal(rec.Body.Bytes(), &services); err != nil {
				t.Fatalf("réponse invalide: %v", err)
			}

			found := 0
			for _, s := range services {
				if s.ProviderID == user.ID {
					found++
				}
			}
			if found != tc.wantCount {
				t.Errorf("services trouvés pour ce provider = %d, attendu %d", found, tc.wantCount)
			}
		})
	}
}

func TestAPI_UpdateService(t *testing.T) {
	if !dbAvailable {
		t.Skip("base de données indisponible")
	}
	router := newRouter(testDB)
	owner := createTestUserWithSkill(t, router, "brico", "Bricolage", "expert")
	other := createTestUserWithSkill(t, router, "autre", "Bricolage", "débutant")
	service := createTestService(t, router, owner.ID,
		`{"titre":"Montage meuble","categorie":"Bricolage","duree_minutes":60,"credits":2}`)

	cases := []struct {
		name       string
		asUserID   int
		body       string
		wantStatus int
	}{
		{"sans header -> 401", 0, `{"titre":"x","categorie":"Bricolage","duree_minutes":30,"credits":1}`, http.StatusUnauthorized},
		{"mauvais propriétaire -> 403", other.ID, `{"titre":"x","categorie":"Bricolage","duree_minutes":30,"credits":1}`, http.StatusForbidden},
		{"catégorie sans compétence correspondante -> 400", owner.ID, `{"titre":"x","categorie":"Musique","duree_minutes":30,"credits":1}`, http.StatusBadRequest},
		{"succès -> 200", owner.ID, `{"titre":"Montage meuble premium","categorie":"Bricolage","duree_minutes":45,"credits":3,"ville":"Lille"}`, http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doRequest(router, http.MethodPut, fmt.Sprintf("/api/services/%d", service.ID), tc.body, tc.asUserID)
			if rec.Code != tc.wantStatus {
				t.Errorf("code = %d, attendu %d (%s)", rec.Code, tc.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestAPI_DeleteService(t *testing.T) {
	if !dbAvailable {
		t.Skip("base de données indisponible")
	}
	router := newRouter(testDB)
	owner := createTestUserWithSkill(t, router, "photo", "Photographie", "expert")
	other := createTestUserWithSkill(t, router, "autre2", "Photographie", "débutant")
	service := createTestService(t, router, owner.ID,
		`{"titre":"Shooting photo","categorie":"Photographie","duree_minutes":60,"credits":2}`)

	t.Run("mauvais propriétaire -> 403", func(t *testing.T) {
		rec := doRequest(router, http.MethodDelete, fmt.Sprintf("/api/services/%d", service.ID), "", other.ID)
		if rec.Code != http.StatusForbidden {
			t.Errorf("code = %d, attendu 403", rec.Code)
		}
	})

	t.Run("succès -> 204 puis 404 sur GET", func(t *testing.T) {
		rec := doRequest(router, http.MethodDelete, fmt.Sprintf("/api/services/%d", service.ID), "", owner.ID)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("code = %d, attendu 204", rec.Code)
		}

		getRec := doRequest(router, http.MethodGet, fmt.Sprintf("/api/services/%d", service.ID), "", 0)
		if getRec.Code != http.StatusNotFound {
			t.Errorf("code = %d, attendu 404", getRec.Code)
		}
	})
}

// TestAPI_DeleteServiceWithActiveExchange couvre le bug corrigé où
// supprimer un service ayant un échange pending/accepted cascadait sur
// exchanges puis credit_transactions, effaçant la trace de crédits déjà
// bloqués chez le demandeur sans aucun moyen de les lui restituer.
func TestAPI_DeleteServiceWithActiveExchange(t *testing.T) {
	if !dbAvailable {
		t.Skip("base de données indisponible")
	}
	router := newRouter(testDB)

	t.Run("échange pending -> suppression refusée (409)", func(t *testing.T) {
		sc := setupExchangeScenario(t, router, "Informatique", 3)
		createTestExchange(t, router, sc.Requester.ID, sc.Service.ID)

		rec := doRequest(router, http.MethodDelete, fmt.Sprintf("/api/services/%d", sc.Service.ID), "", sc.Provider.ID)
		if rec.Code != http.StatusConflict {
			t.Errorf("code = %d, attendu 409 (%s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("échange accepted -> suppression refusée (409), crédits restent traçables", func(t *testing.T) {
		sc := setupExchangeScenario(t, router, "Tutorat", 4)
		exchange := createTestExchange(t, router, sc.Requester.ID, sc.Service.ID)
		doRequest(router, http.MethodPut, fmt.Sprintf("/api/exchanges/%d/accept", exchange.ID), "", sc.Provider.ID)

		rec := doRequest(router, http.MethodDelete, fmt.Sprintf("/api/services/%d", sc.Service.ID), "", sc.Provider.ID)
		if rec.Code != http.StatusConflict {
			t.Errorf("code = %d, attendu 409 (%s)", rec.Code, rec.Body.String())
		}

		getRec := doRequest(router, http.MethodGet, fmt.Sprintf("/api/services/%d", sc.Service.ID), "", 0)
		if getRec.Code != http.StatusOK {
			t.Errorf("le service devrait toujours exister après un refus de suppression, code = %d", getRec.Code)
		}
	})

	t.Run("échange completed -> suppression autorisée", func(t *testing.T) {
		sc := setupExchangeScenario(t, router, "Sport", 2)
		exchange := createTestExchange(t, router, sc.Requester.ID, sc.Service.ID)
		doRequest(router, http.MethodPut, fmt.Sprintf("/api/exchanges/%d/accept", exchange.ID), "", sc.Provider.ID)
		doRequest(router, http.MethodPut, fmt.Sprintf("/api/exchanges/%d/complete", exchange.ID), "", sc.Requester.ID)

		rec := doRequest(router, http.MethodDelete, fmt.Sprintf("/api/services/%d", sc.Service.ID), "", sc.Provider.ID)
		if rec.Code != http.StatusNoContent {
			t.Errorf("code = %d, attendu 204 (un échange terminé ne doit pas bloquer la suppression)", rec.Code)
		}
	})
}
