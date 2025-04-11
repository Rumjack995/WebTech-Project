package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	_ "github.com/lib/pq"
)

var db *sql.DB
var jwtKey = []byte("secret")

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func main() {
	var err error
	db, err = sql.Open("postgres", "user=postgres password=7744 dbname=qr_event_checkin sslmode=disable")
	if err != nil {
		log.Fatal("Error opening DB:", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal("Cannot connect to DB:", err)
	}
	log.Println("✅ Connected to PostgreSQL!")

	// Clear only the checkins table, not users
	db.Exec("DELETE FROM checkins")

	// Create tables if not already there
	db.Exec("CREATE TABLE IF NOT EXISTS users (username TEXT PRIMARY KEY, password TEXT)")
	db.Exec("CREATE TABLE IF NOT EXISTS checkins (id SERIAL PRIMARY KEY, username TEXT, time TEXT)")

	// Clear checkins table on server start
	_, err = db.Exec("DELETE FROM checkins")
	if err != nil {
		log.Println("❌ Error clearing checkins table:", err)
	} else {
		log.Println("🧹 Cleared checkins table on startup.")
	}

	r := gin.Default()

	// 👇 Enable CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	r.POST("/register", register)
	r.POST("/login", login)

	auth := r.Group("/")
	auth.Use(authMiddleware())
	auth.POST("/checkin", checkin)
	auth.GET("/checkins", getCheckins)

	r.Run(":8080")
}

func register(c *gin.Context) {
	var user User
	c.BindJSON(&user)
	_, err := db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, user.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Registered"})
}

func login(c *gin.Context) {
	var user User
	c.BindJSON(&user)
	row := db.QueryRow("SELECT password FROM users WHERE username=$1", user.Username)
	var pw string
	if err := row.Scan(&pw); err != nil || pw != user.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	expiration := time.Now().Add(time.Hour)
	claims := &Claims{
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtKey)

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func checkin(c *gin.Context) {
	claims := c.MustGet("claims").(*Claims)

	// Check if already checked in
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM checkins WHERE username=$1)", claims.Username).Scan(&exists)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Already checked in"})
		return
	}

	// Insert new check-in
	_, err = db.Exec("INSERT INTO checkins (username, time) VALUES ($1, $2)", claims.Username, time.Now().Format(time.RFC3339))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check in"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Checked in"})
}

func getCheckins(c *gin.Context) {
	rows, err := db.Query("SELECT username, time FROM checkins ORDER BY time DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch check-ins"})
		return
	}
	defer rows.Close()

	var checkins []map[string]string
	for rows.Next() {
		var username, timeStr string
		if err := rows.Scan(&username, &timeStr); err != nil {
			continue
		}
		checkins = append(checkins, map[string]string{
			"username": username,
			"time":     timeStr,
		})
	}

	c.JSON(http.StatusOK, gin.H{"checkins": checkins})
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}
