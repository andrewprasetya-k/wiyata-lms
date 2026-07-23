package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"os"
)

var ErrMFAEncryptionKeyMissing = errors.New("MFA_ENCRYPTION_KEY is not set to a valid 32-byte base64 key")

// mfaEncryptionKey reads MFA_ENCRYPTION_KEY once per call — a 32-byte
// AES-256 key, base64-encoded so it can hold arbitrary bytes safely in an
// env var (generate one with `openssl rand -base64 32`). Deliberately a
// separate key from JWT_SECRET: JWT_SECRET only ever signs/verifies and
// never needs to decrypt anything, while this key exists specifically to
// decrypt TOTP secrets back to plaintext for validation — mixing the two
// would mean rotating one silently affects the other for no reason.
func mfaEncryptionKey() ([]byte, error) {
	raw := os.Getenv("MFA_ENCRYPTION_KEY")
	if raw == "" {
		return nil, ErrMFAEncryptionKeyMissing
	}
	key, err := base64.StdEncoding.DecodeString(raw)
	if err != nil || len(key) != 32 {
		return nil, ErrMFAEncryptionKeyMissing
	}
	return key, nil
}

// encryptMFASecret encrypts plaintext (a TOTP secret) with AES-256-GCM,
// returning hex(nonce || ciphertext || tag) as a single string to persist.
func encryptMFASecret(plaintext string) (string, error) {
	key, err := mfaEncryptionKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

// decryptMFASecret reverses encryptMFASecret.
func decryptMFASecret(encoded string) (string, error) {
	key, err := mfaEncryptionKey()
	if err != nil {
		return "", err
	}

	ciphertext, err := hex.DecodeString(encoded)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return "", errors.New("mfa secret ciphertext is malformed")
	}

	nonce, sealed := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
