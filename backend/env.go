package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"

	"github.com/joho/godotenv"
)

const EnvFilePath = "./db/gitpatrol.env"

func initEnv() {
	// Ensure directory exists before loading/saving
	os.MkdirAll("./db", 0755)

	// Try to load existing env
	_ = godotenv.Load(EnvFilePath)

	modified := false
	
	if os.Getenv("JWT_SECRET") == "" {
		secret := generateRandomString(16) // 32 hex chars
		os.Setenv("JWT_SECRET", secret)
		modified = true
	}

	if os.Getenv("PASSWORD_PEPPER") == "" {
		pepper := generateRandomString(16) // 32 hex chars
		os.Setenv("PASSWORD_PEPPER", pepper)
		modified = true
	}

	if modified {
		saveEnv()
	}
}

func saveEnv() {
	env := map[string]string{
		"JWT_SECRET":       os.Getenv("JWT_SECRET"),
		"PASSWORD_PEPPER": os.Getenv("PASSWORD_PEPPER"),
	}

	// Also include MASTER_KEY for future use if it exists
	if mk := os.Getenv("MASTER_KEY"); mk != "" {
		env["MASTER_KEY"] = mk
	} else {
		env["MASTER_KEY"] = generateRandomString(16)
		os.Setenv("MASTER_KEY", env["MASTER_KEY"])
	}

	err := godotenv.Write(env, EnvFilePath)
	if err != nil {
		log.Printf("[ENV] Warning: Could not save .env file: %v", err)
	} else {
		log.Printf("[ENV] Secrets generated and saved to %s", EnvFilePath)
	}
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "efinity-default-fallback-secret-key"
	}
	return hex.EncodeToString(b)
}
