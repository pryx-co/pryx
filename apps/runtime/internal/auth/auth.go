package auth

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// PKCEParams holds PKCE (Proof Key for Code Exchange) parameters
// RFC 7636 - Used to secure the device authorization grant
type PKCEParams struct {
	CodeVerifier  string `json:"code_verifier"`
	CodeChallenge string `json:"code_challenge"`
	Method        string `json:"code_challenge_method"`
}

// GeneratePKCE generates PKCE parameters for OAuth device flow
// Returns code verifier, code challenge, and method (S256)
func GeneratePKCE() (*PKCEParams, error) {
	// Generate code verifier: 43-128 characters (we use 128)
	verifierBytes := make([]byte, 96) // 96 bytes = 128 base64url chars
	if _, err := rand.Read(verifierBytes); err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}
	codeVerifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	// Generate code challenge: BASE64URL(SHA256(code_verifier))
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return &PKCEParams{
		CodeVerifier:  codeVerifier,
		CodeChallenge: codeChallenge,
		Method:        "S256",
	}, nil
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type ErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description,omitempty"`
}

// DeviceFlowRequest represents a device authorization request with optional PKCE
type DeviceFlowRequest struct {
	ClientID string      `json:"client_id,omitempty"`
	PKCE     *PKCEParams `json:"-"` // Not sent in JSON, used internally
}

// StartDeviceFlow initiates the OAuth 2.0 device authorization grant (RFC 8628)
// with optional PKCE support (RFC 7636) for enhanced security
func StartDeviceFlow(apiUrl string, pkce *PKCEParams) (*DeviceCodeResponse, error) {
	var reqBody []byte
	var err error

	if pkce != nil {
		// Include PKCE challenge in the request
		reqData := map[string]string{
			"code_challenge":        pkce.CodeChallenge,
			"code_challenge_method": pkce.Method,
		}
		reqBody, err = json.Marshal(reqData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/auth/device/code", apiUrl),
		"application/json",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return nil, fmt.Errorf("cloud error: %s", errResp.Error)
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var res DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &res, nil
}

// StartDeviceFlowWithPKCE initiates device flow with PKCE enabled (recommended)
func StartDeviceFlowWithPKCE(apiUrl string) (*DeviceCodeResponse, *PKCEParams, error) {
	pkce, err := GeneratePKCE()
	if err != nil {
		return nil, nil, err
	}

	resp, err := StartDeviceFlow(apiUrl, pkce)
	if err != nil {
		return nil, nil, err
	}

	return resp, pkce, nil
}

// PollForToken polls for OAuth token (legacy, without PKCE)
// Deprecated: Use PollForTokenWithPKCE for enhanced security
func PollForToken(apiUrl string, deviceCode string, interval int) (*TokenResponse, error) {
	return PollForTokenWithContext(context.Background(), apiUrl, deviceCode, interval)
}

// PollForTokenWithContext polls for token with context (legacy, without PKCE)
// Deprecated: Use PollForTokenWithPKCE for enhanced security
func PollForTokenWithContext(ctx context.Context, apiUrl string, deviceCode string, interval int) (*TokenResponse, error) {
	return PollForTokenWithPKCE(ctx, apiUrl, deviceCode, interval, "")
}

func requestToken(apiUrl string, deviceCode string, pkceVerifier string) (*TokenResponse, error) {
	reqData := map[string]string{
		"device_code": deviceCode,
		"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
	}

	// Include PKCE verifier if provided (required when PKCE was used in initial request)
	if pkceVerifier != "" {
		reqData["code_verifier"] = pkceVerifier
	}

	payload, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal token request: %w", err)
	}

	resp, err := http.Post(
		fmt.Sprintf("%s/auth/device/token", apiUrl),
		"application/json",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return nil, fmt.Errorf("cloud error: %s", errResp.Error)
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var res TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}
	return &res, nil
}

// PollForTokenWithPKCE polls for token with optional PKCE verifier
func PollForTokenWithPKCE(ctx context.Context, apiUrl string, deviceCode string, interval int, pkceVerifier string) (*TokenResponse, error) {
	if interval <= 0 {
		interval = 5
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			token, err := requestToken(apiUrl, deviceCode, pkceVerifier)
			if err != nil {
				// Continue polling if it's "authorization_pending"
				if err.Error() == "cloud error: authorization_pending" {
					continue
				}
				return nil, err
			}
			return token, nil
		}
	}
}
