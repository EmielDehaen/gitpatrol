package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	JWTSecret      string
	PasswordPepper string
	DBPath         string
	WorkerCount    int
	GithubToken    string
	GitlabToken    string
	envPath        string
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

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./db/gitpatrol.db"
		os.Setenv("DB_PATH", dbPath)
		modified = true
	}

	workerCount, err := strconv.Atoi(os.Getenv("WORKERS"))
	if workerCount == 0 || err != nil {
		workerCount = 3
		os.Setenv("WORKERS", strconv.Itoa(workerCount))
		modified = true
	}

	github := os.Getenv("GITHUB_TOKEN")
	if github == "" {
		log.Printf("[CONFIG] Info: no github env token found. We'll use the public api. Go to settings to add one.")
	}

	gitlab := os.Getenv("GITLAB_TOKEN")
	if gitlab == "" {
		log.Printf("[CONFIG] Info: no gitlab env token found. Can't use gitlab. Go to settings to add one.")
	}

	if modified {
		_ = saveEnv(envPath, jwtSecret, pepper, dbPath, github, gitlab, workerCount)
	}

	return &Config{
		JWTSecret:      jwtSecret,
		PasswordPepper: pepper,
		DBPath:         dbPath,
		WorkerCount:    workerCount,
		GithubToken:    github,
		GitlabToken:    gitlab,
		envPath:        envPath,
	}
}

func (c *Config) UpdateTokens(github, gitlab string) error {
	if github != "" {
		c.GithubToken = github
		os.Setenv("GITHUB_TOKEN", github)
	}
	if gitlab != "" {
		c.GitlabToken = gitlab
		os.Setenv("GITLAB_TOKEN", gitlab)
	}

	err := saveEnv(c.envPath, c.JWTSecret, c.PasswordPepper, c.DBPath, c.GithubToken, c.GitlabToken, c.WorkerCount)

	return err
}

func saveEnv(path, jwt, pepper, dbPath, github, gitlab string, workerCount int) error {
	env := map[string]string{
		"JWT_SECRET":      jwt,
		"PASSWORD_PEPPER": pepper,
		"DB_PATH":         dbPath,
		"WORKERS":         strconv.Itoa(workerCount),
		"GITHUB_TOKEN":    github,
		"GITLAB_TOKEN":    gitlab,
	}

	err := godotenv.Write(env, path)
	if err != nil {
		log.Printf("[CONFIG] Warning: Could not save .env file: %v", err)
		return err
	} else {
		log.Printf("[CONFIG] Secrets generated and saved to %s", path)
		return nil
	}
}

func generateRandomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "efinity-default-fallback-secret-key"
	}
	return hex.EncodeToString(b)
}
