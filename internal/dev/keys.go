package dev

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateSecretKeyBase() (string, error) {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func GenerateAPIKey() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "nrn_" + base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b), nil
}
