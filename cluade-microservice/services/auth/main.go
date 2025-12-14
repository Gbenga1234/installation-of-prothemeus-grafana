package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type AuthServer struct {
	router *gin.Engine
	db     *sql.DB
}

type User struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

func NewAuthServer() *AuthServer {
	router := gin.Default()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/consulting_db?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return &AuthServer{
		router: router,
		db:     db,
	}
}

func (as *AuthServer) SetupRoutes() {
	as.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	as.router.POST("/register", as.handleRegister)
	as.router.POST("/login", as.handleLogin)
	as.router.GET("/validate", as.handleValidate)
}

func (as *AuthServer) handleRegister(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Simple registration - in production, hash password properly
	user := User{
		ID:       "user-" + time.Now().Unix().String(),
		Email:    req.Email,
		Password: req.Password,
	}

	_, err := as.db.Exec(
		"INSERT INTO users (id, email, password) VALUES ($1, $2, $3)",
		user.ID, user.Email, user.Password,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, _ := token.SignedString([]byte("your-secret-key-change-in-production"))
	c.JSON(http.StatusCreated, TokenResponse{Token: tokenString})
}

func (as *AuthServer) handleLogin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user User
	err := as.db.QueryRow(
		"SELECT id, email FROM users WHERE email = $1 AND password = $2",
		req.Email, req.Password,
	).Scan(&user.ID, &user.Email)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, _ := token.SignedString([]byte("your-secret-key-change-in-production"))
	c.JSON(http.StatusOK, TokenResponse{Token: tokenString})
}

func (as *AuthServer) handleValidate(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "valid"})
}

func (as *AuthServer) Start(port string) error {
	return as.router.Run(":" + port)
}

func main() {
	godotenv.Load()

	port := os.Getenv("AUTH_PORT")
	if port == "" {
		port = "3001"
	}

	server := NewAuthServer()
	server.SetupRoutes()

	log.Printf("Auth server starting on port %s\n", port)
	if err := server.Start(port); err != nil {
		log.Fatalf("Failed to start auth server: %v", err)
	}
}
