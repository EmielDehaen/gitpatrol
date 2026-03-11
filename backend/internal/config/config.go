package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTSecret      string
	PasswordPepper string
	DBPath         string
	WorkerCount    int
}

func LoadConfig(envPath string) *Config {
	_ = godotenv.Load(envPath)

	modified := false

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = generateRandomString(16)
		os.Setenv("JWT_SECRET", jwtSecret)
		modified = true
	}

	pepper := os.Getenv("PASSWORD_PEPPER")
	if pepper == "" {
		pepper = generateRandomString(16)
		os.Setenv("PASSWORD_PEPPER", pepper)
		modified = true
	}

	if modified {
		saveEnv(envPath, jwtSecret, pepper)
	}

	return &Config{
		JWTSecret:      jwtSecret,
		PasswordPepper: pepper,
		DBPath:         "./db/gitpatrol.db",
		WorkerCount:    3,
	}
}

func saveEnv(path, jwt, pepper string) {
	env := map[string]string{
		"JWT_SECRET":      jwt,
		"PASSWORD_PEPPER": pepper,
	}

	err := godotenv.Write(env, path)
	if err != nil {
		log.Printf("[CONFIG] Warning: Could not save .env file: %v", err)
	} else {
		log.Printf("[CONFIG] Secrets generated and saved to %s", path)
	}
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "efinity-default-fallback-secret-key"
	}
	return hex.EncodeToString(b)
}
