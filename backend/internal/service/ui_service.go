package service

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// UI is the embedded frontend build
var UI embed.FS

type UIService struct {
	dist fs.FS
}

func NewUIService() *UIService {
	// Sub to the "build" folder if we embed it as "build"
	// For now, we assume the caller will provide the subfs if needed, 
	// but let's make it robust.
	f, _ := fs.Sub(UI, "build")
	if f == nil {
		f = UI
	}
	return &UIService{dist: f}
}

func (s *UIService) RegisterRoutes(e *echo.Echo) {
	// Serve static files from the embedded FS
	fileServer := http.FileServer(http.FS(s.dist))
	
	e.GET("/*", echo.WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the request is for an API or WebSocket, let Echo handle it (shouldn't happen with correct route order)
		if strings.HasPrefix(r.URL.Path, "/api") || strings.HasPrefix(r.URL.Path, "/ws") || strings.HasPrefix(r.URL.Path, "/avatars") {
			return
		}

		// Check if file exists in embedded FS
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		
		_, err := fs.Stat(s.dist, path)
		if err != nil {
			// If file not found, serve index.html (SPA routing)
			r.URL.Path = "/"
		}
		
		fileServer.ServeHTTP(w, r)
	})))
}
