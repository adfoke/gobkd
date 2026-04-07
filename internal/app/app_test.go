package app

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"path/filepath"
	"testing"

	authx "gobkd/internal/auth"
	"gobkd/internal/config"
	"gobkd/internal/logger"
)

func TestPingAndHealthz(t *testing.T) {
	router := newTestRouter(t)

	for _, tc := range []struct {
		name string
		path string
	}{
		{name: "ping", path: "/ping"},
		{name: "healthz", path: "/healthz"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d, body=%s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestAuthLoginAndMe(t *testing.T) {
	router := newTestRouter(t)
	server := httptest.NewServer(router)
	defer server.Close()

	client := newTestClient(t)
	login(client, t, server.URL)

	meResp, err := client.Get(server.URL + "/api/v1/me")
	if err != nil {
		t.Fatalf("me request failed: %v", err)
	}
	defer meResp.Body.Close()

	if meResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from /api/v1/me, got %d", meResp.StatusCode)
	}
}

func TestUnauthorizedResponse(t *testing.T) {
	router := newTestRouter(t)
	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/me")
	if err != nil {
		t.Fatalf("unauthorized request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}

	assertBodyContains(t, resp.Body, `"error":"unauthorized"`)
}

func TestValidationResponse(t *testing.T) {
	router := newTestRouter(t)
	server := httptest.NewServer(router)
	defer server.Close()

	client := newTestClient(t)
	login(client, t, server.URL)

	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/v1/echo", bytes.NewReader([]byte(`{}`)))
	if err != nil {
		t.Fatalf("create echo request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("echo request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", resp.StatusCode)
	}

	assertBodyContains(t, resp.Body, `"error":"validation_failed"`)
}

func TestNotFoundResponse(t *testing.T) {
	router := newTestRouter(t)
	server := httptest.NewServer(router)
	defer server.Close()

	resp, err := http.Get(server.URL + "/missing")
	if err != nil {
		t.Fatalf("not found request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}

	assertBodyContains(t, resp.Body, `"error":"not_found"`)
}

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()

	cfg := config.Config{
		AppEnv:        "dev",
		HTTPAddr:      ":0",
		AppBaseURL:    "http://127.0.0.1:8080",
		SQLitePath:    filepath.Join(t.TempDir(), "test.db"),
		AuthSecret:    "test-secret",
		AuthLocalUser: "admin",
		AuthLocalPass: "admin",
		LogLevel:      "error",
	}

	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		t.Fatalf("init logger: %v", err)
	}

	db, err := openDB(cfg.SQLitePath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := runMigrations(context.Background(), db); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	return buildRouter(db, log, authx.New(cfg))
}

func newTestClient(t *testing.T) *http.Client {
	t.Helper()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("create cookie jar: %v", err)
	}

	return &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func login(client *http.Client, t *testing.T, baseURL string) {
	t.Helper()

	loginBody := map[string]string{
		"user":   "admin",
		"passwd": "admin",
	}

	payload, err := json.Marshal(loginBody)
	if err != nil {
		t.Fatalf("marshal login body: %v", err)
	}

	loginReq, err := http.NewRequest(http.MethodPost, baseURL+"/auth/local/login?session=1", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("create login request: %v", err)
	}
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, err := client.Do(loginReq)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode >= http.StatusBadRequest {
		t.Fatalf("expected successful login, got %d", loginResp.StatusCode)
	}
}

func assertBodyContains(t *testing.T, body io.Reader, want string) {
	t.Helper()

	data, err := io.ReadAll(body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	if !bytes.Contains(data, []byte(want)) {
		t.Fatalf("expected body to contain %q, got %s", want, string(data))
	}
}
