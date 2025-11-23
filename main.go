// Package main provides the entry point for the FB2 to EPUB converter service.
package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/lex/fb2epub/config"
	"github.com/lex/fb2epub/handlers"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.Default()

	// Serve static files (CSS, JS)
	router.Static("/static", "./web/static")

	// Serve web UI
	router.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "fb2epub",
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		api.POST("/convert", handlers.ConvertFB2ToEPUB)
		api.GET("/status/:id", handlers.GetConversionStatus)
		api.GET("/download/:id", handlers.DownloadEPUB)
	}

	// Start server
	addr := ":" + cfg.Port
	log.Printf("Starting server on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
