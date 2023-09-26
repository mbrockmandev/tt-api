package handlers

import (
	"net/http"

	"github.com/mbrockmandev/tometracker/internal/app"
)

type Handler struct {
	App *app.Application
}

type AppHandlers interface {
	// Test Handlers
	KeepApiAlive(http.ResponseWriter, *http.Request)

	// Auth Handlers
	RegisterNewUser(http.ResponseWriter, *http.Request)
	Login(http.ResponseWriter, *http.Request)
	GetRefreshToken(http.ResponseWriter, *http.Request)
	Logout(http.ResponseWriter, *http.Request)

	// Book Handlers
	GetPopularBooks(http.ResponseWriter, *http.Request)
	GetBooksByQuery(http.ResponseWriter, *http.Request)
	GetBookById(http.ResponseWriter, *http.Request)
	GetBookByIsbn(http.ResponseWriter, *http.Request)
	GetBooksByAuthor(http.ResponseWriter, *http.Request)

	BorrowBook(http.ResponseWriter, *http.Request)
	ReturnBook(http.ResponseWriter, *http.Request)
	CreateBook(http.ResponseWriter, *http.Request)
	UpdateBook(http.ResponseWriter, *http.Request)
	DeleteBook(http.ResponseWriter, *http.Request)

	// User Handlers
	GetUserById(http.ResponseWriter, *http.Request)
	GetUserByEmail(http.ResponseWriter, *http.Request)
	CreateUser(http.ResponseWriter, *http.Request)
	UpdateUser(http.ResponseWriter, *http.Request)
	DeleteUser(http.ResponseWriter, *http.Request)
	GetUserInfoFromJWT(http.ResponseWriter, *http.Request)
	GetUserLibrary(http.ResponseWriter, *http.Request)
	UpdateUserLibrary(http.ResponseWriter, *http.Request)
	GetBorrowedBooksByUser(http.ResponseWriter, *http.Request)
	GetRecentlyReturnedBooksByUser(http.ResponseWriter, *http.Request)

	// Library Handlers
	CreateLibrary(http.ResponseWriter, *http.Request)
	GetAllLibraries(http.ResponseWriter, *http.Request)
	GetLibraryById(http.ResponseWriter, *http.Request)
	GetLibraryByName(http.ResponseWriter, *http.Request)
	UpdateLibrary(http.ResponseWriter, *http.Request)
	DeleteLibrary(http.ResponseWriter, *http.Request)

	// Report Handlers
	ReportPopularBooks(http.ResponseWriter, *http.Request)
	ReportPopularGenres(http.ResponseWriter, *http.Request)
	ReportBusyTimes(http.ResponseWriter, *http.Request)
	ReportBooksByLibrary(http.ResponseWriter, *http.Request)

	// Middleware
	EnableCors(h http.Handler) http.Handler
	RequireAuth(h http.Handler) http.Handler
	RateLimit(h http.Handler) http.Handler
	RequireRole(reqRole string) func(next http.Handler) http.Handler

	// Recommender
	RecommendedBooks(http.ResponseWriter, *http.Request)
}
