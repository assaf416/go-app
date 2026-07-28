package main

import (
	"log"
	"os"

	"github.com/labstack/echo/v4/middleware"

	"goapp/internal/app"
	"goapp/internal/db"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "goapp.db"
	}

	conn := db.Open(dbPath)
	defer conn.Close()

	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		sessionKey = "dev-secret-change-me-in-production"
	}

	e := app.New(conn, sessionKey, "web/templates/*.html")
	e.Use(middleware.Logger())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatal(err)
	}
}
