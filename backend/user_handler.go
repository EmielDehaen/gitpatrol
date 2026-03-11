package main

import (
	"fmt"
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

	hash, err := hashPassword(input.Password)
	if err != nil {
		fmt.Printf("[AUTH] Password hashing failed: %v\n", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Password too long."})
	}
	res, err := db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", input.Username, hash)
	if err != nil {
		fmt.Printf("[AUTH] User registration failed: %v\n", err)
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
	secure := os.Getenv("GP_SECURE_COOKIE") == "true"
	cookie := &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  time.Now().Add(72 * time.Hour),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	}
	c.SetCookie(cookie)
}

func updateUser(c echo.Context) error {
	userID := c.Get("user_id").(int)
	var input struct {
		Username    string `json:"username"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.Bind(&input); err != nil {
		return err
	}

	if input.NewPassword != "" {
		var currentHash string
		err := db.QueryRow("SELECT password_hash FROM users WHERE id = ?", userID).Scan(&currentHash)
		if err != nil || !checkPasswordHash(input.OldPassword, currentHash) {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Current password incorrect."})
		}
		newHash, _ := hashPassword(input.NewPassword)
		_, err = db.Exec("UPDATE users SET username = ?, password_hash = ? WHERE id = ?", input.Username, newHash, userID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Update failed (Username might exist)."})
		}
	} else {
		_, err := db.Exec("UPDATE users SET username = ? WHERE id = ?", input.Username, userID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Update failed (Username might exist)."})
		}
	}

	return c.JSON(http.StatusOK, map[string]string{"username": input.Username})
}

func getMe(c echo.Context) error {
	userID := c.Get("user_id").(int)
	var username string
	db.QueryRow("SELECT username FROM users WHERE id = ?", userID).Scan(&username)
	return c.JSON(http.StatusOK, map[string]string{"id": fmt.Sprintf("%d", userID), "username": username})
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
