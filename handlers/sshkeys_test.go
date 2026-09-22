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

func setupSSHTest(t *testing.T) (h *Handler, cleanup func()) {
	t.Helper()
	dir := t.TempDir()
	origPath := authorizedKeysPath
	authorizedKeysPath = filepath.Join(dir, "authorized_keys")
	h = &Handler{apiKey: "test-key"}
	return h, func() { authorizedKeysPath = origPath }
}

// TestAddSSHKey_NoTrailingNewline is the regression test for the bug:
// when authorized_keys has no trailing newline, appending without a leading \n
// concatenates the new key onto the last line, making both keys invalid.
func TestAddSSHKey_NoTrailingNewline(t *testing.T) {
	h, cleanup := setupSSHTest(t)
	defer cleanup()

	// Write first key WITHOUT a trailing newline — this is the problematic state.
	if err := os.WriteFile(authorizedKeysPath, []byte(testKey1), 0600); err != nil {
		t.Fatal(err)
	}

	// Show hex tail of file before
	before, _ := os.ReadFile(authorizedKeysPath)
	t.Logf("BEFORE last 4 bytes hex: %x  (no 0a = no trailing newline)", before[max(0, len(before)-4):])
	t.Logf("BEFORE content:\n%s", string(before))

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
	t.Logf("AFTER content:\n%s", string(after))

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
	h, cleanup := setupSSHTest(t)
	defer cleanup()

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

	after, _ := os.ReadFile(authorizedKeysPath)
	t.Logf("AFTER content:\n%s", string(after))

	keys, err := parseAuthorizedKeys()
	if err != nil {
		t.Fatal("parseAuthorizedKeys:", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 valid keys, got %d", len(keys))
	}
}

func TestAddSSHKey_DuplicateRejected(t *testing.T) {
	h, cleanup := setupSSHTest(t)
	defer cleanup()

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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
