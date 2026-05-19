package main

import (
	"log"
	"net/http"
	"os"

	"github.com/0xMaxMa/getpod-manager/handlers"
)

func main() {
	apiKey := os.Getenv("GATEWAY_API_KEY")
	if apiKey == "" {
		log.Fatal("GATEWAY_API_KEY is required")
	}

	h := handlers.New(apiKey)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.Health)
	mux.Handle("GET /metrics", h.Auth(http.HandlerFunc(h.Metrics)))
	mux.Handle("POST /resize", h.Auth(http.HandlerFunc(h.Resize)))
	mux.Handle("GET /ssh-keys", h.Auth(http.HandlerFunc(h.ListSSHKeys)))
	mux.Handle("POST /ssh-keys", h.Auth(http.HandlerFunc(h.AddSSHKey)))
	mux.Handle("DELETE /ssh-keys/{fingerprint}", h.Auth(http.HandlerFunc(h.DeleteSSHKey)))

	log.Println("listening on :5990")
	log.Fatal(http.ListenAndServe(":5990", mux))
}
