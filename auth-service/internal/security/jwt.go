package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"social-networking-platform/auth-service/internal/domain"
)

type TokenClaims struct {
	Issuer    string `json:"iss"`
	Subject   string `json:"sub"`
	SessionID string `json:"sid"`
	Email     string `json:"email"`
	Name      string `json:"name,omitempty"`
	Picture   string `json:"picture,omitempty"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

type TokenManager struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

func NewTokenManager(secret, issuer string, ttl time.Duration) *TokenManager {
	return &TokenManager{
		secret: []byte(secret),
		issuer: issuer,
		ttl:    ttl,
	}
}

func (m *TokenManager) Issue(user domain.AuthUser, session domain.Session) (string, error) {
	headerSegment, err := encodeSegment(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	claims := TokenClaims{
		Issuer:    m.issuer,
		Subject:   user.ID,
		SessionID: session.ID,
		Email:     user.Email,
		Name:      user.Name,
		Picture:   user.ProfilePicURL,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(m.ttl).Unix(),
	}

	payloadSegment, err := encodeSegment(claims)
	if err != nil {
		return "", err
	}

	signingInput := headerSegment + "." + payloadSegment
	signature := signJWT(signingInput, m.secret)
	return signingInput + "." + signature, nil
}

func (m *TokenManager) Parse(token string) (*TokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("token format is invalid")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decode token header: %w", err)
	}
	var header struct {
		Algorithm string `json:"alg"`
		Type      string `json:"typ"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("parse token header: %w", err)
	}
	if header.Algorithm != "HS256" {
		return nil, fmt.Errorf("token algorithm is not supported")
	}

	signingInput := parts[0] + "." + parts[1]
	expected := signJWT(signingInput, m.secret)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return nil, fmt.Errorf("token signature is invalid")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode token payload: %w", err)
	}

	var claims TokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("parse token payload: %w", err)
	}
	if claims.Subject == "" || claims.SessionID == "" {
		return nil, fmt.Errorf("token claims are incomplete")
	}
	if claims.Issuer != "" && m.issuer != "" && claims.Issuer != m.issuer {
		return nil, fmt.Errorf("token issuer is invalid")
	}
	if claims.ExpiresAt <= time.Now().UTC().Unix() {
		return nil, fmt.Errorf("token has expired")
	}
	return &claims, nil
}

func signJWT(message string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
