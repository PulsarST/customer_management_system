package routers

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		RegisterSiteRoutes(api)
	}

	registerStatic(r)
}

// resolveFrontendDir locates the frontend/ directory so the Go server can serve
// the chat widget directly. Honours FRONTEND_DIR, otherwise walks up from both
// the executable directory and the working directory looking for
// frontend/index.html. The upward walk makes it work no matter where the binary
// lives (backend/bin, backend/cmd via `go run`, repo root, ...).
func resolveFrontendDir() string {
	if p := os.Getenv("FRONTEND_DIR"); p != "" {
		return p
	}
	return findUp(startDirs(), "frontend", "index.html")
}

// startDirs returns the directories to begin an upward search from: the
// executable's directory and the current working directory.
func startDirs() []string {
	var dirs []string
	if exe, err := os.Executable(); err == nil {
		dirs = append(dirs, filepath.Dir(exe))
	}
	if wd, err := os.Getwd(); err == nil {
		dirs = append(dirs, wd)
	}
	return dirs
}

// findUp walks up from each start directory (up to a few levels) looking for
// <dir>/<rel>/<marker>. Returns the matching <dir>/<rel> path, or "" if none.
func findUp(starts []string, rel, marker string) string {
	const maxLevels = 6
	for _, start := range starts {
		dir := start
		for i := 0; i < maxLevels; i++ {
			candidate := filepath.Join(dir, rel)
			if _, err := os.Stat(filepath.Join(candidate, marker)); err == nil {
				return candidate
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return ""
}

// registerStatic serves the frontend widget from the same origin as the API,
// which avoids cross-origin / file:// issues when opening the chat. A NoRoute
// file server is used so the static catch-all does not conflict with the
// registered /api wildcard routes.
func registerStatic(r *gin.Engine) {
	dir := resolveFrontendDir()
	if dir == "" {
		return
	}

	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	fs := http.FileServer(http.Dir(dir))
	r.NoRoute(func(c *gin.Context) {
		// Never let the static server shadow the API namespace.
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		fs.ServeHTTP(c.Writer, c.Request)
	})
}
