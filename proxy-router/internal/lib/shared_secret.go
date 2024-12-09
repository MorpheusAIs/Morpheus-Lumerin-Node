package lib

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
)

const (
	KeySize   = 32
	NonceSize = 12
)

func GenerateEphemeralKeyPair() (publicKey, privateKey []byte, err error) {
	privateKey = make([]byte, KeySize)
	_, err = rand.Read(privateKey)
	if err != nil {
		return nil, nil, err
	}

	publicKey, err = curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return nil, nil, err
	}

	return publicKey, privateKey, nil
}

func ComputeSharedSecret(privateKey, peerPublicKey []byte) ([]byte, error) {
	sharedSecret, err := curve25519.X25519(privateKey, peerPublicKey)
	if err != nil {
		return nil, err
	}
	return sharedSecret, nil
}

func DeriveKeysFromSharedSecret(sharedSecret []byte) (encryptionKey []byte, err error) {
	hash := sha256.New
	hkdf := hkdf.New(hash, sharedSecret, nil, nil)
	encryptionKey = make([]byte, KeySize)
	_, err = io.ReadFull(hkdf, encryptionKey)
	if err != nil {
		return nil, err
	}
	return encryptionKey, nil
}

func SharedSecretDecrypt(encryptionKey, encryptedData []byte) (dectyptedData []byte, err error) {
	aead, err := chacha20poly1305.New(encryptionKey)
	if err != nil {
		return nil, err
	}

	if len(encryptedData) < NonceSize {
		return nil, fmt.Errorf("encrypted prompt too short")
	}

	nonce := encryptedData[:NonceSize]
	ciphertext := encryptedData[NonceSize:]

	prompt, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return prompt, nil
}

func SharedSecretEncrypt(encryptionKey, data []byte) (encryptedData []byte, err error) {
	aead, err := chacha20poly1305.New(encryptionKey)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, NonceSize)
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, err
	}

	encryptedData = aead.Seal(nonce, nonce, data, nil)
	return encryptedData, nil
}
