package cloudauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OIDCMetadata is a subset of /.well-known/openid-configuration.
type OIDCMetadata struct {
	Issuer                        string `json:"issuer"`
	DeviceAuthorizationEndpoint   string `json:"device_authorization_endpoint"`
	TokenEndpoint                 string `json:"token_endpoint"`
}

// FetchOIDCMetadata loads discovery JSON from issuer (trailing slash optional).
func FetchOIDCMetadata(ctx context.Context, issuer string, hc *http.Client) (*OIDCMetadata, error) {
	if hc == nil {
		hc = http.DefaultClient
	}
	base := strings.TrimSuffix(strings.TrimSpace(issuer), "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/.well-known/openid-configuration", nil)
	if err != nil {
		return nil, err
	}
	res, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
		return nil, fmt.Errorf("discovery: status %d: %s", res.StatusCode, strings.TrimSpace(string(b)))
	}
	var meta OIDCMetadata
	if err := json.NewDecoder(res.Body).Decode(&meta); err != nil {
		return nil, err
	}
	if meta.DeviceAuthorizationEndpoint == "" || meta.TokenEndpoint == "" {
		return nil, fmt.Errorf("OIDC metadata missing device_authorization_endpoint or token_endpoint")
	}
	return &meta, nil
}

// DeviceAuthorizationResponse is the initial device flow payload (RFC 8628).
type DeviceAuthorizationResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// StartDeviceFlow requests a device code from the authorization server.
func StartDeviceFlow(ctx context.Context, meta *OIDCMetadata, clientID, scope string, hc *http.Client) (*DeviceAuthorizationResponse, error) {
	if hc == nil {
		hc = http.DefaultClient
	}
	if scope == "" {
		scope = "openid profile email offline_access"
	}
	form := url.Values{}
	form.Set("client_id", clientID)
	form.Set("scope", scope)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, meta.DeviceAuthorizationEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device auth: status %d: %s", res.StatusCode, strings.TrimSpace(string(b)))
	}
	var out DeviceAuthorizationResponse
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	if out.DeviceCode == "" {
		return nil, fmt.Errorf("device auth: empty device_code: %s", strings.TrimSpace(string(b)))
	}
	return &out, nil
}

// TokenResponse holds OAuth2 token endpoint fields used for cloud API access.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	Scope        string `json:"scope"`
}

// PollDeviceToken exchanges device_code for tokens until success, expiry, or ctx cancel.
func PollDeviceToken(ctx context.Context, meta *OIDCMetadata, clientID, deviceCode string, dev *DeviceAuthorizationResponse, hc *http.Client) (*TokenResponse, error) {
	if hc == nil {
		hc = http.DefaultClient
	}
	interval := time.Duration(dev.Interval) * time.Second
	if interval <= 0 {
		interval = 5 * time.Second
	}
	deadline := time.Now().Add(time.Duration(dev.ExpiresIn) * time.Second)
	if dev.ExpiresIn <= 0 {
		deadline = time.Now().Add(15 * time.Minute)
	}

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
		form := url.Values{}
		form.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
		form.Set("device_code", deviceCode)
		form.Set("client_id", clientID)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, meta.TokenEndpoint, strings.NewReader(form.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		res, err := hc.Do(req)
		if err != nil {
			return nil, err
		}
		b, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
		_ = res.Body.Close()
		if res.StatusCode == http.StatusOK {
			var tok TokenResponse
			if err := json.Unmarshal(b, &tok); err != nil {
				return nil, err
			}
			return &tok, nil
		}
		var errBody struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		_ = json.Unmarshal(b, &errBody)
		switch errBody.Error {
		case "authorization_pending", "slow_down":
			if errBody.Error == "slow_down" {
				interval += 5 * time.Second
			}
			continue
		default:
			msg := strings.TrimSpace(errBody.ErrorDescription)
			if msg == "" {
				msg = string(b)
			}
			return nil, fmt.Errorf("token: %s (%s)", errBody.Error, msg)
		}
	}
	return nil, fmt.Errorf("device flow expired before authorization completed")
}
