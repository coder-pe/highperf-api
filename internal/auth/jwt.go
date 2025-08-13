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
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"highperf-api/internal/config"
	"highperf-api/internal/errors"
)

// Claims represents JWT claims
type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// JWTService handles JWT operations
type JWTService struct {
	secret        []byte
	tokenExpiry   time.Duration
	refreshExpiry time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(cfg config.AuthConfig) *JWTService {
	return &JWTService{
		secret:        []byte(cfg.JWTSecret),
		tokenExpiry:   cfg.TokenExpiry,
		refreshExpiry: cfg.RefreshExpiry,
	}
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// GenerateTokenPair generates access and refresh tokens
func (j *JWTService) GenerateTokenPair(userID int64, email string) (*TokenPair, error) {
	now := time.Now()

	// Generate access token
	accessClaims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "highperf-api",
			Subject:   fmt.Sprintf("user:%d", userID),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(j.secret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "highperf-api",
			Subject:   fmt.Sprintf("refresh:%d", userID),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(j.secret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    now.Add(j.tokenExpiry),
		TokenType:    "Bearer",
	}, nil
}

// ValidateToken validates a JWT token and returns claims
func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.ErrUnauthorized.WithMessage("Invalid token")
	}

	return claims, nil
}

// RefreshToken generates a new access token from a refresh token
func (j *JWTService) RefreshToken(refreshTokenString string) (*TokenPair, error) {
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, errors.ErrUnauthorized.WithMessage("Invalid refresh token")
	}

	// Check if this is actually a refresh token
	if claims.Subject != fmt.Sprintf("refresh:%d", claims.UserID) {
		return nil, errors.ErrUnauthorized.WithMessage("Not a refresh token")
	}

	// Generate new token pair
	return j.GenerateTokenPair(claims.UserID, claims.Email)
}

// ExtractTokenFromBearer extracts token from "Bearer <token>" format
func ExtractTokenFromBearer(authHeader string) (string, error) {
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.ErrUnauthorized.WithMessage("Invalid authorization header format")
	}
	return authHeader[len(bearerPrefix):], nil
}