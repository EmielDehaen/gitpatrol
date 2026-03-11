package main

import (
	"database/sql"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	initDB()
	defer db.Close()

	initSyncManager(3)

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
	e.GET("/api/repositories/:id/badge", getHealthBadge)
	e.GET("/api/incidents", getIncidents)
	e.DELETE("/api/incidents", clearIncidents)
	e.Static("/avatars", "./data/avatars")

	e.GET("/api/repositories/:id/assets/*", getAsset)

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
						manager.Enqueue(id, url, name)
					}
				}
				rows.Close()
			}
			time.Sleep(10 * time.Second)
		}
	}()

	e.Logger.Fatal(e.Start(":8080"))
}
