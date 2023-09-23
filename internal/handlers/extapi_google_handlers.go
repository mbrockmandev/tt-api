package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mbrockmandev/tometracker/internal/jsonHelper"
	"github.com/mbrockmandev/tometracker/internal/models"
)

type GoogleBooksResponse struct {
	Items []struct {
		VolumeInfo struct {
			Title               string   `json:"title"`
			Authors             []string `json:"authors"`
			PublishedDate       string   `json:"publishedDate"`
			Description         string   `json:"description"`
			IndustryIdentifiers []struct {
				Type       string `json:"type"`
				Identifier string `json:"identifier"`
			} `json:"industryIdentifiers"`
		} `json:"volumeInfo"`
	} `json:"items"`
}

func (h *Handler) GetGoogleBooksByQuery(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	apiKey := os.Getenv("GOOG_API_KEY")
	query := chi.URLParam(r, "query")
	url := fmt.Sprintf("https://www.googleapis.com/books/v1/volumes?q=%s&key=%s&maxResults=10", query, apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	res, err := client.Do(req)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("API returned status %v", res.StatusCode)
		jsonHelper.ErrorJson(w, errors.New(errMsg), res.StatusCode)
		return
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	var booksResponse GoogleBooksResponse
	if err := json.Unmarshal(bodyBytes, &booksResponse); err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	var books []models.Book
	for _, item := range booksResponse.Items {
		book := models.Book{
			Title:       item.VolumeInfo.Title,
			Author:      strings.Join(item.VolumeInfo.Authors, ", "),
			PublishedAt: parsePublishedDate(item.VolumeInfo.PublishedDate),
			Summary:     item.VolumeInfo.Description,
			ISBN:        getISBN(item.VolumeInfo.IndustryIdentifiers),
		}
		books = append(books, book)
	}

	if err := jsonHelper.WriteJson(w, http.StatusOK, books); err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
	}
}

func (h *Handler) GetGoogleBookByISBN(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	apiKey := os.Getenv("GOOG_API_KEY")
	isbn := chi.URLParam(r, "isbn")
	url := fmt.Sprintf("https://www.googleapis.com/books/v1/volumes?q=+isbn=%s&key=%s&maxResults=10", isbn, apiKey)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	res, err := client.Do(req)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("API returned status %v", res.StatusCode)
		jsonHelper.ErrorJson(w, errors.New(errMsg), res.StatusCode)
		return
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	var booksResponse GoogleBooksResponse
	if err := json.Unmarshal(bodyBytes, &booksResponse); err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	var books []models.Book
	for _, item := range booksResponse.Items {
		book := models.Book{
			Title:       item.VolumeInfo.Title,
			Author:      strings.Join(item.VolumeInfo.Authors, ", "),
			PublishedAt: parsePublishedDate(item.VolumeInfo.PublishedDate),
			Summary:     item.VolumeInfo.Description,
			ISBN:        getISBN(item.VolumeInfo.IndustryIdentifiers),
		}
		books = append(books, book)
	}

	if err := jsonHelper.WriteJson(w, http.StatusOK, books); err != nil {
		jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
	}
}

func parsePublishedDate(date string) time.Time {
	t, _ := time.Parse("2006-01-02", date)
	return t
}

func getISBN(identifiers []struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
}) string {
	for _, id := range identifiers {
		if id.Type == "ISBN_13" {
			return id.Identifier
		}
	}
	return ""
}
