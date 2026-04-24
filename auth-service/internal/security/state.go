package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type StateManager struct {
	secret []byte
	ttl    time.Duration
}

type statePayload struct {
	Nonce     string `json:"nonce"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

func NewStateManager(secret string, ttl time.Duration) *StateManager {
	return &StateManager{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (m *StateManager) Generate() (string, error) {
	nonce, err := randomToken(16)
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	payload := statePayload{
		Nonce:     nonce,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(m.ttl).Unix(),
	}

	encodedPayload, err := encodeSegment(payload)
	if err != nil {
		return "", err
	}
	signature := signHMAC(encodedPayload, m.secret)
	return encodedPayload + "." + signature, nil
}

func (m *StateManager) Validate(state string) error {
	parts := strings.Split(state, ".")
	if len(parts) != 2 {
		return fmt.Errorf("state format is invalid")
	}

	expected := signHMAC(parts[0], m.secret)
	if !hmac.Equal([]byte(expected), []byte(parts[1])) {
		return fmt.Errorf("state signature is invalid")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return fmt.Errorf("decode state: %w", err)
	}

	var payload statePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return fmt.Errorf("parse state: %w", err)
	}
	if payload.ExpiresAt < time.Now().UTC().Unix() {
		return fmt.Errorf("state has expired")
	}
	return nil
}

func randomToken(size int) (string, error) {
	raw := make([]byte, size)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate random token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func signHMAC(message string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(message))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func encodeSegment(value interface{}) (string, error) {
	bytes, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("marshal segment: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
