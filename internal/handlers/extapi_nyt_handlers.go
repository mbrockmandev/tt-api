package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/mbrockmandev/tometracker/internal/jsonHelper"
	"github.com/mbrockmandev/tometracker/internal/models"
)

type NYTimesResponse struct {
	Status     string `json:"status"`
	Copyright  string `json:"copyright"`
	NumResults int    `json:"num_results"`
	Results    []struct {
		Title           string `json:"title"`
		Description     string `json:"description"`
		Contributor     string `json:"contributor"`
		Author          string `json:"author"`
		ContributorNote string `json:"contributor_note"`
		Price           string `json:"price"`
		AgeGroup        string `json:"age_group"`
		Publisher       string `json:"publisher"`
		Isbns           []struct {
			Isbn10 string `json:"isbn10"`
			Isbn13 string `json:"isbn13"`
		} `json:"isbns"`
		RanksHistory []struct {
			PrimaryIsbn10   string `json:"primary_isbn10"`
			PrimaryIsbn13   string `json:"primary_isbn13"`
			Rank            int    `json:"rank"`
			ListName        string `json:"list_name"`
			DisplayName     string `json:"display_name"`
			PublishedDate   string `json:"published_date"`
			BestsellersDate string `json:"bestsellers_date"`
			WeeksOnList     int    `json:"weeks_on_list"`
			RankLastWeek    int    `json:"rank_last_week"`
			Asterisk        int    `json:"asterisk"`
			Dagger          int    `json:"dagger"`
		} `json:"ranks_history"`
		Reviews []struct {
			BookReviewLink     string `json:"book_review_link"`
			FirstChapterLink   string `json:"first_chapter_link"`
			SundayReviewLink   string `json:"sunday_review_link"`
			ArticleChapterLink string `json:"article_chapter_link"`
		} `json:"reviews"`
	} `json:"results"`
}

func (h *Handler) GetPopularBooksFromNYT(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}

	url := fmt.Sprintf("https://api.nytimes.com/svc/books/v3/lists/best-sellers/history.json?api-key=%s", os.Getenv("NYT_API_KEY"))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		_ = jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	res, err := client.Do(req)
	if err != nil {
		_ = jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("API returned status %v", res.StatusCode)
		_ = jsonHelper.ErrorJson(w, errors.New(errMsg), res.StatusCode)
		return
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		_ = jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	var apiResponse NYTimesResponse
	if err := json.Unmarshal(bodyBytes, &apiResponse); err != nil {
		_ = jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	var books []models.Book
	for _, item := range apiResponse.Results {
		// get isbn10, if exists
		isbn10 := ""
		if len(item.Isbns) > 0 {
			isbn10 = item.Isbns[0].Isbn10
		}

		// get and parse published date
		publishedAt := time.Time{}
		if len(item.RanksHistory) > 0 {
			publishedDate := item.RanksHistory[0].PublishedDate
			publishedAt = parsePublishedDate(publishedDate)
		}

		book := models.Book{
			Title:       item.Title,
			Author:      item.Author,
			PublishedAt: publishedAt,
			Summary:     item.Description,
			ISBN:        isbn10,
		}
		books = append(books, book)
	}

	if err := jsonHelper.WriteJson(w, http.StatusOK, books); err != nil {
		_ = jsonHelper.ErrorJson(w, err, http.StatusInternalServerError)
	}
}
