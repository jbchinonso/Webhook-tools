// Package hashes centralises all password/token hashing so the algorithm
// choice is visible in one place and swappable without touching modules.
package hashes

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// HashPassword returns a bcrypt hash of the plaintext password.
// Cost 12 is the current "safe enough, still fast" default.
func HashPassword(plain string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(h), nil
}

// ComparePassword returns nil if the plaintext matches the stored hash.
// bcrypt's own ErrMismatchedHashAndPassword is returned unchanged so callers
// can decide how to surface it.
func ComparePassword(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
}

// NewOpaqueToken generates n random bytes, base64url-encodes them, and
// returns (encoded, sha256Hex). The caller stores the hash and returns the
// encoded token to the client. The raw token is never persisted.
func NewOpaqueToken(n int) (token string, hash string, err error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("random: %w", err)
	}
	token = base64.RawURLEncoding.EncodeToString(buf)
	hash = HashToken(token)
	return token, hash, nil
}

// HashToken returns the hex-encoded SHA-256 of a token. SHA-256 (not bcrypt)
// is correct here: tokens are high-entropy (32 bytes of randomness), so they
// don't need a slow KDF.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// NewNumericCode returns a cryptographically random string of n digits,
// zero-padded. Uses crypto/rand, not math/rand.
func NewNumericCode(n int) (string, error) {
	if n <= 0 || n > 10 {
		return "", fmt.Errorf("invalid code length: %d", n)
	}
	max := int64(1)
	for i := 0; i < n; i++ {
		max *= 10
	}
	v, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", n, v.Int64()), nil
}