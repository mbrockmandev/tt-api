package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mbrockmandev/tometracker/internal/jsonHelper"
	"github.com/mbrockmandev/tometracker/internal/models"
)

// debug only
func (h *Handler) KeepApiAlive(w http.ResponseWriter, r *http.Request) {
	payload := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Version string `json:"version"`
	}{
		Status:  "active",
		Message: "TomeTracker is running!",
		Version: "0.0.1",
	}

	jsonHelper.WriteJson(w, http.StatusOK, payload)
}

func (h *Handler) CreateBook(w http.ResponseWriter,
	r *http.Request,
) {
	var req struct {
		Book      models.Book `json:"book"`
		LibraryId int         `json:"library_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	_, err = h.App.DB.CreateBook(&req.Book, req.LibraryId)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	jsonHelper.WriteJson(
		w,
		http.StatusCreated,
		map[string]string{"message": "Book created successfully."},
	)
}

func (h *Handler) GetBookById(w http.ResponseWriter,
	r *http.Request,
) {
	bookId := chi.URLParam(r, "id")
	if bookId == "" {
		jsonHelper.ErrorJson(w, fmt.Errorf("book id is required"), http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(bookId)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid book id"), http.StatusBadRequest)
		return
	}

	var book *models.Book
	var bookMetadata *models.BookMetadata
	libraryId, err := strconv.Atoi(r.URL.Query().Get("library_id"))
	if err != nil {
		book, bookMetadata, err = h.App.DB.GetBookById(id)
		if err != nil {
			jsonHelper.ErrorJson(
				w,
				fmt.Errorf("failed to retrieve book with id %v; error: %s", id, err),
				http.StatusBadRequest,
			)
			return
		}
	} else {
		book, bookMetadata, err = h.App.DB.GetBookByIdForLibrary(id, libraryId)
		if err != nil {
			jsonHelper.ErrorJson(
				w,
				fmt.Errorf("failed to retrieve book with id %v at library %v; error: %s", id, libraryId, err),
				http.StatusBadRequest,
			)
			return
		}
	}

	if book == nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("book not found"), http.StatusNotFound)
		return
	}

	if bookMetadata == nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("book metadata not found"), http.StatusNotFound)
		return
	}

	type BookWithMetadata struct {
		Book     *models.Book         `json:"book"`
		Metadata *models.BookMetadata `json:"metadata"`
	}

	var bookWithMetadata BookWithMetadata

	bookWithMetadata.Book = book
	bookWithMetadata.Metadata = bookMetadata

	jsonHelper.WriteJson(w, http.StatusOK, bookWithMetadata)
}

func (h *Handler) GetBookByIsbn(w http.ResponseWriter,
	r *http.Request,
) {
	isbn := chi.URLParam(r, "isbn")
	if isbn == "" {
		jsonHelper.ErrorJson(w, fmt.Errorf("isbn is required"), http.StatusBadRequest)
		return
	}

	var book *models.Book
	var bookMetadata *models.BookMetadata

	libraryId, err := strconv.Atoi(r.URL.Query().Get("library_id"))
	if err != nil {
		book, bookMetadata, err = h.App.DB.GetBookByIsbn(isbn)
		if err != nil {
			jsonHelper.ErrorJson(
				w,
				fmt.Errorf("failed to retrieve book with isbn %v; error: %s", isbn, err),
				http.StatusBadRequest,
			)
			return
		}
	} else {
		book, bookMetadata, err = h.App.DB.GetBookByIsbnForLibrary(isbn, libraryId)
		if err != nil {
			jsonHelper.ErrorJson(
				w,
				fmt.Errorf("failed to retrieve book with isbn %v at library %v; error: %s", isbn, libraryId, err),
				http.StatusBadRequest,
			)
			return
		}
	}

	if book == nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("book not found"), http.StatusNotFound)
		return
	}

	if bookMetadata == nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("book metadata not found"), http.StatusNotFound)
		return
	}

	type BookWithMetadata struct {
		Book     *models.Book         `json:"book"`
		Metadata *models.BookMetadata `json:"metadata"`
	}

	var bookWithMetadata BookWithMetadata

	bookWithMetadata.Book = book
	bookWithMetadata.Metadata = bookMetadata

	jsonHelper.WriteJson(w, http.StatusOK, bookWithMetadata)
}

func (h *Handler) GetPopularBooks(w http.ResponseWriter,
	r *http.Request,
) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		jsonHelper.ErrorJson(
			w,
			fmt.Errorf("page not included as a url param: %v", err),
			http.StatusBadRequest,
		)
		return
	}

	if page <= 0 {
		page = 1
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		jsonHelper.ErrorJson(
			w,
			fmt.Errorf("limit not included as a url param: %v", err),
			http.StatusBadRequest,
		)
		return
	}

	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit

	books, totalBooks, err := h.App.DB.GetPopularBooks(limit, offset)
	if err != nil {
		jsonHelper.ErrorJson(
			w,
			fmt.Errorf("failed to retrieve books: %v", err),
			http.StatusBadRequest,
		)
		return
	}

	metadata := map[string]int{
		"currentPage": page,
		"totalPages":  int(math.Ceil(float64(totalBooks) / float64(limit))),
	}

	response := struct {
		Books    []*models.Book `json:"books"`
		Metadata map[string]int `json:"metadata"`
	}{
		Books:    books,
		Metadata: metadata,
	}

	jsonHelper.WriteJson(w, http.StatusOK, response)
}

func (h *Handler) GetBooksByQuery(w http.ResponseWriter,
	r *http.Request,
) {
	queryText := r.URL.Query().Get("q")
	queryType := r.URL.Query().Get("type")
	if queryText == "" || queryType == "" {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid query"), http.StatusBadRequest)
		return
	}

	validTypes := map[string]bool{"isbn": true, "author": true, "title": true}
	if !validTypes[queryType] {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid query type"), http.StatusBadRequest)
		return
	}

	var page int
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 1
	}

	if page <= 0 {
		page = 1
	}

	var limit int
	limit, err = strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		limit = 10
	}

	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit

	books, totalBooks, err := h.App.DB.GetBooksByQuery(queryText, queryType, limit, offset)
	if err != nil {
		jsonHelper.ErrorJson(
			w,
			fmt.Errorf("unable to complete the request: %s", err),
			http.StatusBadRequest,
		)
	}

	if books == nil {
		jsonHelper.WriteJson(w, http.StatusNoContent, nil)
		return
	}

	metadata := map[string]int{
		"currentPage": page,
		"totalPages":  int(math.Ceil(float64(totalBooks) / float64(limit))),
	}

	response := struct {
		Books    []*models.Book `json:"books"`
		Metadata map[string]int `json:"metadata"`
	}{
		Books:    books,
		Metadata: metadata,
	}

	jsonHelper.WriteJson(w, http.StatusOK, response)
}

func (h *Handler) GetBooksByAuthor(w http.ResponseWriter,
	r *http.Request,
) {
	author := r.URL.Query().Get("author")
	if author == "" {
		jsonHelper.ErrorJson(w, fmt.Errorf("author is required"), http.StatusBadRequest)
		return
	}

	books, err := h.App.DB.GetBooksByAuthor(author)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	jsonHelper.WriteJson(w, http.StatusOK, books)
}

func (h *Handler) BorrowBook(w http.ResponseWriter,
	r *http.Request,
) {
	var req struct {
		UserId    int `json:"user_id"`
		BookId    int `json:"book_id"`
		LibraryId int `json:"library_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	bookMetadata, err := h.App.DB.BorrowBook(req.UserId, req.BookId, req.LibraryId)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusConflict)
		return
	}

	jsonHelper.WriteJson(
		w,
		http.StatusOK,
		bookMetadata,
	)
}

func (h *Handler) ReturnBook(w http.ResponseWriter,
	r *http.Request,
) {
	var req struct {
		UserId    int `json:"user_id"`
		BookId    int `json:"book_id"`
		LibraryId int `json:"library_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid request body"), http.StatusBadRequest)
		return
	}

	bookMetadata, err := h.App.DB.ReturnBook(req.UserId, req.BookId, req.LibraryId)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	jsonHelper.WriteJson(
		w,
		http.StatusOK,
		bookMetadata,
	)
}

func (h *Handler) UpdateBook(w http.ResponseWriter,
	r *http.Request,
) {
	bookIdStr := chi.URLParam(r, "id")
	bookId, err := strconv.Atoi(bookIdStr)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid id"), http.StatusBadRequest)
		return
	}
	var book models.Book

	err = json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	book.ID = bookId
	err = h.App.DB.UpdateBook(bookId, &book)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	jsonHelper.WriteJson(
		w,
		http.StatusOK,
		map[string]string{"message": "Book updated successfully."},
	)
}

func (h *Handler) DeleteBook(w http.ResponseWriter,
	r *http.Request,
) {
	bookIdStr := chi.URLParam(r, "id")
	bookId, err := strconv.Atoi(bookIdStr)
	if err != nil {
		jsonHelper.ErrorJson(w, fmt.Errorf("invalid id"), http.StatusBadRequest)
		return
	}

	err = h.App.DB.DeleteBook(bookId)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	jsonHelper.WriteJson(
		w,
		http.StatusOK,
		map[string]string{"message": "Book updated successfully."},
	)
}

func (h *Handler) RecommendedBooks(w http.ResponseWriter, r *http.Request) {
	// books, err := python.Run("recommended_books.py")
	// if err != nil {
	// 	jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
	// 	return
	// }

	jsonHelper.WriteJson(w, http.StatusOK, nil)
}
