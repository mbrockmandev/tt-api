package database

import (
	"fmt"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/assert"

	"github.com/mbrockmandev/tometracker/internal/models"
)

func TestCreateBook(t *testing.T) {
	repo := setupTestDB()
	defer teardownTestDB(repo)

	l, err := repo.GetLibraryByName("Central Library")
	if err != nil {
		t.Fatal(err)
	}

	book := models.Book{
		Title:       "Test Title",
		Author:      "Test Author",
		ISBN:        "9999999999999",
		PublishedAt: time.Now(),
		Summary:     "Test summary",
		Thumbnail:   "https://www.google.com/",
	}

	id, err := repo.CreateBook(&book, l.ID)
	if err != nil {
		t.Fatal(err)
	}

	_ = repo.DeleteBook(id)
}

func TestUpdateBook(t *testing.T) {
	repo := setupTestDB()
	defer teardownTestDB(repo)

	l, err := repo.GetLibraryByName("Central Library")
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.GetLibraryById(l.ID)
	if err != nil {
		t.Fatal(err)
	}

	book := models.Book{
		Title:       "Test Title",
		Author:      "Test Author",
		ISBN:        "9999999999999",
		PublishedAt: time.Now(),
		Summary:     "Test summary",
		Thumbnail:   "https://www.google.com/",
	}

	bookId, err := repo.CreateBook(&book, l.ID)
	if err != nil {
		t.Log(err)
	}

	addedBook, _, err := repo.GetBookById(bookId)
	if err != nil {
		t.Log(err)
	}

	t.Log("addedBook.Summary: ", addedBook.Summary)

	book.Summary = "There can be only one."

	err = repo.UpdateBook(addedBook.ID, &book)
	if err != nil {
		t.Log(err)
	}

	updatedBook, _, _ := repo.GetBookById(addedBook.ID)

	assert.Equal(
		t,
		book.Summary,
		updatedBook.Summary,
		"mismatched result, must not have updated correctly",
	)

	_ = repo.DeleteBook(addedBook.ID)
}

func TestDeleteBook(t *testing.T) {
	repo := setupTestDB()
	defer teardownTestDB(repo)

	l, err := repo.GetLibraryByName("Central Library")
	if err != nil {
		t.Fatal(err)
	}

	book := models.Book{
		Title:       "Test Title",
		Author:      "Test Author",
		ISBN:        "9999999999999",
		PublishedAt: time.Now(),
		Summary:     "Test summary",
	}

	existingBook, _, err := repo.GetBookById(book.ID)
	if err != nil {
		t.Log(err)
	}
	if existingBook != nil {
		err = repo.DeleteBook(existingBook.ID)
		if err != nil {
			t.Fatal(err)
		}
		return
	}

	_, err = repo.CreateBook(&book, l.ID)
	if err != nil {
		t.Fatal(err)
	}

	addedBook, err := repo.GetBookByIsbn("9999999999999")
	if err != nil {
		t.Log(err)
	}

	err = repo.DeleteBook(addedBook.ID)
	if err != nil {
		t.Fatal(err)
	}

	shouldNotBeFound, err := repo.GetBookByIsbn("9999999999999")
	if err == nil {
		t.Fatal("Book should have been deleted")
	}
	if shouldNotBeFound != nil {
		t.Fatal("Book should have been deleted")
	}
}

func TestGetAllBooks(t *testing.T) {
	repo := setupTestDB()
	defer teardownTestDB(repo)

	books, err := repo.GetAllBooks()
	if err != nil {
		fmt.Println(err)
	}
	if books == nil {
		t.Fatal("books should not be nil")
	}

	if len(books) == 0 {
		t.Fatal("expected at least one user, got none")
	}

	for _, book := range books {
		t.Log(book.Title)
	}
}

func TestBorrowBook(t *testing.T) {
	repo := setupTestDB()
	defer teardownTestDB(repo)

	l, err := repo.GetLibraryByName("Central Library")
	if err != nil {
		t.Fatal(err)
	}

	book, err := repo.GetBookByIsbn("9780143124672")
	if err != nil {
		t.Fatal(err)
	}
	user, err := repo.GetUserByEmail("admin@example.com")
	if err != nil {
		t.Fatal(err)
	}
	_, err = repo.BorrowBook(user.ID, book.ID, l.ID)
	if err != nil {
		t.Fatal(err)
	}

	_, _ = repo.ReturnBook(book.ID, user.ID, l.ID)
}

func TestReturnBook(t *testing.T) {
	repo := setupTestDB()
	defer teardownTestDB(repo)

	l, err := repo.GetLibraryByName("Central Library")
	if err != nil {
		t.Fatal(err)
	}

	book, err := repo.GetBookByIsbn("9780143124672")
	if err != nil {
		t.Fatal(err)
	}
	user, err := repo.GetUserByEmail("admin@example.com")
	if err != nil {
		t.Fatal(err)
	}
	_, err = repo.ReturnBook(user.ID, book.ID, l.ID)
	if err != nil {
		t.Logf("error: %v", err)
	}

	_, _ = repo.BorrowBook(user.ID, book.ID, l.ID)
}

func TestReportPopularBooks(t *testing.T) {
	repo := setupTestDB()
	defer teardownTestDB(repo)

	books, err := repo.ReportPopularBooks()
	if err != nil {
		t.Fatalf("Failed to retrieve popular books report: %s", err)
	}

	if len(books) == 0 {
		t.Fatalf("expected at least one book, got none")
	}

	for _, book := range books {
		t.Logf(book.Title)
	}
}
