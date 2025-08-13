/*
 * Copyright (C) 2025 Miguel Mamani <miguel.coder.per@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 */

package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2 parameters
	saltLength   = 32
	keyLength    = 32
	argon2Time   = 3
	memory       = 64 * 1024
	threads      = 4
)

// PasswordHasher handles password hashing and verification
type PasswordHasher struct{}

// NewPasswordHasher creates a new password hasher
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{}
}

// HashPassword hashes a password using Argon2id
func (ph *PasswordHasher) HashPassword(password string) (string, error) {
	salt, err := generateRandomSalt(saltLength)
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argon2Time, memory, threads, keyLength)

	// Encode salt and hash to base64
	saltBase64 := base64.RawStdEncoding.EncodeToString(salt)
	hashBase64 := base64.RawStdEncoding.EncodeToString(hash)

	// Format: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
	encodedHash := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		memory, argon2Time, threads, saltBase64, hashBase64)

	return encodedHash, nil
}

// VerifyPassword verifies a password against its hash
func (ph *PasswordHasher) VerifyPassword(password, hashedPassword string) bool {
	salt, hash, err := ph.decodeHash(hashedPassword)
	if err != nil {
		return false
	}

	otherHash := argon2.IDKey([]byte(password), salt, argon2Time, memory, threads, keyLength)

	return subtle.ConstantTimeCompare(hash, otherHash) == 1
}

// decodeHash decodes the hash string and extracts salt and hash
func (ph *PasswordHasher) decodeHash(encodedHash string) (salt, hash []byte, err error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, fmt.Errorf("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return nil, nil, fmt.Errorf("incompatible hash algorithm")
	}

	if parts[2] != "v=19" {
		return nil, nil, fmt.Errorf("incompatible argon2 version")
	}

	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode salt: %w", err)
	}

	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode hash: %w", err)
	}

	return salt, hash, nil
}

// generateRandomSalt generates a random salt of the specified length
func generateRandomSalt(length int) ([]byte, error) {
	salt := make([]byte, length)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// GenerateSecureToken generates a secure random token for password reset, etc.
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}