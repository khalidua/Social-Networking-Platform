package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
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

type TokenVerifier struct {
	secret []byte
	issuer string
}

func NewTokenVerifier(secret, issuer string) *TokenVerifier {
	return &TokenVerifier{
		secret: []byte(secret),
		issuer: issuer,
	}
}

func (v *TokenVerifier) Parse(token string) (*TokenClaims, error) {
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
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("parse token header: %w", err)
	}
	if header.Algorithm != "HS256" {
		return nil, fmt.Errorf("token algorithm is not supported")
	}

	expected := signJWT(parts[0]+"."+parts[1], v.secret)
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
	if claims.Issuer != "" && v.issuer != "" && claims.Issuer != v.issuer {
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
