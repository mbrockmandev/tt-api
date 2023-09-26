package handlers

import (
	"fmt"
	"net/http"

	"golang.org/x/time/rate"

	"github.com/mbrockmandev/tometracker/internal/jsonHelper"
)

func (h *Handler) EnableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "https://tometracker-react.onrender.com")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().
				Set("Access-Control-Allow-Headers", "Accept, Content-Type, X-CSRF-Token, Authorization, Credentials")
			w.WriteHeader(http.StatusNoContent)
			return
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func (h *Handler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		_, err := r.Cookie("AccessCookie")
		if err != nil {
			jsonHelper.ErrorJson(w, fmt.Errorf("no auth token provided: %v", err), http.StatusUnauthorized)
			return
		}

		_, _, err = h.App.Auth.GetTokenAndVerify(w, r)
		if err != nil {
			jsonHelper.ErrorJson(w, fmt.Errorf("invalid token provided: %v", err), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) RequireRole(reqRole string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, claims, err := h.App.Auth.GetTokenAndVerify(w, r)
			if err != nil {
				jsonHelper.ErrorJson(w, fmt.Errorf("problem with the access token: %v", err), http.StatusUnauthorized)
				return
			}

			// admin can access user endpoints
			if reqRole == "user" && claims.Role == "admin" {
				next.ServeHTTP(w, r)
				return
			}

			// admin can also access staff endpoints
			if reqRole == "staff" && claims.Role == "admin" {
				next.ServeHTTP(w, r)
				return
			}

			// staff can access user endpoints
			if reqRole == "user" && claims.Role == "staff" {
				next.ServeHTTP(w, r)
				return
			}

			// mismatched
			if claims.Role != reqRole {
				jsonHelper.ErrorJson(w, fmt.Errorf("unauthorized: role mismatch (%s != %s)", claims.Role, reqRole), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (h *Handler) RateLimit(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(2, 10)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			jsonHelper.ErrorJson(w, fmt.Errorf("too many requests"), http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
