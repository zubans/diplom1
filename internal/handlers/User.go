package handlers

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"time"
)

type Server struct {
	DB        *sql.DB
	JWTSecret []byte
}

func New(db *sql.DB, secret string) *Server {
	return &Server{DB: db, JWTSecret: []byte(secret)}
}

func (s *Server) Register(c *gin.Context) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	_, err = s.DB.Exec(
		"INSERT INTO users (username, password_hash) VALUES ($1, $2)",
		req.Login, string(hashedPassword))

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "login already exists"})
		return
	}

	s.generateAndSendToken(c, req.Login)
}

func (s *Server) Login(c *gin.Context) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	var storedHash string
	row := s.DB.QueryRow(
		"SELECT password_hash FROM users WHERE username = $1", req.Login)
	if err := row.Scan(&storedHash); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(storedHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	s.generateAndSendToken(c, req.Login)
}

func (s *Server) generateAndSendToken(c *gin.Context, login string) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": login,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(s.JWTSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.Header("Authorization", "Bearer "+tokenString)
	c.Status(http.StatusOK)
}
