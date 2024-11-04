package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAccessToken(t *testing.T) {
	secretKey := "test-secret-key"
	tokenService := NewTokenUseCase(secretKey)

	claims := JwtCustomClaims{
		Username: "testuser",
		Role:     "user",
		FullName: "Test User",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // Token expired after 1 day
			Issuer:    "go-todo-app",
		},
	}

	// Memanggil fungsi GenerateAccessToken
	token, err := tokenService.GenerateAccessToken(claims)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verifikasi bahwa token dapat di-parse dengan benar
	parsedToken, err := jwt.ParseWithClaims(token, &JwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	assert.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	// Memastikan klaim yang dihasilkan benar
	parsedClaims, ok := parsedToken.Claims.(*JwtCustomClaims)
	assert.True(t, ok)
	assert.Equal(t, claims.Username, parsedClaims.Username)
	assert.Equal(t, claims.Role, parsedClaims.Role)
	assert.Equal(t, claims.FullName, parsedClaims.FullName)
	assert.Equal(t, claims.Issuer, parsedClaims.Issuer)
}
