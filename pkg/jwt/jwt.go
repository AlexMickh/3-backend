package jwt

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtManager struct {
	secret string
	jwtTtl time.Duration
}

func New(secret string, jwtTtl time.Duration) *JwtManager {
	return &JwtManager{
		secret: secret,
		jwtTtl: jwtTtl,
	}
}

func (j *JwtManager) NewJwt(userID int64) (string, error) {
	const op = "pkg.jwt.NewJwt"

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = userID
	claims["exp"] = time.Now().Add(j.jwtTtl).Unix()

	tokenString, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return tokenString, nil
}

func (j *JwtManager) NewRefresh() (string, error) {
	const op = "pkg.jwt.NewRefresh"

	b := make([]byte, 32)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return fmt.Sprintf("%x", b), nil
}

func (j *JwtManager) Validate(token string) (int64, error) {
	const op = "pkg.jwt.Validate"

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		if t.Method == jwt.SigningMethodHS256 {
			return []byte(j.secret), nil
		}

		return struct{}{}, errors.New("invalid token")
	})
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	exp, err := claims.GetExpirationTime()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	if exp.Compare(time.Now()) == 1 {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, ok := claims["sub"].(int64)
	if !ok {
		return 0, fmt.Errorf("%s: failed to get id", op)
	}

	return id, nil
}
