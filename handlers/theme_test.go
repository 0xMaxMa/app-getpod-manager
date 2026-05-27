package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func setupThemeTest(t *testing.T) (h *Handler, cleanup func()) {
	t.Helper()
	dir := t.TempDir()
	origPath := codeServerSettingsPath
	codeServerSettingsPath = filepath.Join(dir, "User", "settings.json")
	h = &Handler{apiKey: "test-key"}
	return h, func() { codeServerSettingsPath = origPath }
}

func TestSetTheme_DarkTheme(t *testing.T) {
	h, cleanup := setupThemeTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPatch, "/theme", bytes.NewBufferString(`{"theme":"dark"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.SetTheme(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal("failed to decode response:", err)
	}
	if resp["theme"] != "dark" {
		t.Errorf("expected theme=dark, got %s", resp["theme"])
	}
	if resp["colorTheme"] != "Default Dark Modern" {
		t.Errorf("expected colorTheme=Default Dark Modern, got %s", resp["colorTheme"])
	}

	// Verify settings.json was written with correct key
	data, err := os.ReadFile(codeServerSettingsPath)
	if err != nil {
		t.Fatal("settings.json not written:", err)
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatal("invalid settings.json:", err)
	}
	if settings["workbench.colorTheme"] != "Default Dark Modern" {
		t.Errorf("unexpected colorTheme in file: %v", settings["workbench.colorTheme"])
	}
}

func TestSetTheme_LightTheme(t *testing.T) {
	h, cleanup := setupThemeTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPatch, "/theme", bytes.NewBufferString(`{"theme":"light"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.SetTheme(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["colorTheme"] != "Default Light Modern" {
		t.Errorf("expected Default Light Modern, got %s", resp["colorTheme"])
	}
}

func TestSetTheme_MergesExistingSettings(t *testing.T) {
	h, cleanup := setupThemeTest(t)
	defer cleanup()

	// Pre-populate settings.json with existing keys
	if err := os.MkdirAll(filepath.Dir(codeServerSettingsPath), 0755); err != nil {
		t.Fatal(err)
	}
	existing := map[string]interface{}{"editor.fontSize": 14, "editor.tabSize": 2}
	data, _ := json.MarshalIndent(existing, "", "  ")
	if err := os.WriteFile(codeServerSettingsPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPatch, "/theme", bytes.NewBufferString(`{"theme":"dark"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.SetTheme(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var settings map[string]interface{}
	saved, _ := os.ReadFile(codeServerSettingsPath)
	json.Unmarshal(saved, &settings)

	// Existing keys must be preserved
	if settings["editor.fontSize"] != float64(14) {
		t.Errorf("editor.fontSize was clobbered: %v", settings["editor.fontSize"])
	}
	if settings["workbench.colorTheme"] != "Default Dark Modern" {
		t.Errorf("colorTheme not set: %v", settings["workbench.colorTheme"])
	}
}

func TestSetTheme_InvalidTheme(t *testing.T) {
	h, cleanup := setupThemeTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPatch, "/theme", bytes.NewBufferString(`{"theme":"blue"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.SetTheme(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["error"] == "" {
		t.Error("expected error message")
	}
}

func TestSetTheme_InvalidJSON(t *testing.T) {
	h, cleanup := setupThemeTest(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPatch, "/theme", bytes.NewBufferString(`not-json`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.SetTheme(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
