package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var clients = make(map[*websocket.Conn]bool)
var clientsMu sync.Mutex

func handleWebSocket(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	clientsMu.Lock()
	clients[ws] = true
	clientsMu.Unlock()

	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			clientsMu.Lock()
			delete(clients, ws)
			clientsMu.Unlock()
			break
		}
	}
	return nil
}

func broadcastStatus(id int, status, errMsg string) {
	msg, _ := json.Marshal(map[string]interface{}{
		"type":    "status_update",
		"id":      id,
		"status":  status,
		"error":   errMsg,
	})
	
	clientsMu.Lock()
	for client := range clients {
		client.WriteMessage(websocket.TextMessage, msg)
	}
	clientsMu.Unlock()
}

type Repository struct {
	ID              int    `json:"id"`
	Name            string `json:"name"`
	URL             string `json:"url"`
	IntervalMinutes int    `json:"interval_minutes"`
	LastSync        string `json:"last_sync"`
	Status          string `json:"status"`
	LastCommit      string `json:"last_commit"`
	ErrorMessage    string `json:"error_message"`
	Stars           int    `json:"stars"`
	Forks           int    `json:"forks"`
	OpenIssues      int    `json:"open_issues"`
	CommitHistory   string `json:"commit_history"`
	HealthScore     int    `json:"health_score"`
	DefaultBranch   string `json:"default_branch"`
}

func getRepositories(c echo.Context) error {
	rows, err := db.Query("SELECT id, name, url, interval_minutes, last_sync, status, last_commit, error_message, stars, forks, open_issues, commit_history, health_score FROM repositories")
	if err != nil {
		return err
	}
	defer rows.Close()

	repos := []Repository{}
	for rows.Next() {
		var r Repository
		var lastSync sql.NullString
		rows.Scan(&r.ID, &r.Name, &r.URL, &r.IntervalMinutes, &lastSync, &r.Status, &r.LastCommit, &r.ErrorMessage, &r.Stars, &r.Forks, &r.OpenIssues, &r.CommitHistory, &r.HealthScore)
		r.LastSync = lastSync.String
		
		// Get default branch name from git
		repoPath := filepath.Join("./data", r.Name)
		cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
		out, _ := cmd.CombinedOutput()
		r.DefaultBranch = strings.TrimSpace(string(out))
		if r.DefaultBranch == "" { r.DefaultBranch = "main" }

		repos = append(repos, r)
	}
	return c.JSON(http.StatusOK, repos)
}

func addRepository(c echo.Context) error {
	var r Repository
	if err := c.Bind(&r); err != nil {
		return err
	}
	
	result, err := db.Exec("INSERT INTO repositories (name, url, interval_minutes) VALUES (?, ?, ?)", r.Name, r.URL, r.IntervalMinutes)
	if err != nil {
		return err
	}
	
	id, _ := result.LastInsertId()
	go syncRepo(int(id), r.URL, r.Name)
	
	return c.JSON(http.StatusCreated, map[string]int{"id": int(id)})
}

func deleteRepository(c echo.Context) error {
	id := c.Param("id")
	
	// Get name to delete folder
	var name string
	err := db.QueryRow("SELECT name FROM repositories WHERE id = ?", id).Scan(&name)
	if err != nil {
		return err
	}

	// Delete from DB
	_, err = db.Exec("DELETE FROM repositories WHERE id = ?", id)
	if err != nil {
		return err
	}

	// Delete folder
	os.RemoveAll("./data/" + name)
	
	return c.NoContent(http.StatusNoContent)
}

func getReadme(c echo.Context) error {
	id := c.Param("id")
	var name string
	err := db.QueryRow("SELECT name FROM repositories WHERE id = ?", id).Scan(&name)
	if err != nil {
		return err
	}

	repoPath := filepath.Join("./data", name)
	// Try common readme filenames
	filenames := []string{"README.md", "readme.md", "README.txt", "README"}
	for _, f := range filenames {
		content, err := os.ReadFile(filepath.Join(repoPath, f))
		if err == nil {
			return c.String(http.StatusOK, string(content))
		}
	}

	return c.String(http.StatusNotFound, "No README found")
}
