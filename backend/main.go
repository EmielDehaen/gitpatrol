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

	os.MkdirAll("./data", 0755)

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.GET("/ws", handleWebSocket)
	e.GET("/api/repositories", getRepositories)
	e.POST("/api/repositories", addRepository)

	// Scheduler
	go func() {
		for {
			rows, _ := db.Query("SELECT id, name, url, interval_minutes, last_sync FROM repositories")
			for rows.Next() {
				var id int
				var name, url string
				var interval int
				var lastSync sql.NullString
				rows.Scan(&id, &name, &url, &interval, &lastSync)
				
				shouldSync := true
				if lastSync.Valid {
					t, _ := time.Parse(time.RFC3339, lastSync.String)
					if time.Since(t) < time.Duration(interval)*time.Minute {
						shouldSync = false
					}
				}
				
				if shouldSync {
					go syncRepo(id, url, name)
				}
			}
			rows.Close()
			time.Sleep(1 * time.Minute)
		}
	}()

	e.Logger.Fatal(e.Start(":8080"))
}
