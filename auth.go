package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const defaultTokenURL = "https://api.dropboxapi.com/oauth2/token"

// tokenResponse is the response from Dropbox's /oauth2/token endpoint.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

// refreshAccessToken uses a refresh token to get a fresh short-lived access token.
func refreshAccessToken(tokenURL, appKey, appSecret, refreshToken string) (string, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {appKey},
		"client_secret": {appSecret},
	}

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return "", fmt.Errorf("refresh request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("token refresh failed (status %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result tokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}
	return result.AccessToken, nil
}

// resolveToken gets an access token using the priority chain:
// 1. DROPBOX_TOKEN env var (direct)
// 2. Refresh token + app key/secret (auto-refresh)
// 3. Error
func resolveToken(cfg *Config) (string, error) {
	if token := os.Getenv("DROPBOX_TOKEN"); token != "" {
		return token, nil
	}

	if cfg.RefreshToken != "" && cfg.AppKey != "" && cfg.AppSecret != "" {
		token, err := refreshAccessToken(defaultTokenURL, cfg.AppKey, cfg.AppSecret, cfg.RefreshToken)
		if err != nil {
			return "", fmt.Errorf("refresh token invalid, run: dropbox-appender auth\n  (%w)", err)
		}
		return token, nil
	}

	return "", fmt.Errorf("no authentication configured, run: dropbox-appender auth")
}
