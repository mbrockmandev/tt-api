package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mbrockmandev/tometracker/internal/jsonHelper"
	"github.com/mbrockmandev/tometracker/internal/models"
)

func (h *Handler) GetUserById(w http.ResponseWriter,
	r *http.Request,
) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("user id is required"), http.StatusBadRequest)
		return
	}

	user, err := h.App.DB.GetUserById(id)
	if err != nil {
		jsonHelper.ErrorJson(
			w,
			fmt.Errorf("failed to retrieve user with id %v", id),
			http.StatusBadRequest,
		)
		return
	}

	if user == nil || user.ID == 0 {
		jsonHelper.ErrorJson(w, fmt.Errorf("user not found"), http.StatusNotFound)
		return
	}

	jsonHelper.WriteJson(w, http.StatusOK, user)
}

func (h *Handler) GetUserByEmail(w http.ResponseWriter,
	r *http.Request,
) {
	email := r.URL.Query().Get("email")
	if email == "" {
		jsonHelper.ErrorJson(w, fmt.Errorf("email is required"), http.StatusBadRequest)
		return
	}

	user, err := h.App.DB.GetUserByEmail(email)
	if err != nil {
		jsonHelper.ErrorJson(
			w,
			fmt.Errorf("failed to retrieve user with email %s", email),
			http.StatusBadRequest,
		)
		return
	}

	if user == nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("user not found"), http.StatusNotFound)
		return
	}

	jsonHelper.WriteJson(w, http.StatusOK, user)
}

func (h *Handler) GetUserRoleByEmail(w http.ResponseWriter,
	r *http.Request,
) {
	email := r.URL.Query().Get("email")
	if email == "" {
		jsonHelper.ErrorJson(w, fmt.Errorf("email is required"), http.StatusNotFound)
		return
	}

	role, err := h.App.DB.GetUserRoleByEmail(email)
	if err != nil {
		jsonHelper.ErrorJson(
			w,
			fmt.Errorf("failed to retrieve user with email %s", email),
			http.StatusBadRequest,
		)
		return
	}

	jsonHelper.WriteJson(w, http.StatusOK, role)
}

func (h *Handler) CreateUser(w http.ResponseWriter,
	r *http.Request,
) {
	var user models.User

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	_, err = h.App.DB.CreateUser(&user)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("failed to create user: %v", err), http.StatusInternalServerError)
		return
	}

	jsonHelper.WriteJson(
		w,
		http.StatusCreated,
		map[string]string{"message": "User created successfully."},
	)
}

func (h *Handler) UpdateUser(w http.ResponseWriter,
	r *http.Request,
) {
	userIdStr := chi.URLParam(r, "id")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid id"), http.StatusBadRequest)
		return
	}
	var user models.User

	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	user.ID = userId
	err = h.App.DB.UpdateUser(userId, &user)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	jsonHelper.WriteJson(
		w,
		http.StatusOK,
		map[string]string{"message": "User updated successfully."},
	)
}

func (h *Handler) DeleteUser(w http.ResponseWriter,
	r *http.Request,
) {
	userIdStr := chi.URLParam(r, "id")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid id"), http.StatusBadRequest)
		return
	}

	err = h.App.DB.DeleteUser(userId)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	jsonHelper.WriteJson(
		w,
		http.StatusOK,
		map[string]string{"message": "User deleted successfully."},
	)
}

func (h *Handler) UpdateUserLibrary(w http.ResponseWriter, r *http.Request) {
	libraryIdStr := chi.URLParam(r, "id")
	libraryId, err := strconv.Atoi(libraryIdStr)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid library id: %v", err), http.StatusBadRequest)
		return
	}

	if libraryId == 0 {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid library id: %v", err), http.StatusBadRequest)
		return
	}

	userIdStr := r.URL.Query().Get("userId")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid user id: %v", err), http.StatusBadRequest)
		return
	}

	err = h.App.DB.UpdateUserLibrary(libraryId, userId)
	if err != nil && err != sql.ErrNoRows {
		jsonHelper.ErrorJson(w, fmt.Errorf("conflicting info: %v", err), http.StatusConflict)
		return
	}
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("%v", err), http.StatusInternalServerError)
		return
	}

	jsonHelper.WriteJson(
		w,
		http.StatusOK,
		map[string]string{"message": "user's home library updated successfully."},
	)
}

func (h *Handler) GetUserLibrary(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(userIdStr)
	if err != nil {
		jsonHelper.ErrorJson(
			w,
			fmt.Errorf("invalid user id or id not included: %v", err),
			http.StatusBadRequest,
		)
		return
	}

	user, err := h.App.DB.GetUserById(id)
	if err != nil {
		jsonHelper.ErrorJson(
			w,
			fmt.Errorf("failed to retrieve user with id: %v", id),
			http.StatusBadRequest,
		)
		return
	}

	if user == nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("user not found"), http.StatusNotFound)
		return
	}

	homeLibraryId, err := h.App.DB.GetUserHomeLibrary(id)
	if err != nil {
		jsonHelper.ErrorJson(
			w,
			fmt.Errorf("failed to retrieve home library for user: %v", err),
			http.StatusBadRequest,
		)
		return
	}

	if homeLibraryId == 0 {
		jsonHelper.WriteJson(w, http.StatusNoContent, nil)
		return
	}

	jsonHelper.WriteJson(w, http.StatusOK, homeLibraryId)
}

func (h *Handler) GetBorrowedBooksByUser(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(userIdStr)
	if err != nil {
		jsonHelper.ErrorJson(
			w,
			fmt.Errorf("invalid user id or id not included: %v", err),
			http.StatusBadRequest,
		)
		return
	}

	user, err := h.App.DB.GetUserById(id)
	if err != nil {
		jsonHelper.ErrorJson(
			w,
			fmt.Errorf("failed to retrieve user with id %v", id),
			http.StatusBadRequest,
		)
		return
	}

	if user == nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("user not found"), http.StatusNotFound)
		return
	}

	borrowedBooks, err := h.App.DB.GetBorrowedBooksByUserId(id)
	if err != nil {
		jsonHelper.ErrorJson(
			w,
			fmt.Errorf("failed to retrieve borrowed books for user: %v", err),
			http.StatusBadRequest,
		)
		return
	}

	if borrowedBooks == nil {
		jsonHelper.WriteJson(w, http.StatusNoContent, nil)
		return
	}

	jsonHelper.WriteJson(w, http.StatusOK, borrowedBooks)
}

func (h *Handler) GetRecentlyReturnedBooksByUser(w http.ResponseWriter, r *http.Request) {
	userIdStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(userIdStr)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid user id: %v", err), http.StatusBadRequest)
		return
	}

	user, err := h.App.DB.GetUserById(id)
	if err != nil {
		jsonHelper.ErrorJson(
			w,
			fmt.Errorf("unable to find user with id: %v", id),
			http.StatusBadRequest,
		)
		return
	}

	if user == nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("user not found"), http.StatusNotFound)
	}

	returnedBooks, err := h.App.DB.GetRecentlyReturnedBooksByUserId(id)
	if err != nil {
		jsonHelper.ErrorJson(
			w,
			fmt.Errorf("problem retrieving recently returned books: %v", err),
			http.StatusInternalServerError,
		)
		return
	}

	if returnedBooks == nil {
		jsonHelper.WriteJson(w, http.StatusNoContent, nil)
		return
	}

	jsonHelper.WriteJson(w, http.StatusOK, returnedBooks)
}
