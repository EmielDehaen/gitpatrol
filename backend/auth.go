package main

import (
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func getJWTKey() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

func getPepper() string {
	return os.Getenv("PASSWORD_PEPPER")
}

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func hashPassword(password string) (string, error) {
	peppered := password + getPepper()
	bytes, err := bcrypt.GenerateFromPassword([]byte(peppered), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	peppered := password + getPepper()
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(peppered))
	return err == nil
}

func generateToken(id int, username string) (string, error) {
	expirationTime := time.Now().Add(72 * time.Hour)
	claims := &Claims{
		UserID:   id,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTKey())
}

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("session_token")
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Unauthorized"})
		}

		tokenString := cookie.Value
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return getJWTKey(), nil
		})

		if err != nil || !token.Valid {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid session"})
		}

		c.Set("user_id", claims.UserID)
		return next(c)
	}
}
