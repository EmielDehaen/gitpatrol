package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	initDB()
	defer db.Close()

	os.MkdirAll("./data", 0755)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/ws", handleWebSocket)
	e.GET("/api/repositories", getRepositories)
	e.POST("/api/repositories", addRepository)
	e.PATCH("/api/repositories/:id", updateRepository)
	e.DELETE("/api/repositories/:id", deleteRepository)
	e.POST("/api/repositories/:id/sync", syncRepositoryNow)
	e.GET("/api/repositories/:id/readme", getReadme)
	e.Static("/avatars", "./data/avatars")

	// Dynamic static assets from repositories
	e.GET("/api/repositories/:id/assets/*", func(c echo.Context) error {
		id := c.Param("id")
		var name string
		db.QueryRow("SELECT name FROM repositories WHERE id = ?", id).Scan(&name)
		
		filePath := c.Param("*")
		return c.File(filepath.Join("./data", name, filePath))
	})

	// Scheduler
	go func() {
		for {
			// Only pick repos that are NOT already syncing and have auto_patrol enabled
			rows, _ := db.Query("SELECT id, name, url, interval_minutes, last_sync FROM repositories WHERE auto_patrol = 1 AND status != 'syncing'")
			if rows != nil {
				for rows.Next() {
					var id int
					var name, url string
					var interval int
					var lastSync sql.NullTime // Use NullTime for cleaner SQLite integration
					
					if err := rows.Scan(&id, &name, &url, &interval, &lastSync); err != nil {
						continue
					}
					
					shouldSync := true
					if lastSync.Valid {
						if time.Since(lastSync.Time) < time.Duration(interval)*time.Minute {
							shouldSync = false
						}
					}
					
					if shouldSync {
						go syncRepo(id, url, name)
					}
				}
				rows.Close()
			}
			time.Sleep(1 * time.Minute)
		}
	}()

	e.Logger.Fatal(e.Start(":8080"))
}
