package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/joho/godotenv/autoload"

	"github.com/mbrockmandev/tometracker/internal/app"
	"github.com/mbrockmandev/tometracker/internal/auth"
	db "github.com/mbrockmandev/tometracker/internal/database"
	"github.com/mbrockmandev/tometracker/internal/handlers"
	"github.com/mbrockmandev/tometracker/internal/routes"
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func main() {
	var application app.Application

	port := os.Getenv("PORT")
	application.APIKey = os.Getenv("API_KEY")
	application.JWTSecret = os.Getenv("JWT_SECRET")
	application.JWTIssuer = os.Getenv("JWT_ISSUER")
	application.JWTAudience = os.Getenv("JWT_AUDIENCE")
	application.CookieDomain = os.Getenv("COOKIE_DOMAIN")
	application.JSONDBFile = os.Getenv("JSONDB_FILE")

	application.Auth = auth.Auth{
		Issuer:        application.JWTIssuer,
		Audience:      application.JWTAudience,
		Secret:        application.JWTSecret,
		TokenExpiry:   time.Minute * 1,
		RefreshExpiry: time.Hour * 24,
		CookiePath:    "/",
		CookieName:    "RefreshCookie",
		CookieDomain:  application.CookieDomain,
	}

	h := &handlers.Handler{
		App: &application,
	}

	conn, err := application.ConnectDB()
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	application.DB = &db.PostgresDBRepo{DB: conn}
	defer application.DB.Connection().Close()

	log.Println("Starting server on port", port)

	router := routes.NewRouter(h)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
