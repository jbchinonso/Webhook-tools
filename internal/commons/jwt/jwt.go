// Package jwt mints and verifies the short-lived access tokens used across
// the API. Refresh tokens are NOT JWTs — they are opaque strings managed by
// the sessions table.
package jwt

import (
	"errors"
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

// Claims is the payload we put in access tokens. It is deliberately small.
type Claims struct {
	UserID    string `json:"uid"`
	Role      string `json:"role"`
	SessionID string `json:"sid"`
	gojwt.RegisteredClaims
}

type Issuer struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

func NewIssuer(secret, issuer string, ttl time.Duration) *Issuer {
	return &Issuer{secret: []byte(secret), issuer: issuer, ttl: ttl}
}

// Issue creates a signed access token. now is injected for testability.
func (i *Issuer) Issue(userID, role, sessionID string, now time.Time) (string, time.Time, error) {
	exp := now.Add(i.ttl)
	claims := Claims{
		UserID:    userID,
		Role:      role,
		SessionID: sessionID,
		RegisteredClaims: gojwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    i.issuer,
			IssuedAt:  gojwt.NewNumericDate(now),
			ExpiresAt: gojwt.NewNumericDate(exp),
			NotBefore: gojwt.NewNumericDate(now),
		},
	}
	t := gojwt.NewWithClaims(gojwt.SigningMethodHS256, claims)
	signed, err := t.SignedString(i.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign jwt: %w", err)
	}
	return signed, exp, nil
}

// Verify parses and validates the token, returning the claims.
// All errors are wrapped so the caller can distinguish expired from invalid.
var (
	ErrExpired = errors.New("jwt: token expired")
	ErrInvalid = errors.New("jwt: token invalid")
)

func (i *Issuer) Verify(token string) (*Claims, error) {
	parsed, err := gojwt.ParseWithClaims(token, &Claims{}, func(t *gojwt.Token) (any, error) {
		if _, ok := t.Method.(*gojwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return i.secret, nil
	}, gojwt.WithIssuer(i.issuer), gojwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		if errors.Is(err, gojwt.ErrTokenExpired) {
			return nil, ErrExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalid, err)
	}
	claims, ok := parsed.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalid
	}
	return claims, nil
}