package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type BookingServer struct {
	router *gin.Engine
	db     *sql.DB
}

type Booking struct {
	ID           string    `json:"id"`
	ConsultantID string    `json:"consultant_id"`
	UserID       string    `json:"user_id"`
	ScheduledAt  time.Time `json:"scheduled_at"`
	Duration     int       `json:"duration"`
	Status       string    `json:"status"`
}

type BookingRequest struct {
	ConsultantID string    `json:"consultant_id" binding:"required"`
	UserID       string    `json:"user_id" binding:"required"`
	ScheduledAt  time.Time `json:"scheduled_at" binding:"required"`
	Duration     int       `json:"duration" binding:"required"`
}

func NewBookingServer() *BookingServer {
	router := gin.Default()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/consulting_db?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return &BookingServer{
		router: router,
		db:     db,
	}
}

func (bs *BookingServer) SetupRoutes() {
	bs.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	bs.router.GET("/bookings", bs.handleGetBookings)
	bs.router.POST("/bookings", bs.handleCreateBooking)
	bs.router.GET("/bookings/:id", bs.handleGetBooking)
	bs.router.PUT("/bookings/:id", bs.handleUpdateBooking)
}

func (bs *BookingServer) handleGetBookings(c *gin.Context) {
	rows, err := bs.db.Query("SELECT id, consultant_id, user_id, scheduled_at, duration, status FROM bookings")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch bookings"})
		return
	}
	defer rows.Close()

	bookings := []Booking{}
	for rows.Next() {
		var booking Booking
		if err := rows.Scan(&booking.ID, &booking.ConsultantID, &booking.UserID, &booking.ScheduledAt, &booking.Duration, &booking.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to scan booking"})
			return
		}
		bookings = append(bookings, booking)
	}

	c.JSON(http.StatusOK, gin.H{"bookings": bookings})
}

func (bs *BookingServer) handleCreateBooking(c *gin.Context) {
	var req BookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	booking := Booking{
		ID:           "booking-" + time.Now().Unix().String(),
		ConsultantID: req.ConsultantID,
		UserID:       req.UserID,
		ScheduledAt:  req.ScheduledAt,
		Duration:     req.Duration,
		Status:       "confirmed",
	}

	_, err := bs.db.Exec(
		"INSERT INTO bookings (id, consultant_id, user_id, scheduled_at, duration, status) VALUES ($1, $2, $3, $4, $5, $6)",
		booking.ID, booking.ConsultantID, booking.UserID, booking.ScheduledAt, booking.Duration, booking.Status,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create booking"})
		return
	}

	c.JSON(http.StatusCreated, booking)
}

func (bs *BookingServer) handleGetBooking(c *gin.Context) {
	id := c.Param("id")
	var booking Booking

	err := bs.db.QueryRow(
		"SELECT id, consultant_id, user_id, scheduled_at, duration, status FROM bookings WHERE id = $1",
		id,
	).Scan(&booking.ID, &booking.ConsultantID, &booking.UserID, &booking.ScheduledAt, &booking.Duration, &booking.Status)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "booking not found"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch booking"})
		return
	}

	c.JSON(http.StatusOK, booking)
}

func (bs *BookingServer) handleUpdateBooking(c *gin.Context) {
	id := c.Param("id")
	var req Booking
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := bs.db.Exec(
		"UPDATE bookings SET status = $1 WHERE id = $2",
		req.Status, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update booking"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "booking updated"})
}

func (bs *BookingServer) Start(port string) error {
	return bs.router.Run(":" + port)
}

func main() {
	godotenv.Load()

	port := os.Getenv("BOOKING_PORT")
	if port == "" {
		port = "3003"
	}

	server := NewBookingServer()
	server.SetupRoutes()

	log.Printf("Booking server starting on port %s\n", port)
	if err := server.Start(port); err != nil {
		log.Fatalf("Failed to start booking server: %v", err)
	}
}
