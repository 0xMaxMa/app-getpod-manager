package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
)

type Handler struct {
	apiKey string
	cpuMu  sync.RWMutex
	cpu    *cpuStat
}

type cpuStat struct {
	Cores        int
	UsagePercent float64
}

func New(apiKey string) *Handler {
	h := &Handler{apiKey: apiKey}
	go h.runCPULoop()
	return h
}

func (h *Handler) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != h.apiKey {
			jsonErr(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok", "version": "1.0.0"})
}

func jsonErr(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
