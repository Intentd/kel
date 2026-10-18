package core

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"text/template"
)

// randomFuncs — Random
//
// Domain: random
//
// Generate cryptographically secure random values: hex, base64 and
// alphanumeric strings, and UUIDs. They return a different value on every
// call, so evaluate them once, when a secret is first created, and keep the
// result; evaluated on every reconcile, a password changes every time.
func randomFuncs() template.FuncMap {
	return template.FuncMap{
		// RandomAlphanumeric n — returns n random alphanumeric characters
		"randomAlphanumeric": RandomAlphanumeric,

		// RandomHex n — returns 2n hex characters (n random bytes)
		"randomHex": RandomHex,

		// RandomBase64 n — returns URL-safe base64 from n random bytes
		"randomBase64": RandomBase64,

		// Uuidv4 — returns a random UUID v4 string
		"uuidv4": Uuidv4,
	}
}

// charset for randomAlphanumeric
const alphanumCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandomAlphanumeric generates a cryptographically random alphanumeric string
// of exactly `n` characters. Characters are drawn from `[a-zA-Z0-9]`.
//
// Keywords: random, secret, password, alphanumeric, generate, crypto, credentials
//
// Example:
//
//	{{ randomAlphanumeric 12 }}  → "k7Xm3pQs9vR2"
func RandomAlphanumeric(n int) (string, error) {
	if n <= 0 {
		return "", nil
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("randomAlphanumeric: %w", err)
	}
	result := make([]byte, n)
	charsetLen := byte(len(alphanumCharset))
	for i := range result {
		result[i] = alphanumCharset[b[i]%charsetLen]
	}
	return string(result), nil
}

// RandomHex generates `n` random bytes and returns them as a hex-encoded
// string. The output is `2n` characters long.
//
// Keywords: random, secret, token, hex, generate, crypto, api-key, session
//
// Example:
//
//	{{ randomHex 8 }}  → "a3f8b2c1d4e5f6a7"
func RandomHex(n int) (string, error) {
	if n <= 0 {
		return "", nil
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("randomHex: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// Uuidv4 generates a random UUID v4 string in standard `8-4-4-4-12` hex
// format. Same entropy as `randomHex 16`, formatted as a UUID for systems that
// expect that shape.
//
// Keywords: random, uuid, secret, token, id, identifier, generate, crypto
//
// Example:
//
//	{{ uuidv4 }}  → "f47ac10b-58cc-4372-a567-0e02b2c3d479"
func Uuidv4() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("uuidv4: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

// RandomBase64 generates `n` random bytes and returns them as a URL-safe base64
// string. The output length is approximately `ceil(n * 4/3)` characters.
//
// Keywords: random, secret, jwt, base64, generate, crypto, signing, hmac, oauth
//
// Example:
//
//	{{ randomBase64 16 }}  → "47H1Wtj_DBL6gy0hBE-49w=="
func RandomBase64(n int) (string, error) {
	if n <= 0 {
		return "", nil
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("randomBase64: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
