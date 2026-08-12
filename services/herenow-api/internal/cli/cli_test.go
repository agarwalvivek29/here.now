package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// TestPublishRemoteSendsBearerAndReturnsURL exercises the remote-publish helper
// against a stub server: it must POST the file body, carry the id_token as a
// Bearer credential and a title query, and return the `url` from the response.
func TestPublishRemoteSendsBearerAndReturnsURL(t *testing.T) {
	const token = "id-token-xyz"
	const wantURL = "https://here.now/a/abc123"
	const payload = "<h1>hi here.now</h1>"

	var gotMethod, gotAuth, gotTitle, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotAuth = r.Header.Get("Authorization")
		gotTitle = r.URL.Query().Get("title")
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"slug": "abc123", "url": wantURL})
	}))
	defer srv.Close()

	path := filepath.Join(t.TempDir(), "index.html")
	if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	gotURL, err := publishRemote(srv.URL, token, path)
	if err != nil {
		t.Fatalf("publishRemote: %v", err)
	}
	if gotURL != wantURL {
		t.Fatalf("url = %q, want %q", gotURL, wantURL)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("method = %q, want POST", gotMethod)
	}
	if gotAuth != "Bearer "+token {
		t.Fatalf("Authorization = %q, want %q", gotAuth, "Bearer "+token)
	}
	if gotTitle != "index.html" {
		t.Fatalf("title = %q, want %q", gotTitle, "index.html")
	}
	if gotBody != payload {
		t.Fatalf("uploaded body = %q, want %q", gotBody, payload)
	}
}

// TestPublishRemoteErrorsOnNon201 confirms the helper surfaces a non-201 server
// response (e.g. a rejected token) as an error rather than a bogus URL.
func TestPublishRemoteErrorsOnNon201(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	path := filepath.Join(t.TempDir(), "index.html")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	if _, err := publishRemote(srv.URL, "tok", path); err == nil {
		t.Fatalf("publishRemote: expected error on 401, got nil")
	}
}
