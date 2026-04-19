package vault

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newFakeVaultServer(t *testing.T, path string, payload map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": payload})
	}))
}

func TestSplitMountAndPath(t *testing.T) {
	cases := []struct {
		input     string
		expMount  string
		expSub    string
		wantErr   bool
	}{
		{"secret/myapp/prod", "secret", "myapp/prod", false},
		{"kv/db", "kv", "db", false},
		{"onlymount", "", "", true},
		{"/secret/app", "secret", "app", false},
	}

	for _, tc := range cases {
		m, s, err := splitMountAndPath(tc.input)
		if tc.wantErr {
			if err == nil {
				t.Errorf("splitMountAndPath(%q): expected error, got nil", tc.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("splitMountAndPath(%q): unexpected error: %v", tc.input, err)
		}
		if m != tc.expMount || s != tc.expSub {
			t.Errorf("splitMountAndPath(%q) = (%q, %q), want (%q, %q)", tc.input, m, s, tc.expMount, tc.expSub)
		}
	}
}

func TestFlattenData(t *testing.T) {
	input := map[string]interface{}{
		"DB_PASS": "hunter2",
		"PORT":    8080,
		"ENABLED": true,
	}
	out := flattenData(input)
	if out["DB_PASS"] != "hunter2" {
		t.Errorf("expected DB_PASS=hunter2, got %q", out["DB_PASS"])
	}
	if out["PORT"] != "8080" {
		t.Errorf("expected PORT=8080, got %q", out["PORT"])
	}
	if out["ENABLED"] != "true" {
		t.Errorf("expected ENABLED=true, got %q", out["ENABLED"])
	}
}

func TestNewClient_InvalidAddress(t *testing.T) {
	_, err := NewClient("://bad-address", "token")
	if err == nil {
		t.Error("expected error for invalid address, got nil")
	}
}

func TestGetSecrets_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.GetSecrets(context.Background(), "secret/missing")
	if err == nil {
		t.Error("expected error for missing secret, got nil")
	}
}
