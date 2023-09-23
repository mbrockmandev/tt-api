package app

import (
	"github.com/mbrockmandev/tometracker/internal/auth"
	repo "github.com/mbrockmandev/tometracker/internal/database"
)

type Application struct {
	DSN          string
	Domain       string
	DB           repo.DatabaseRepo
	Auth         auth.Auth
	JWTSecret    string
	JWTIssuer    string
	JWTAudience  string
	CookieDomain string
	APIKey       string
	JSONDBFile   string
}
