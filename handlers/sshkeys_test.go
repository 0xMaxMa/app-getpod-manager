package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	testKey1 = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFa+9EtJ6rJpoRdgiqPgUejuPmbudJ1n1D8R2m9ScJ/w test1@test"
	testKey2 = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIO+AWw+i9Ga9o5gJm3g1EONbL3T5vuL3eOFhejnCtqoq test2@test"
)

func setupSSHTest(t *testing.T) *Handler {
	t.Helper()
	origPath := authorizedKeysPath
	authorizedKeysPath = filepath.Join(t.TempDir(), "authorized_keys")
	t.Cleanup(func() { authorizedKeysPath = origPath })
	return &Handler{apiKey: "test-key"}
}

// TestAddSSHKey_NoTrailingNewline is the regression test for the bug:
// when authorized_keys has no trailing newline, appending without a leading \n
// concatenates the new key onto the last line, making both keys invalid.
func TestAddSSHKey_NoTrailingNewline(t *testing.T) {
	h := setupSSHTest(t)

	// Write first key WITHOUT a trailing newline — this is the problematic state.
	if err := os.WriteFile(authorizedKeysPath, []byte(testKey1), 0600); err != nil {
		t.Fatal(err)
	}

	// Add second key via handler
	body, _ := json.Marshal(map[string]string{"key": testKey2})
	req := httptest.NewRequest(http.MethodPost, "/ssh-keys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.AddSSHKey(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	after, _ := os.ReadFile(authorizedKeysPath)

	// Count non-empty key lines
	var keyLines []string
	for _, l := range strings.Split(string(after), "\n") {
		if strings.TrimSpace(l) != "" {
			keyLines = append(keyLines, l)
		}
	}

	if len(keyLines) != 2 {
		t.Fatalf("expected 2 separate key lines, got %d: %v", len(keyLines), keyLines)
	}
	if keyLines[0] != testKey1 {
		t.Errorf("first key corrupted: %q", keyLines[0])
	}
	if keyLines[1] != testKey2 {
		t.Errorf("second key corrupted: %q", keyLines[1])
	}

	// Verify both keys parse correctly
	keys, err := parseAuthorizedKeys()
	if err != nil {
		t.Fatal("parseAuthorizedKeys:", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 parseable keys, got %d", len(keys))
	}
}

func TestAddSSHKey_WithTrailingNewline(t *testing.T) {
	h := setupSSHTest(t)

	// Write first key WITH trailing newline
	if err := os.WriteFile(authorizedKeysPath, []byte(testKey1+"\n"), 0600); err != nil {
		t.Fatal(err)
	}

	body, _ := json.Marshal(map[string]string{"key": testKey2})
	req := httptest.NewRequest(http.MethodPost, "/ssh-keys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.AddSSHKey(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	keys, err := parseAuthorizedKeys()
	if err != nil {
		t.Fatal("parseAuthorizedKeys:", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 valid keys, got %d", len(keys))
	}
}

func TestAddSSHKey_DuplicateRejected(t *testing.T) {
	h := setupSSHTest(t)

	for i, wantStatus := range []int{http.StatusCreated, http.StatusConflict} {
		body, _ := json.Marshal(map[string]string{"key": testKey1})
		req := httptest.NewRequest(http.MethodPost, "/ssh-keys", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		h.AddSSHKey(w, req)
		if w.Code != wantStatus {
			t.Errorf("attempt %d: expected %d, got %d: %s", i+1, wantStatus, w.Code, w.Body.String())
		}
	}
}

func postSSHKey(t *testing.T, h *Handler, key string) *httptest.ResponseRecorder {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"key": key})
	req := httptest.NewRequest(http.MethodPost, "/ssh-keys", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.AddSSHKey(w, req)
	return w
}

func readKeysFile(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(authorizedKeysPath)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestAddSSHKey_MultiLineRejected(t *testing.T) {
	for name, key := range map[string]string{
		"LF":           "garbage line\n" + testKey1 + "\n" + `command="/bin/evil" ` + testKey2,
		"CR":           testKey1 + "\r" + testKey2,
		"trailing LF+": testKey1 + "\n" + testKey2 + "\n",
	} {
		t.Run(name, func(t *testing.T) {
			h := setupSSHTest(t)
			w := postSSHKey(t, h, key)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
			}
			if _, err := os.Stat(authorizedKeysPath); !os.IsNotExist(err) {
				t.Errorf("authorized_keys must not be written, stat err=%v content=%q", err, readKeysFile(t))
			}
		})
	}
}

func TestAddSSHKey_RawMatchesWrittenLine(t *testing.T) {
	h := setupSSHTest(t)
	w := postSSHKey(t, h, "  "+testKey1+"\n")
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var resp sshKey
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if got := readKeysFile(t); got != resp.Raw+"\n" {
		t.Errorf("file %q does not match response raw %q", got, resp.Raw)
	}
}

func TestAddSSHKey_MissingFile(t *testing.T) {
	h := setupSSHTest(t)
	if w := postSSHKey(t, h, testKey1); w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if got, want := readKeysFile(t), testKey1+"\n"; got != want {
		t.Errorf("content = %q, want %q", got, want)
	}
}

func TestAddSSHKey_EmptyFile(t *testing.T) {
	h := setupSSHTest(t)
	if err := os.WriteFile(authorizedKeysPath, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if w := postSSHKey(t, h, testKey1); w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if got, want := readKeysFile(t), testKey1+"\n"; got != want {
		t.Errorf("content = %q, want %q", got, want)
	}
}

func TestAddSSHKey_NoBlankLineGrowth(t *testing.T) {
	h := setupSSHTest(t)
	for _, k := range []string{testKey1, testKey2} {
		if w := postSSHKey(t, h, k); w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
	}
	if got, want := readKeysFile(t), testKey1+"\n"+testKey2+"\n"; got != want {
		t.Errorf("content = %q, want %q", got, want)
	}
}
