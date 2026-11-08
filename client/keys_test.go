package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"strings"
	"testing"
)

// The repository ships the public key at the root. This test makes sure the
// private key embedded in the client and publickey.pem are still the same pair
// after someone rotates keys.
func TestPublicKeyFileMatchesEmbeddedPrivateKey(t *testing.T) {
	privBlock, _ := pem.Decode([]byte(strings.TrimSpace(privateKeyPEM)))
	if privBlock == nil {
		t.Fatal("cannot decode embedded private key")
	}
	priv, err := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
	if err != nil {
		t.Fatalf("parse embedded private key: %v", err)
	}

	privFile, err := os.ReadFile("../privatekey.pem")
	if err != nil {
		t.Fatalf("read privatekey.pem: %v", err)
	}
	if strings.TrimSpace(string(privFile)) != strings.TrimSpace(privateKeyPEM) {
		t.Fatal("privatekey.pem does not match the private key embedded in token.go")
	}

	data, err := os.ReadFile("../publickey.pem")
	if err != nil {
		t.Fatalf("read publickey.pem: %v", err)
	}
	pubBlock, _ := pem.Decode(data)
	if pubBlock == nil {
		t.Fatal("cannot decode ../publickey.pem")
	}
	pubAny, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
	if err != nil {
		t.Fatalf("parse publickey.pem: %v", err)
	}
	pub, ok := pubAny.(*rsa.PublicKey)
	if !ok {
		t.Fatal("publickey.pem is not an RSA public key")
	}

	if priv.N.Cmp(pub.N) != 0 || priv.E != pub.E {
		t.Fatal("publickey.pem does not match the private key embedded in token.go")
	}
}
