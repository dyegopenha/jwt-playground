package jwtutil

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"github.com/dyegopenha/jwt-playground/internal/config/env"
	"github.com/golang-jwt/jwt/v5"
)

type JWTUtil struct {
	e *env.Env
}

func NewJWTUtil(e *env.Env) *JWTUtil {
	return &JWTUtil{
		e: e,
	}
}

// Claims represents the JWT payload used across the application.
// It embeds jwt.RegisteredClaims to take advantage of the built-in
// validations provided by the golang-jwt library.
type Claims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

func generateRandomBase64(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).
			EncodeToString(b),
		nil
}

// SignAccessToken creates and signs a short-lived access token.
func (j *JWTUtil) SignAccessToken(
	userID, role string,
	ttl time.Duration,
) (string, error) {
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    userID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return tok.SignedString([]byte(j.e.HMACKey))
}

// ParseAndVerify validates the signature and returns the Claims inside a token.
func (j *JWTUtil) ParseAndVerify(tokenStr string) (*Claims, error) {
	tok, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(t *jwt.Token) (any, error) { return []byte(j.e.HMACKey), nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return nil, err
	}
	claims, ok := tok.Claims.(*Claims)
	if !ok || !tok.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

// GenerateRefreshToken generates a refresh token
func (j *JWTUtil) GenerateRefreshToken() (string, error) {
	refreshToken, err := generateRandomBase64(32)
	if err != nil {
		return "", err
	}
	return refreshToken, nil
}

// IssueTokenPair returns an access token and a refresh token.
func (j *JWTUtil) IssueTokenPair(
	userID, role string,
	accessTTL, refreshTTL time.Duration,
) (string, string, error) {
	accessTok, err := j.SignAccessToken(userID, role, accessTTL)
	if err != nil {
		return "", "", err
	}
	refreshTok, err := j.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}
	return accessTok, refreshTok, nil
}
