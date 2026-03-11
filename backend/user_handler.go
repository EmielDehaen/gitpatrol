package main

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func register(c echo.Context) error {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.Bind(&input); err != nil {
		return err
	}

	// Community Edition: Only allow one user
	var userCount int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if userCount > 0 {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "Registration is closed (Community Edition: Single User Only)."})
	}

	hash, _ := hashPassword(input.Password)
	res, err := db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", input.Username, hash)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Username already exists."})
	}

	id, _ := res.LastInsertId()
	token, _ := generateToken(int(id), input.Username)
	setAuthCookie(c, token)

	return c.JSON(http.StatusCreated, map[string]string{"username": input.Username})
}

func login(c echo.Context) error {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.Bind(&input); err != nil {
		return err
	}

	var id int
	var hash string
	err := db.QueryRow("SELECT id, password_hash FROM users WHERE username = ?", input.Username).Scan(&id, &hash)
	if err != nil || !checkPasswordHash(input.Password, hash) {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid credentials."})
	}

	token, _ := generateToken(id, input.Username)
	setAuthCookie(c, token)

	return c.JSON(http.StatusOK, map[string]string{"username": input.Username})
}

func logout(c echo.Context) error {
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	}
	c.SetCookie(cookie)
	return c.NoContent(http.StatusNoContent)
}

func setAuthCookie(c echo.Context, token string) {
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  time.Now().Add(72 * time.Hour),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		// Secure: false by default for local Community setups without HTTPS
	}
	c.SetCookie(cookie)
}

func checkAuthStatus(c echo.Context) error {
	var userCount int
	db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	
	loggedIn := false
	cookie, err := c.Cookie("session_token")
	if err == nil && cookie.Value != "" {
		loggedIn = true
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"needs_bootstrap": userCount == 0,
		"logged_in":       loggedIn,
	})
}
