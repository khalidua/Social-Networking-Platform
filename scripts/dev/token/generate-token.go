package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

func main() {
	secret := "secret"
	issuer := "auth-service"
	subject := "google:123"
	sessionID := "session-1"
	email := "user@example.com"

	header, _ := json.Marshal(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})

	payload, _ := json.Marshal(map[string]interface{}{
		"iss":   issuer,
		"sub":   subject,
		"sid":   sessionID,
		"email": email,
		"iat":   time.Now().UTC().Unix(),
		"exp":   time.Now().UTC().Add(1 * time.Hour).Unix(),
	})

	headerSegment := base64.RawURLEncoding.EncodeToString(header)
	payloadSegment := base64.RawURLEncoding.EncodeToString(payload)

	signingInput := headerSegment + "." + payloadSegment

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	token := signingInput + "." + signature

	fmt.Println("TOKEN:")
	fmt.Println(token)
}