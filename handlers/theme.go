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

	if err := os.MkdirAll(filepath.Dir(codeServerSettingsPath), 0755); err != nil {
		jsonErr(w, "failed to create settings directory", http.StatusInternalServerError)
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
	if err := os.WriteFile(codeServerSettingsPath, data, 0644); err != nil {
		jsonErr(w, "failed to write settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"theme": req.Theme, "colorTheme": colorTheme})
}
