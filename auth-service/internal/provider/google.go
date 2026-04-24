package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"social-networking-platform/auth-service/internal/domain"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	IDToken     string `json:"id_token"`
}

type OAuthProvider interface {
	AuthCodeURL(state string) (string, error)
	ExchangeCode(ctx context.Context, code string) (*TokenResponse, error)
	FetchUser(ctx context.Context, accessToken string) (*domain.AuthUser, error)
}

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
}

type GoogleProvider struct {
	cfg    GoogleConfig
	client *http.Client
}

func NewGoogleProvider(cfg GoogleConfig, client *http.Client) *GoogleProvider {
	return &GoogleProvider{cfg: cfg, client: client}
}

func (p *GoogleProvider) AuthCodeURL(state string) (string, error) {
	if p.cfg.ClientID == "" || p.cfg.RedirectURL == "" || p.cfg.AuthURL == "" {
		return "", fmt.Errorf("Google OAuth provider is not configured")
	}
	values := url.Values{}
	values.Set("client_id", p.cfg.ClientID)
	values.Set("redirect_uri", p.cfg.RedirectURL)
	values.Set("response_type", "code")
	values.Set("scope", "openid email profile")
	values.Set("state", state)
	values.Set("access_type", "online")
	return p.cfg.AuthURL + "?" + values.Encode(), nil
}

func (p *GoogleProvider) ExchangeCode(ctx context.Context, code string) (*TokenResponse, error) {
	if p.cfg.ClientID == "" || p.cfg.ClientSecret == "" || p.cfg.RedirectURL == "" || p.cfg.TokenURL == "" {
		return nil, fmt.Errorf("Google OAuth provider is not configured")
	}
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", p.cfg.ClientID)
	form.Set("client_secret", p.cfg.ClientSecret)
	form.Set("redirect_uri", p.cfg.RedirectURL)
	form.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchange auth code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var upstream map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&upstream)
		return nil, fmt.Errorf("google token endpoint returned %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}
	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("google token response missing access token")
	}
	return &tokenResp, nil
}

func (p *GoogleProvider) FetchUser(ctx context.Context, accessToken string) (*domain.AuthUser, error) {
	if p.cfg.UserInfoURL == "" {
		return nil, fmt.Errorf("Google userinfo endpoint is not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.cfg.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch google user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo endpoint returned %d", resp.StatusCode)
	}

	var payload struct {
		Subject string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode userinfo response: %w", err)
	}
	if payload.Subject == "" || payload.Email == "" {
		return nil, fmt.Errorf("google userinfo response missing subject or email")
	}

	return &domain.AuthUser{
		ID:              "google:" + payload.Subject,
		Provider:        "google",
		ProviderSubject: payload.Subject,
		Email:           payload.Email,
		Name:            payload.Name,
		ProfilePicURL:   payload.Picture,
	}, nil
}
