package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

var (
	testDB      *sql.DB
	dbAvailable bool
)

func TestMain(m *testing.M) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		envOr("DB_HOST", "localhost"),
		envOr("DB_PORT", "5432"),
		envOr("DB_USER", "postgres"),
		envOr("DB_PASSWORD", ""),
		envOr("DB_NAME", "barterswap"),
	)

	db, err := sql.Open("postgres", connStr)
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = db.PingContext(ctx)
		cancel()
	}

	if err != nil {
		fmt.Println("⚠️  base de données indisponible, les tests d'intégration seront ignorés:", err)
	} else {
		testDB = db
		dbAvailable = true
	}

	code := m.Run()

	if testDB != nil {
		testDB.Close()
	}
	os.Exit(code)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func uniqueSuffix() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

func createTestUser(t *testing.T, router http.Handler, pseudoPrefix string) User {
	t.Helper()

	body, _ := json.Marshal(map[string]string{"pseudo": pseudoPrefix + "_" + uniqueSuffix()})
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("création de l'utilisateur de test échouée: %d %s", rec.Code, rec.Body.String())
	}

	var user User
	if err := json.Unmarshal(rec.Body.Bytes(), &user); err != nil {
		t.Fatalf("réponse de création invalide: %v", err)
	}

	t.Cleanup(func() {
		testDB.Exec(`DELETE FROM users WHERE id = $1`, user.ID)
	})

	return user
}

func doRequest(router http.Handler, method, path, body string, userID int) *httptest.ResponseRecorder {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if userID != 0 {
		req.Header.Set("X-User-ID", strconv.Itoa(userID))
	}

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestAPI_CreateUser(t *testing.T) {
	if !dbAvailable {
		t.Skip("base de données indisponible")
	}
	router := newRouter(testDB)

	t.Run("succès avec 10 crédits de bienvenue", func(t *testing.T) {
		user := createTestUser(t, router, "alice")
		if user.CreditBalance != 10 {
			t.Errorf("credit_balance = %d, attendu 10", user.CreditBalance)
		}
	})

	t.Run("pseudo vide -> 400", func(t *testing.T) {
		rec := doRequest(router, http.MethodPost, "/api/users", `{"pseudo":""}`, 0)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("code = %d, attendu 400", rec.Code)
		}
	})
}

func TestAPI_GetUser(t *testing.T) {
	if !dbAvailable {
		t.Skip("base de données indisponible")
	}
	router := newRouter(testDB)
	user := createTestUser(t, router, "bob")

	t.Run("profil existant", func(t *testing.T) {
		rec := doRequest(router, http.MethodGet, fmt.Sprintf("/api/users/%d", user.ID), "", 0)
		if rec.Code != http.StatusOK {
			t.Fatalf("code = %d, attendu 200 (%s)", rec.Code, rec.Body.String())
		}
	})

	t.Run("profil inexistant -> 404", func(t *testing.T) {
		rec := doRequest(router, http.MethodGet, "/api/users/999999999", "", 0)
		if rec.Code != http.StatusNotFound {
			t.Errorf("code = %d, attendu 404", rec.Code)
		}
	})
}

func TestAPI_UpdateUser(t *testing.T) {
	if !dbAvailable {
		t.Skip("base de données indisponible")
	}
	router := newRouter(testDB)
	user := createTestUser(t, router, "carol")
	other := createTestUser(t, router, "dave")

	cases := []struct {
		name       string
		asUserID   int
		targetID   int
		body       string
		wantStatus int
	}{
		{"sans header X-User-ID -> 401", 0, user.ID, `{"pseudo":"x"}`, http.StatusUnauthorized},
		{"mauvais propriétaire -> 403", other.ID, user.ID, `{"pseudo":"x"}`, http.StatusForbidden},
		{"pseudo vide -> 400", user.ID, user.ID, `{"pseudo":""}`, http.StatusBadRequest},
		{"succès -> 200", user.ID, user.ID, `{"pseudo":"carol2","ville":"Nantes"}`, http.StatusOK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := doRequest(router, http.MethodPut, fmt.Sprintf("/api/users/%d", tc.targetID), tc.body, tc.asUserID)
			if rec.Code != tc.wantStatus {
				t.Errorf("code = %d, attendu %d (%s)", rec.Code, tc.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestAPI_Skills(t *testing.T) {
	if !dbAvailable {
		t.Skip("base de données indisponible")
	}
	router := newRouter(testDB)
	user := createTestUser(t, router, "erin")

	t.Run("PUT sans header -> 401", func(t *testing.T) {
		rec := doRequest(router, http.MethodPut, fmt.Sprintf("/api/users/%d/skills", user.ID),
			`[{"nom":"Jardinage","niveau":"expert"}]`, 0)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("code = %d, attendu 401", rec.Code)
		}
	})

	t.Run("PUT niveau invalide -> 400", func(t *testing.T) {
		rec := doRequest(router, http.MethodPut, fmt.Sprintf("/api/users/%d/skills", user.ID),
			`[{"nom":"Jardinage","niveau":"pro"}]`, user.ID)
		if rec.Code != http.StatusBadRequest {
			t.Errorf("code = %d, attendu 400", rec.Code)
		}
	})

	t.Run("PUT puis GET round-trip", func(t *testing.T) {
		putRec := doRequest(router, http.MethodPut, fmt.Sprintf("/api/users/%d/skills", user.ID),
			`[{"nom":"Jardinage","niveau":"expert"},{"nom":"Cuisine","niveau":"débutant"}]`, user.ID)
		if putRec.Code != http.StatusOK {
			t.Fatalf("PUT code = %d, attendu 200 (%s)", putRec.Code, putRec.Body.String())
		}

		getRec := doRequest(router, http.MethodGet, fmt.Sprintf("/api/users/%d/skills", user.ID), "", 0)
		if getRec.Code != http.StatusOK {
			t.Fatalf("GET code = %d, attendu 200", getRec.Code)
		}

		var skills []Skill
		if err := json.Unmarshal(getRec.Body.Bytes(), &skills); err != nil {
			t.Fatalf("réponse GET skills invalide: %v", err)
		}
		if len(skills) != 2 {
			t.Errorf("nb compétences = %d, attendu 2", len(skills))
		}
	})
}

func TestAPI_UserStats(t *testing.T) {
	if !dbAvailable {
		t.Skip("base de données indisponible")
	}
	router := newRouter(testDB)
	user := createTestUser(t, router, "frank")

	t.Run("utilisateur neuf: stats à zéro", func(t *testing.T) {
		rec := doRequest(router, http.MethodGet, fmt.Sprintf("/api/users/%d/stats", user.ID), "", 0)
		if rec.Code != http.StatusOK {
			t.Fatalf("code = %d, attendu 200 (%s)", rec.Code, rec.Body.String())
		}

		var stats UserStats
		if err := json.Unmarshal(rec.Body.Bytes(), &stats); err != nil {
			t.Fatalf("réponse stats invalide: %v", err)
		}
		if stats.CreditBalance != 10 || stats.ServicesActifs != 0 || stats.EchangesCompletes != 0 || stats.NbAvis != 0 {
			t.Errorf("stats inattendues pour un utilisateur neuf: %+v", stats)
		}
	})

	t.Run("utilisateur inexistant -> 404", func(t *testing.T) {
		rec := doRequest(router, http.MethodGet, "/api/users/999999999/stats", "", 0)
		if rec.Code != http.StatusNotFound {
			t.Errorf("code = %d, attendu 404", rec.Code)
		}
	})
}
