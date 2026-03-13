package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"os"
	"sync"
)

var (
	rsaKey     *rsa.PrivateKey
	rsaKeyOnce sync.Once
	keyMu      sync.RWMutex
)

const (
	KeySize = 2048
)

func init() {
	loadOrGenerateKey()
}

func loadOrGenerateKey() {
	rsaKeyOnce.Do(func() {
		keyPath := "rsa_private_key.pem"
		pubKeyPath := "rsa_public_key.pem"

		var err error
		rsaKey, err = loadPrivateKey(keyPath)
		if err != nil {
			rsaKey, err = generateKey()
			if err != nil {
				return
			}
			savePrivateKey(keyPath, rsaKey)
			savePublicKey(pubKeyPath, &rsaKey.PublicKey)
		}
	})
}

func generateKey() (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, KeySize)
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("invalid private key format")
	}

	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func savePrivateKey(path string, key *rsa.PrivateKey) {
	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: keyBytes}
	data := pem.EncodeToMemory(block)
	os.WriteFile(path, data, 0600)
}

func savePublicKey(path string, key *rsa.PublicKey) {
	keyBytes, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		return
	}
	block := &pem.Block{Type: "PUBLIC KEY", Bytes: keyBytes}
	data := pem.EncodeToMemory(block)
	os.WriteFile(path, data, 0644)
}

func GetPublicKey() string {
	keyMu.RLock()
	defer keyMu.RUnlock()

	if rsaKey == nil {
		return ""
	}

	keyBytes, err := x509.MarshalPKIXPublicKey(&rsaKey.PublicKey)
	if err != nil {
		return ""
	}

	block := &pem.Block{Type: "PUBLIC KEY", Bytes: keyBytes}
	return base64.StdEncoding.EncodeToString(pem.EncodeToMemory(block))
}

func GetPublicKeyPEM() string {
	keyMu.RLock()
	defer keyMu.RUnlock()

	if rsaKey == nil {
		return ""
	}

	keyBytes, err := x509.MarshalPKIXPublicKey(&rsaKey.PublicKey)
	if err != nil {
		return ""
	}

	block := &pem.Block{Type: "PUBLIC KEY", Bytes: keyBytes}
	return string(pem.EncodeToMemory(block))
}

func Decrypt(ciphertextBase64 string) (string, error) {
	keyMu.RLock()
	defer keyMu.RUnlock()

	if rsaKey == nil {
		return "", errors.New("RSA key not initialized")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return "", err
	}

	plaintext, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		rsaKey,
		ciphertext,
		nil,
	)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

func DecryptBytes(ciphertext []byte) ([]byte, error) {
	keyMu.RLock()
	defer keyMu.RUnlock()

	if rsaKey == nil {
		return nil, errors.New("RSA key not initialized")
	}

	return rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		rsaKey,
		ciphertext,
		nil,
	)
}
