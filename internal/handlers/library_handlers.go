package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/mbrockmandev/tometracker/internal/jsonHelper"
	"github.com/mbrockmandev/tometracker/internal/models"
)

func (h *Handler) CreateLibrary(w http.ResponseWriter,
	r *http.Request) {
	var req struct {
		Name          string `db:"name" json:"name"`
		City          string `db:"city" json:"city"`
		StreetAddress string `db:"street_address" json:"street_address"`
		PostalCode    string `db:"postal_code" json:"postal_code"`
		Country       string `db:"country" json:"country"`
		Phone         string `db:"phone" json:"phone"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.City == "" || req.StreetAddress == "" || req.PostalCode == "" || req.Country == "" || req.Phone == "" {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid request body, field left blank: %v", err), http.StatusBadRequest)
		return
	}

	newLibrary := models.Library{
		Name:          req.Name,
		City:          req.City,
		StreetAddress: req.StreetAddress,
		PostalCode:    req.PostalCode,
		Country:       req.Country,
		Phone:         req.Phone,
	}

	_, err = h.App.DB.CreateLibrary(&newLibrary)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("unable to create library: %v", err), http.StatusInternalServerError)
		return
	}

	jsonHelper.WriteJson(w, http.StatusCreated, map[string]string{"message": "Library created successfully."})
}

func (h *Handler) GetAllLibraries(w http.ResponseWriter, r *http.Request) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("page not included as a url param: %v", err), http.StatusBadRequest)
		return
	}

	if page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("limit not included as a url param: %v", err), http.StatusBadRequest)
		return
	}

	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit

	libraries, totalLibraries, err := h.App.DB.GetAllLibraries(limit, offset)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("unable to retrieve library list: %v", err), http.StatusBadRequest)
		return
	}

	if libraries == nil {
		jsonHelper.WriteJson(w, http.StatusNoContent, nil)
		return
	}

	metadata := map[string]int{
		"currentPage": page,
		"totalPages":  int(math.Ceil(float64(totalLibraries) / float64(limit))),
	}

	response := struct {
		Libraries []*models.Library `json:"libraries"`
		Metadata  map[string]int    `json:"metadata"`
	}{
		Libraries: libraries,
		Metadata:  metadata,
	}

	jsonHelper.WriteJson(w, http.StatusOK, response)
}

func (h *Handler) GetLibraryById(w http.ResponseWriter,
	r *http.Request) {
	libraryId := chi.URLParam(r, "id")
	if libraryId == "" {
		jsonHelper.ErrorJson(w, fmt.Errorf("library id is required"), http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(libraryId)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid library id format"), http.StatusBadRequest)
		return
	}

	library, err := h.App.DB.GetLibraryById(id)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("failed to retrieve library with id %v", id), http.StatusBadRequest)
		return
	}

	if library == nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("library not found"), http.StatusNotFound)
		return
	}

	jsonHelper.WriteJson(w, http.StatusOK, library)
}

func (h *Handler) GetLibraryByName(w http.ResponseWriter,
	r *http.Request) {
	libraryName := r.URL.Query().Get("name")
	if libraryName == "" {
		jsonHelper.ErrorJson(w, fmt.Errorf("library name is required"), http.StatusBadRequest)
		return
	}

	lowerCaseLibraryName := strings.ToLower(libraryName)

	library, err := h.App.DB.GetLibraryByName(lowerCaseLibraryName)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("failed to retrieve library with name %v", libraryName), http.StatusBadRequest)
		return
	}

	if library == nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("library not found"), http.StatusNotFound)
		return
	}

	jsonHelper.WriteJson(w, http.StatusOK, library)
}

func (h *Handler) UpdateLibrary(w http.ResponseWriter,
	r *http.Request) {
	libraryIdStr := chi.URLParam(r, "id")
	libraryId, err := strconv.Atoi(libraryIdStr)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid id"), http.StatusBadRequest)
		return
	}
	var library models.Library

	err = json.NewDecoder(r.Body).Decode(&library)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	library.ID = libraryId
	err = h.App.DB.UpdateLibrary(libraryId, &library)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	jsonHelper.WriteJson(w, http.StatusOK, map[string]string{"message": "library updated successfully."})

}

func (h *Handler) DeleteLibrary(w http.ResponseWriter,
	r *http.Request) {
	libraryStr := chi.URLParam(r, "id")
	libraryId, err := strconv.Atoi(libraryStr)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid id"), http.StatusBadRequest)
		return
	}

	err = h.App.DB.DeleteLibrary(libraryId)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	jsonHelper.WriteJson(w, http.StatusOK, map[string]string{"message": "library updated successfully."})
}
