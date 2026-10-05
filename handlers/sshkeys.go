package handlers

import (
	"bufio"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
)

var authorizedKeysPath = "/host-ssh/authorized_keys"

// authorizedKeysMu serializes read-modify-write of authorized_keys so
// concurrent add/delete requests can't lose or duplicate keys.
var authorizedKeysMu sync.Mutex

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
	authorizedKeysMu.Lock()
	keys, err := parseAuthorizedKeys()
	authorizedKeysMu.Unlock()
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
	// ParseAuthorizedKey only validates the first line; reject anything that
	// would smuggle extra lines into authorized_keys.
	line := strings.TrimSpace(req.Key)
	if strings.ContainsAny(line, "\r\n") {
		jsonErr(w, "key must be a single line", http.StatusBadRequest)
		return
	}

	pub, comment, _, _, err := ssh.ParseAuthorizedKey([]byte(line))
	if err != nil {
		jsonErr(w, "invalid public key format", http.StatusBadRequest)
		return
	}
	fp := ssh.FingerprintSHA256(pub)

	authorizedKeysMu.Lock()
	defer authorizedKeysMu.Unlock()

	existing, err := parseAuthorizedKeys()
	if err != nil {
		jsonErr(w, "failed to read authorized_keys", http.StatusInternalServerError)
		return
	}
	for _, k := range existing {
		if k.Fingerprint == fp {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"error": "key already exists", "fingerprint": fp})
			return
		}
	}

	f, err := os.OpenFile(authorizedKeysPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		jsonErr(w, "failed to write authorized_keys", http.StatusInternalServerError)
		return
	}
	defer f.Close()
	// Start on a fresh line if the file doesn't already end with one, otherwise
	// the new key would be concatenated onto the last existing line.
	entry := line + "\n"
	needsNewline, err := missingTrailingNewline(f)
	if err != nil {
		jsonErr(w, "failed to read authorized_keys", http.StatusInternalServerError)
		return
	}
	if needsNewline {
		entry = "\n" + entry
	}
	if _, err := f.WriteString(entry); err != nil {
		jsonErr(w, "failed to write authorized_keys", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sshKey{Fingerprint: fp, Comment: comment, Raw: line})
}

// missingTrailingNewline reports whether f is non-empty and its last byte is not '\n'.
func missingTrailingNewline(f *os.File) (bool, error) {
	info, err := f.Stat()
	if err != nil || info.Size() == 0 {
		return false, err
	}
	last := make([]byte, 1)
	if _, err := f.ReadAt(last, info.Size()-1); err != nil {
		return false, err
	}
	return last[0] != '\n', nil
}

func (h *Handler) DeleteSSHKey(w http.ResponseWriter, r *http.Request) {
	fp, err := url.PathUnescape(r.PathValue("fingerprint"))
	if err != nil {
		jsonErr(w, "invalid fingerprint", http.StatusBadRequest)
		return
	}

	authorizedKeysMu.Lock()
	defer authorizedKeysMu.Unlock()

	data, err := os.ReadFile(authorizedKeysPath)
	if err != nil && !os.IsNotExist(err) {
		jsonErr(w, "failed to read authorized_keys", http.StatusInternalServerError)
		return
	}

	// Drop only the lines holding the matching key; keep comments, blank and
	// unparseable lines exactly as they were.
	var kept strings.Builder
	found := false
	for _, line := range strings.SplitAfter(string(data), "\n") {
		if pub, _, _, _, err := ssh.ParseAuthorizedKey([]byte(line)); err == nil && ssh.FingerprintSHA256(pub) == fp {
			found = true
			continue
		}
		kept.WriteString(line)
	}
	if !found {
		jsonErr(w, "key not found", http.StatusNotFound)
		return
	}

	if err := os.WriteFile(authorizedKeysPath, []byte(kept.String()), 0600); err != nil {
		jsonErr(w, "failed to write authorized_keys", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"deleted": true, "fingerprint": fp})
}
