package handlers

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

const authorizedKeysPath = "/host-ssh/authorized_keys"

type sshKey struct {
	Fingerprint string `json:"fingerprint"`
	Comment     string `json:"comment"`
	Raw         string `json:"raw"`
}

func parseAuthorizedKeys() ([]sshKey, error) {
	f, err := os.Open(authorizedKeysPath)
	if os.IsNotExist(err) {
		return []sshKey{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var keys []sshKey
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		pub, comment, _, _, err := ssh.ParseAuthorizedKey([]byte(line))
		if err != nil {
			continue
		}
		keys = append(keys, sshKey{
			Fingerprint: ssh.FingerprintSHA256(pub),
			Comment:     comment,
			Raw:         line,
		})
	}
	if keys == nil {
		keys = []sshKey{}
	}
	return keys, scanner.Err()
}

func (h *Handler) ListSSHKeys(w http.ResponseWriter, r *http.Request) {
	keys, err := parseAuthorizedKeys()
	if err != nil {
		jsonErr(w, "failed to read authorized_keys", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keys)
}

func (h *Handler) AddSSHKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Key) == "" {
		jsonErr(w, "key is required", http.StatusBadRequest)
		return
	}

	pub, comment, _, _, err := ssh.ParseAuthorizedKey([]byte(req.Key))
	if err != nil {
		jsonErr(w, "invalid public key format", http.StatusBadRequest)
		return
	}
	fp := ssh.FingerprintSHA256(pub)

	existing, _ := parseAuthorizedKeys()
	for _, k := range existing {
		if k.Fingerprint == fp {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"error": "key already exists", "fingerprint": fp})
			return
		}
	}

	f, err := os.OpenFile(authorizedKeysPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		jsonErr(w, "failed to write authorized_keys", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	f.WriteString(strings.TrimSpace(req.Key) + "\n")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sshKey{Fingerprint: fp, Comment: comment, Raw: strings.TrimSpace(req.Key)})
}

func (h *Handler) DeleteSSHKey(w http.ResponseWriter, r *http.Request) {
	fp, err := url.PathUnescape(r.PathValue("fingerprint"))
	if err != nil {
		jsonErr(w, "invalid fingerprint", http.StatusBadRequest)
		return
	}

	keys, err := parseAuthorizedKeys()
	if err != nil {
		jsonErr(w, "failed to read authorized_keys", http.StatusInternalServerError)
		return
	}

	var remaining []sshKey
	found := false
	for _, k := range keys {
		if k.Fingerprint == fp {
			found = true
		} else {
			remaining = append(remaining, k)
		}
	}
	if !found {
		jsonErr(w, "key not found", http.StatusNotFound)
		return
	}

	f, err := os.OpenFile(authorizedKeysPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		jsonErr(w, "failed to write authorized_keys", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	for _, k := range remaining {
		f.WriteString(k.Raw + "\n")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"deleted": true, "fingerprint": fp})
}
