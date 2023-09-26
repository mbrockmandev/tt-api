package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/mbrockmandev/tometracker/internal/handlers"
)

func NewRouter(h handlers.AppHandlers) http.Handler {
	r := chi.NewRouter()
	r.Route("/api", func(r chi.Router) {
		// middleware
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"https://tometracker-react.onrender.com", "http://tometracker-react.onrender.com"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			AllowCredentials: true,
			ExposedHeaders:   []string{"Link"},
			MaxAge:           300,
		}))

		r.Use(middleware.Recoverer)

		r.Post("/register", h.RegisterNewUser)
		r.Route("/auth", func(r chi.Router) {
			r.Use(h.RateLimit)
			r.Post("/login", h.Login)
		})

		r.Get("/logout", h.Logout)
		r.Get("/refresh", h.GetRefreshToken)
		r.Get("/jwtInfo", h.GetUserInfoFromJWT)

		// book routes
		r.Get("/books/popular", h.GetPopularBooks)
		r.Get("/books/{id}", h.GetBookById)
		r.Get("/books/isbn/{isbn}", h.GetBookByIsbn)
		r.Get("/books/author/{author}", h.GetBooksByAuthor)
		r.Get("/books/search", h.GetBooksByQuery)

		r.Get("/libraries", h.GetAllLibraries)
		r.Get("/libraries/{id}", h.GetLibraryById)

		// user only routes
		r.Route("/users/", func(r chi.Router) {
			r.Use(h.RequireRole("user"))

			r.Get("/{id}", h.GetUserById)
			r.Get("/{id}/borrowed", h.GetBorrowedBooksByUser)
			r.Get("/{id}/returned", h.GetRecentlyReturnedBooksByUser)
			r.Get("/{id}/homeLibrary", h.GetUserLibrary)

			r.Put("/libraries/{id}/setHomeLocation", h.UpdateUserLibrary)

			r.Post("/books/borrow", h.BorrowBook)
			r.Post("/books/return", h.ReturnBook)
		})

		// staff only routes
		r.Route("/staff", func(r chi.Router) {
			r.Use(h.RequireRole("staff"))

			// user routes
			r.Get("/users/{id}", h.GetUserById)
			r.Get("/users", h.GetUserByEmail)
			r.Put("/users/{id}", h.UpdateUser)

			// book routes
			r.Get("/books/{id}", h.GetBookById)
			r.Get("/books/isbn/{isbn}", h.GetBookByIsbn)
			r.Put("/books/{id}", h.UpdateBook)

			// library routes
			r.Get("/libraries", h.GetLibraryByName)
			r.Get("/libraries/{id}", h.GetLibraryById)
			r.Put("/libraries/{id}", h.UpdateLibrary)

			// report routes
			r.Get("/reports/popularBooks", h.ReportPopularBooks)
			r.Get("/reports/popularGenres", h.ReportPopularGenres)
			r.Get("/reports/peakTimes", h.ReportBusyTimes)
			r.Get("/reports/booksByLibrary", h.ReportBooksByLibrary)
		})

		// admin only routes
		r.Route("/admin", func(r chi.Router) {
			r.Use(h.RequireRole("admin"))

			// routes from staff as well
			r.Get("/users/{id}", h.GetUserById)
			r.Get("/users", h.GetUserByEmail)
			r.Put("/users/{id}", h.UpdateUser)

			r.Get("/libraries", h.GetLibraryByName)
			r.Get("/libraries/{id}", h.GetLibraryById)
			r.Put("/libraries/{id}", h.UpdateLibrary)

			r.Get("/books/{id}", h.GetBookById)
			r.Get("/books/isbn/{isbn}", h.GetBookByIsbn)
			r.Put("/books/{id}", h.UpdateBook)

			// user routes
			r.Post("/users", h.CreateUser)
			r.Delete("/users/{id}", h.DeleteUser)

			// library routes
			r.Post("/libraries", h.CreateLibrary)
			r.Delete("/libraries/{id}", h.DeleteLibrary)

			// book routes
			r.Post("/books", h.CreateBook)
			r.Delete("/books/{id}", h.DeleteBook)
		})

	})

	return r
}
