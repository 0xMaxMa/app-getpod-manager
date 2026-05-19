package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

var socketClient = &http.Client{
	Transport: &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", "/run/gateway.sock")
		},
	},
	Timeout: 65 * time.Second,
}

type scriptResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exitCode"`
}

func callScript(name string, args map[string]string) (string, error) {
	if args == nil {
		args = map[string]string{}
	}
	body, _ := json.Marshal(map[string]interface{}{"args": args})
	resp, err := socketClient.Post(
		"http://gateway/tool/script/"+name,
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result scriptResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.ExitCode != 0 {
		return "", fmt.Errorf("%s: %s", name, result.Stderr)
	}
	return result.Stdout, nil
}
