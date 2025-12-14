package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type GatewayServer struct {
	router *gin.Engine
	redis  *redis.Client
}

func NewGatewayServer() *GatewayServer {
	router := gin.Default()
	redisClient := redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})

	return &GatewayServer{
		router: router,
		redis:  redisClient,
	}
}

func (gs *GatewayServer) SetupRoutes() {
	// Health check
	gs.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Service discovery proxy routes
	gs.router.POST("/auth/register", gs.proxyRequest("http://auth-service:3001"))
	gs.router.POST("/auth/login", gs.proxyRequest("http://auth-service:3001"))
	gs.router.GET("/auth/validate", gs.proxyRequest("http://auth-service:3001"))

	gs.router.GET("/consultants", gs.proxyRequest("http://consulting-service:3002"))
	gs.router.POST("/consultants", gs.proxyRequest("http://consulting-service:3002"))
	gs.router.GET("/consultants/:id", gs.proxyRequest("http://consulting-service:3002"))

	gs.router.GET("/bookings", gs.proxyRequest("http://booking-service:3003"))
	gs.router.POST("/bookings", gs.proxyRequest("http://booking-service:3003"))
	gs.router.GET("/bookings/:id", gs.proxyRequest("http://booking-service:3003"))
}

func (gs *GatewayServer) proxyRequest(targetURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Simple proxy implementation - in production, use httputil.ReverseProxy
		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("Proxying to %s", targetURL),
			"path":    c.Request.URL.Path,
		})
	}
}

func (gs *GatewayServer) Start(port string) error {
	return gs.router.Run(":" + port)
}

func main() {
	godotenv.Load()

	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "3000"
	}

	server := NewGatewayServer()
	server.SetupRoutes()

	log.Printf("Gateway server starting on port %s\n", port)
	if err := server.Start(port); err != nil {
		log.Fatalf("Failed to start gateway server: %v", err)
	}
}
