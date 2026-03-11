package main

import (
	"database/sql"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	initEnv()
	initDB()
	defer db.Close()

	initSyncManager(3)

	os.MkdirAll("./data", 0755)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowCredentials: true,
	}))

	// Public Auth Routes
	e.GET("/api/auth/status", checkAuthStatus)
	e.POST("/api/auth/register", register)
	e.POST("/api/auth/login", login)
	e.POST("/api/auth/logout", logout)

	// Protected API Group
	api := e.Group("/api", AuthMiddleware)
	api.GET("/repositories", getRepositories)
	api.POST("/repositories", addRepository)
	api.PATCH("/repositories/:id", updateRepository)
	api.DELETE("/repositories/:id", deleteRepository)
	api.POST("/repositories/:id/sync", syncRepositoryNow)
	api.GET("/repositories/:id/readme", getReadme)
	api.GET("/repositories/:id/badge", getHealthBadge)
	api.GET("/repositories/:id/assets/*", getAsset)
	api.GET("/incidents", getIncidents)
	api.DELETE("/incidents", clearIncidents)

	e.GET("/ws", handleWebSocket)
	e.Static("/avatars", "./data/avatars")

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
			time.Sleep(1 * time.Minute)
		}
	}()

	e.Logger.Fatal(e.Start(":8080"))
}
