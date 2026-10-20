// Package auth implements JWT handling, password hashing and the auth middleware.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenService issues and verifies JWTs compatible with the former Spring
// backend (jjwt): HMAC-SHA with the algorithm chosen by secret length,
// "sub" = e-mail, "userId" = numeric user id, "iat" and "exp".
type TokenService struct {
	secret     []byte
	method     jwt.SigningMethod
	validAlgs  []string
	expiration time.Duration
	now        func() time.Time
}

type claims struct {
	UserID *int64 `json:"userId"`
	jwt.RegisteredClaims
}

func NewTokenService(secret []byte, expiration time.Duration) (*TokenService, error) {
	bits := len(secret) * 8
	if bits < 256 {
		return nil, errors.New("JWT secret must be at least 256 bits (32 bytes)")
	}
	// Same selection jjwt's Keys.hmacShaKeyFor + signWith(key) performs.
	method := jwt.SigningMethod(jwt.SigningMethodHS256)
	valid := []string{"HS256"}
	if bits >= 384 {
		method = jwt.SigningMethodHS384
		valid = append(valid, "HS384")
	}
	if bits >= 512 {
		method = jwt.SigningMethodHS512
		valid = append(valid, "HS512")
	}
	return &TokenService{
		secret:     secret,
		method:     method,
		validAlgs:  valid,
		expiration: expiration,
		now:        time.Now,
	}, nil
}

// Algorithm returns the signing algorithm in use (e.g. "HS384").
func (s *TokenService) Algorithm() string { return s.method.Alg() }

func (s *TokenService) Generate(userID int64, email string) (string, error) {
	now := s.now()
	c := claims{
		UserID: &userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   email,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.expiration)),
		},
	}
	return jwt.NewWithClaims(s.method, c).SignedString(s.secret)
}

// Parse verifies the token and returns the user id it was issued for.
func (s *TokenService) Parse(token string) (int64, error) {
	var c claims
	_, err := jwt.ParseWithClaims(token, &c,
		func(*jwt.Token) (any, error) { return s.secret, nil },
		jwt.WithValidMethods(s.validAlgs),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(s.now),
	)
	if err != nil {
		return 0, fmt.Errorf("invalid token: %w", err)
	}
	if c.UserID == nil {
		return 0, errors.New("invalid token: missing userId claim")
	}
	return *c.UserID, nil
}
