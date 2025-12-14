package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type ConsultingServer struct {
	router *gin.Engine
	db     *sql.DB
}

type Consultant struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Speciality  string  `json:"speciality"`
	HourlyRate  float64 `json:"hourly_rate"`
	Description string  `json:"description"`
}

func NewConsultingServer() *ConsultingServer {
	router := gin.Default()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/consulting_db?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return &ConsultingServer{
		router: router,
		db:     db,
	}
}

func (cs *ConsultingServer) SetupRoutes() {
	cs.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	cs.router.GET("/consultants", cs.handleGetConsultants)
	cs.router.POST("/consultants", cs.handleCreateConsultant)
	cs.router.GET("/consultants/:id", cs.handleGetConsultant)
}

func (cs *ConsultingServer) handleGetConsultants(c *gin.Context) {
	rows, err := cs.db.Query("SELECT id, name, speciality, hourly_rate, description FROM consultants")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch consultants"})
		return
	}
	defer rows.Close()

	consultants := []Consultant{}
	for rows.Next() {
		var consultant Consultant
		if err := rows.Scan(&consultant.ID, &consultant.Name, &consultant.Speciality, &consultant.HourlyRate, &consultant.Description); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan consultant"})
			return
		}
		consultants = append(consultants, consultant)
	}

	c.JSON(http.StatusOK, gin.H{"consultants": consultants})
}

func (cs *ConsultingServer) handleCreateConsultant(c *gin.Context) {
	var consultant Consultant
	if err := c.ShouldBindJSON(&consultant); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	consultant.ID = "consultant-" + os.Getenv("HOSTNAME")
	_, err := cs.db.Exec(
		"INSERT INTO consultants (id, name, speciality, hourly_rate, description) VALUES ($1, $2, $3, $4, $5)",
		consultant.ID, consultant.Name, consultant.Speciality, consultant.HourlyRate, consultant.Description,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create consultant"})
		return
	}

	c.JSON(http.StatusCreated, consultant)
}

func (cs *ConsultingServer) handleGetConsultant(c *gin.Context) {
	id := c.Param("id")
	var consultant Consultant

	err := cs.db.QueryRow(
		"SELECT id, name, speciality, hourly_rate, description FROM consultants WHERE id = $1",
		id,
	).Scan(&consultant.ID, &consultant.Name, &consultant.Speciality, &consultant.HourlyRate, &consultant.Description)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "consultant not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch consultant"})
		return
	}

	c.JSON(http.StatusOK, consultant)
}

func (cs *ConsultingServer) Start(port string) error {
	return cs.router.Run(":" + port)
}

func main() {
	godotenv.Load()

	port := os.Getenv("CONSULTING_PORT")
	if port == "" {
		port = "3002"
	}

	server := NewConsultingServer()
	server.SetupRoutes()

	log.Printf("Consulting server starting on port %s\n", port)
	if err := server.Start(port); err != nil {
		log.Fatalf("Failed to start consulting server: %v", err)
	}
}
