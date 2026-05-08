//go:build integration

package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/lib/pq"

	"social-networking-platform/users-service/internal/apiresponse"
	"social-networking-platform/users-service/internal/dbmigrate"
	handlers "social-networking-platform/users-service/internal/handler/http"
	"social-networking-platform/users-service/internal/middleware"
	userkafka "social-networking-platform/users-service/internal/repository/kafka"
	pgrepo "social-networking-platform/users-service/internal/repository/postgres"
	"social-networking-platform/users-service/internal/service"
	httptransport "social-networking-platform/users-service/internal/transport/http"
)

// Requires Postgres via INTEGRATION_PG_DSN, e.g.
//
//	set INTEGRATION_PG_DSN=postgres://postgres:postgres@localhost:5433/users_db?sslmode=disable
//
//	cd users-service
//
//	go test -tags=integration -count=1 ./internal/integration/

func integrationDSN(t *testing.T) string {
	t.Helper()
	dsn := strings.TrimSpace(os.Getenv("INTEGRATION_PG_DSN"))
	if dsn == "" {
		t.Skip(`set INTEGRATION_PG_DSN to run integration tests, e.g. postgres://postgres:postgres@localhost:5433/users_db?sslmode=disable`)
	}
	return dsn
}

func migrationsDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	abs, err := filepath.Abs(filepath.Join(wd, "..", "..", "migrations"))
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func truncateTables(t *testing.T, db *sql.DB) {
	t.Helper()
	if _, err := db.Exec(`TRUNCATE TABLE users CASCADE`); err != nil {
		t.Fatalf("truncate: %v", err)
	}
}

func stackRouter(tb testing.TB, db *sql.DB) http.Handler {
	tb.Helper()
	userRepo := pgrepo.NewSQLUserRepository(db)
	followRepo := pgrepo.NewSQLFollowRepository(db)
	pub := userkafka.NewStubFollowProducer()
	svc := service.NewUserService(userRepo, followRepo, pub)
	h := handlers.NewUserHandler(svc)
	return middleware.RequestID(
		httptransport.NewRouter("users-integration", h),
	)
}

func withReqID(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), middleware.RequestIDKey, "int-req"))
}

func TestIntegration_ProfileAndFollowSmoke(t *testing.T) {
	dsn := integrationDSN(t)

	migAbs := migrationsDir(t)
	if err := dbmigrate.PostgresUp(dsn, migAbs); err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("postgres not reachable (%v); skipping", err)
	}

	truncateTables(t, db)
	h := stackRouter(t, db)

	rec := httptest.NewRecorder()
	r := withReqID(httptest.NewRequest(http.MethodGet, "/api/v1/users/me", http.NoBody))
	r.Header.Set("X-User-ID", "user:alice")
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("GetMe expected 200 got %d %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	r = withReqID(httptest.NewRequest(http.MethodPatch, "/api/v1/users/me", strings.NewReader(`{"name":"Alice","bio":"hi"}`)))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", "user:alice")
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("UpdateMe expected 200 got %d %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	r = withReqID(httptest.NewRequest(http.MethodPost, "/api/v1/users/user:bob/follow", http.NoBody))
	r.Header.Set("X-User-ID", "user:alice")
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Follow expected 204 got %d %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	r = withReqID(httptest.NewRequest(http.MethodPost, "/api/v1/users/user:bob/follow", http.NoBody))
	r.Header.Set("X-User-ID", "user:alice")
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("duplicate Follow expected 204 got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	r = withReqID(httptest.NewRequest(http.MethodGet, "/api/v1/users/user:bob", http.NoBody))
	r.Header.Set("X-User-ID", "user:alice")
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("GetByID bob expected 200 got %d %s", rec.Code, rec.Body.String())
	}
	var envelope apiresponse.SuccessEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if !envelope.Success {
		t.Fatal("invalid success envelope")
	}

	rec = httptest.NewRecorder()
	r = withReqID(httptest.NewRequest(http.MethodPost, "/api/v1/users/user:alice/follow", http.NoBody))
	r.Header.Set("X-User-ID", "user:alice")
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("self follow expected 403 got %d", rec.Code)
	}
}
