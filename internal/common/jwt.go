package common

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTMaker struct {
	Secret   []byte
	Issuer   string
	Audience string
}

type Claims struct {
	UserID string `json:"uid"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func (j *JWTMaker) GenerateAccess(userID, email string, ttl time.Duration) (token string, jti string, exp time.Time, err error) {
	now := time.Now()
	exp = now.Add(ttl)
	jti = uuid.NewString()

	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.Issuer,
			Audience:  []string{j.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
			ID:        jti, // <-- JTI
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = t.SignedString(j.Secret)
	return
}

func (j *JWTMaker) Parse(token string) (*Claims, error) {
	parsed, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (any, error) {
		return j.Secret, nil
	}, jwt.WithAudience(j.Audience), jwt.WithIssuer(j.Issuer))
	if err != nil {
		return nil, err
	}
	if c, ok := parsed.Claims.(*Claims); ok && parsed.Valid {
		return c, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}
