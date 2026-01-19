package jwt

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

func (j *JwtManager) NewJwt(userID uuid.UUID, role models.UserRole) (string, error) {
	const op = "pkg.jwt.NewJwt"

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = userID.String()
	claims["exp"] = time.Now().Add(j.jwtTtl).Unix()
	claims["role"] = string(role)

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
