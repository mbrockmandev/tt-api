package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/mbrockmandev/tometracker/internal/jsonHelper"
	"github.com/mbrockmandev/tometracker/internal/models"
)

type BookReport struct {
	ID          int `json:"id"`
	Title       int `json:"title"`
	Author      int `json:"author"`
	BorrowCount int `json:"borrow_count"`
}

type BookMetadata struct {
	ID          int `json:"id"`
	TotalCopies int `json:"total_copies"`
	Available   int `json:"available"`
	Borrowed    int `json:"borrowed"`
}

func (h *Handler) ReportPopularBooks(w http.ResponseWriter,
	r *http.Request) {
	books, err := h.App.DB.ReportPopularBooks()
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("failed to retrieve popular books report: %s", err), http.StatusBadRequest)
		return
	}

	if books == nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("no books found"), http.StatusNotFound)
		return
	}

	jsonHelper.WriteJson(w, http.StatusOK, books)
}

func (h *Handler) ReportPopularGenres(w http.ResponseWriter,
	r *http.Request) {
	genres, err := h.App.DB.ReportPopularGenres()
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("failed to retrieve popular genres report: %s", err), http.StatusBadRequest)
		return
	}

	if genres == nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("no genres found"), http.StatusNotFound)
		return
	}

	jsonHelper.WriteJson(w, http.StatusOK, genres)
}

func (h *Handler) ReportBusyTimes(w http.ResponseWriter,
	r *http.Request) {
	busyTimes, err := h.App.DB.ReportBusyTimes()
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("failed to retrieve busy times report: %s", err), http.StatusBadRequest)
		return
	}

	if busyTimes == nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("no times found"), http.StatusNotFound)
		return
	}

	jsonHelper.WriteJson(w, http.StatusOK, busyTimes)
}

func (h *Handler) ReportBooksByLibrary(w http.ResponseWriter, r *http.Request) {

	libraryId, err := strconv.Atoi(r.URL.Query().Get("library_id"))
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid library id query parameter"), http.StatusBadRequest)
		return
	}

	if libraryId == 0 {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid library id of id 0"), http.StatusBadRequest)
		return
	}

	library, err := h.App.DB.GetLibraryById(libraryId)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("failed to retrieve library with id %v", libraryId), http.StatusBadRequest)
		return
	}

	if library == nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("library not found"), http.StatusNotFound)
		return
	}

	books, bookData, err := h.App.DB.GetBooksByLibrary(libraryId)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("failed to retrieve books report: %s", err), http.StatusBadRequest)
		return
	}

	if books == nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("no books found"), http.StatusNotFound)
		return
	}

	type BookWithMetadata struct {
		Book     *models.Book         `json:"book"`
		Metadata *models.BookMetadata `json:"metadata"`
	}

	var booksWithMetadata []BookWithMetadata
	for _, book := range books {
		var bookMetadata *models.BookMetadata
		for _, data := range bookData {
			if data.BookId == book.ID {
				bookMetadata = data
				break
			}
		}
		booksWithMetadata = append(booksWithMetadata, BookWithMetadata{
			Book:     book,
			Metadata: bookMetadata,
		})
	}

	metadata := map[string]int{
		"currentPage": 1,
		"totalPages":  len(books) / (10),
	}

	response := struct {
		Books    []BookWithMetadata `json:"books"`
		Metadata map[string]int     `json:"metadata"`
	}{
		Books:    booksWithMetadata,
		Metadata: metadata,
	}

	jsonHelper.WriteJson(w, http.StatusOK, response)
}
