package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
)

var codeServerSettingsPath = "/host-code-server/User/settings.json"

var themeMap = map[string]string{
	"light": "Default Light Modern",
	"dark":  "Default Dark Modern",
}

func (h *Handler) SetTheme(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Theme string `json:"theme"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonErr(w, "invalid request", http.StatusBadRequest)
		return
	}
	colorTheme, ok := themeMap[req.Theme]
	if !ok {
		jsonErr(w, "theme must be 'light' or 'dark'", http.StatusBadRequest)
		return
	}

	dir := filepath.Dir(codeServerSettingsPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		jsonErr(w, "failed to create settings directory", http.StatusInternalServerError)
		return
	}
	// Make dir writable by ubuntu so VS Code can atomically save settings (needs to create temp file)
	if err := os.Chmod(dir, 0777); err != nil {
		jsonErr(w, "failed to set directory permissions", http.StatusInternalServerError)
		return
	}

	settings := map[string]interface{}{}
	if data, err := os.ReadFile(codeServerSettingsPath); err == nil {
		_ = json.Unmarshal(data, &settings)
	}
	settings["workbench.colorTheme"] = colorTheme

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		jsonErr(w, "failed to encode settings", http.StatusInternalServerError)
		return
	}

	// Atomic rename: write temp then rename so code-server picks up the inotify
	// CREATE/MOVED_TO event reliably (IN_MODIFY from Docker bind mounts can be missed)
	tmpPath := codeServerSettingsPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		jsonErr(w, "failed to write settings", http.StatusInternalServerError)
		return
	}
	if err := os.Rename(tmpPath, codeServerSettingsPath); err != nil {
		os.Remove(tmpPath)
		jsonErr(w, "failed to write settings", http.StatusInternalServerError)
		return
	}
	// chmod 0666: root owns file but ubuntu (others) can read+write — CAP_CHOWN not available
	if err := os.Chmod(codeServerSettingsPath, 0666); err != nil {
		jsonErr(w, "failed to set file permissions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"theme": req.Theme, "colorTheme": colorTheme})
}
