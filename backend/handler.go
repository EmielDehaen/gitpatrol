package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

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
	AutoPatrol      int    `json:"auto_patrol"`
}

func getRepositories(c echo.Context) error {
	rows, err := db.Query("SELECT id, name, url, interval_minutes, last_sync, status, last_commit, error_message, stars, forks, open_issues, commit_history, health_score, default_branch, auto_patrol FROM repositories")
	if err != nil {
		return err
	}
	defer rows.Close()

	repos := []Repository{}
	for rows.Next() {
		var r Repository
		var lastSync, lastCommit, errMsg, commitHistory sql.NullString
		err := rows.Scan(&r.ID, &r.Name, &r.URL, &r.IntervalMinutes, &lastSync, &r.Status, &lastCommit, &errMsg, &r.Stars, &r.Forks, &r.OpenIssues, &commitHistory, &r.HealthScore, &r.DefaultBranch, &r.AutoPatrol)
		if err != nil {
			return err
		}
		r.LastSync = lastSync.String
		r.LastCommit = lastCommit.String
		r.ErrorMessage = errMsg.String
		r.CommitHistory = commitHistory.String
		repos = append(repos, r)
	}
	return c.JSON(http.StatusOK, repos)
}

func addRepository(c echo.Context) error {
	var r Repository
	if err := c.Bind(&r); err != nil {
		return err
	}

	r.URL = NormalizeURL(r.URL)
	if !strings.Contains(r.URL, "http") && !strings.Contains(r.URL, "@") {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid Repository URL. Please provide a valid Git URL."})
	}

	res, err := db.Exec("INSERT INTO repositories (name, url, interval_minutes, status, auto_patrol) VALUES (?, ?, ?, ?, ?)", r.Name, r.URL, r.IntervalMinutes, "pending", r.AutoPatrol)
	if err != nil {
		return err
	}

	id, _ := res.LastInsertId()
	manager.Enqueue(int(id), r.URL, r.Name)
	
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

func updateRepository(c echo.Context) error {
	id := c.Param("id")
	var r Repository
	if err := c.Bind(&r); err != nil {
		return err
	}

	_, err := db.Exec("UPDATE repositories SET interval_minutes = ?, auto_patrol = ? WHERE id = ?", r.IntervalMinutes, r.AutoPatrol, id)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func syncRepositoryNow(c echo.Context) error {
	id := c.Param("id")
	var name, url string
	err := db.QueryRow("SELECT name, url FROM repositories WHERE id = ?", id).Scan(&name, &url)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Repository not found"})
	}

	// Trigger async sync via manager
	var repoID int
	fmt.Sscanf(id, "%d", &repoID)
	manager.Enqueue(repoID, url, name)

	return c.JSON(http.StatusOK, map[string]string{"status": "Syncing started"})
}

func getIncidents(c echo.Context) error {
	rows, err := db.Query("SELECT id, repo_id, repo_name, message, created_at FROM incidents WHERE resolved = 0 ORDER BY created_at DESC")
	if err != nil { return err }
	defer rows.Close()

	type Incident struct {
		ID        int       `json:"id"`
		RepoID    int       `json:"repo_id"`
		RepoName  string    `json:"repo_name"`
		Message   string    `json:"message"`
		CreatedAt time.Time `json:"created_at"`
	}

	incidents := []Incident{}
	for rows.Next() {
		var i Incident
		rows.Scan(&i.ID, &i.RepoID, &i.RepoName, &i.Message, &i.CreatedAt)
		incidents = append(incidents, i)
	}
	return c.JSON(http.StatusOK, incidents)
}

func clearIncidents(c echo.Context) error {
	_, err := db.Exec("UPDATE incidents SET resolved = 1")
	if err != nil { return err }
	return c.NoContent(http.StatusNoContent)
}

func getHealthBadge(c echo.Context) error {
	id := c.Param("id")
	var name string
	var score int
	err := db.QueryRow("SELECT name, health_score FROM repositories WHERE id = ?", id).Scan(&name, &score)
	if err != nil {
		return c.String(http.StatusNotFound, "Not found")
	}

	color := "#ff4d4d" // Red
	if score > 70 {
		color = "#00ff88" // Green
	} else if score > 40 {
		color = "#ffcc00" // Yellow
	}

	svg := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg width="120" height="20" version="1.1" xmlns="http://www.w3.org/2000/svg">
  <linearGradient id="a" x2="0" y2="100%%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <rect rx="3" width="120" height="20" fill="#555"/>
  <rect rx="3" x="70" width="50" height="20" fill="%s"/>
  <path fill="%s" d="M70 0h4v20H70z"/>
  <rect rx="3" width="120" height="20" fill="url(#a)"/>
  <g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11">
    <text x="35" y="15" fill="#010101" fill-opacity=".3">GitPatrol</text>
    <text x="35" y="14">GitPatrol</text>
    <text x="95" y="15" fill="#010101" fill-opacity=".3">%d%%</text>
    <text x="95" y="14">%d%%</text>
  </g>
</svg>`, color, color, score, score)

	return c.Blob(http.StatusOK, "image/svg+xml", []byte(svg))
}

func getAsset(c echo.Context) error {
	id := c.Param("id")
	var name, defaultBranch string
	err := db.QueryRow("SELECT name, default_branch FROM repositories WHERE id = ?", id).Scan(&name, &defaultBranch)
	if err != nil {
		return err
	}

	filePath := c.Param("*")

	// 1. If it's metadata (Issues/Releases JSON), serve from disk
	if strings.HasPrefix(filePath, "metadata/") {
		return c.File(filepath.Join("./data", name, filePath))
	}

	// 2. If it's Wiki, serve from the wiki git db (usually master/main in wiki repos)
	if strings.HasPrefix(filePath, "wiki/") {
		wikiPath := filepath.Join("./data", name, "wiki")
		wikiFile := strings.TrimPrefix(filePath, "wiki/")
		// Wikis are cloned as normal repos, so we can use HEAD or origin/master
		return serveGitFile(c, wikiPath, "HEAD", wikiFile)
	}

	// 3. Otherwise, serve from main git db using the remote tracking branch
	repoPath := filepath.Join("./data", name)
	return serveGitFile(c, repoPath, "origin/"+defaultBranch, filePath)
}

func serveGitFile(c echo.Context, repoPath, ref, filePath string) error {
	cmd := exec.Command("git", "-C", repoPath, "show", ref+":"+filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback to local HEAD if remote ref fails
		cmd = exec.Command("git", "-C", repoPath, "show", "HEAD:"+filePath)
		output, err = cmd.CombinedOutput()
		if err != nil {
			// Final fallback: check if it's physically on disk
			return c.File(filepath.Join(repoPath, filePath))
		}
	}

	// Detect mime type for correct rendering
	ext := filepath.Ext(filePath)
	mime := "text/plain"
	switch strings.ToLower(ext) {
	case ".png":
		mime = "image/png"
	case ".jpg", ".jpeg":
		mime = "image/jpeg"
	case ".gif":
		mime = "image/gif"
	case ".svg":
		mime = "image/svg+xml"
	case ".md":
		mime = "text/markdown"
	case ".json":
		mime = "application/json"
	}

	return c.Blob(http.StatusOK, mime, output)
}

func getReadme(c echo.Context) error {
	id := c.Param("id")
	var name, defaultBranch string
	err := db.QueryRow("SELECT name, default_branch FROM repositories WHERE id = ?", id).Scan(&name, &defaultBranch)
	if err != nil {
		return err
	}

	repoPath := filepath.Join("./data", name)
	// Try common readme filenames using git show
	filenames := []string{"README.md", "readme.md", "README.txt", "README"}
	
	// Try to get the latest from the remote tracking branch (fetched)
	ref := "origin/" + defaultBranch
	
	for _, f := range filenames {
		cmd := exec.Command("git", "-C", repoPath, "show", ref+":"+f)
		output, err := cmd.CombinedOutput()
		if err == nil {
			return c.String(http.StatusOK, string(output))
		}
	}

	// Fallback to local HEAD if remote ref fails
	for _, f := range filenames {
		cmd := exec.Command("git", "-C", repoPath, "show", "HEAD:"+f)
		output, err := cmd.CombinedOutput()
		if err == nil {
			return c.String(http.StatusOK, string(output))
		}
	}

	return c.String(http.StatusNotFound, "No README found in Git database")
}
