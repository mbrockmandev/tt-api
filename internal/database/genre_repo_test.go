package database

import (
	"testing"

	"github.com/mbrockmandev/tometracker/internal/models"
)

func TestCreateGenre(t *testing.T) {
	repo := setupTestDB()
	defer teardownTestDB(repo)

	_, err := repo.CreateGenre(&models.Genre{Name: "Sci-Fi"})
	if err != nil {
		t.Fatal(err)
	}
}
func TestGetAllGenres(t *testing.T) {
	repo := setupTestDB()
	defer teardownTestDB(repo)

	genres, err := repo.GetAllGenres()
	if err != nil {
		t.Fatalf("Failed to retrieve all genres: %s", err)
	}
	if len(genres) == 0 {
		t.Fatalf("expected at least one genre, got none")
	}

}
func TestGetGenreById(t *testing.T) {
	repo := setupTestDB()
	defer teardownTestDB(repo)

	id, err := repo.CreateGenre(&models.Genre{Name: "Sci-Fi"})
	if err != nil {
		t.Fatal(err)
	}

	genre, err := repo.GetGenreById(id)
	if err != nil {
		t.Fatalf("Failed to retrieve genre: %s", err)
	}
	if genre == nil {
		t.Fatal("genre should not be nil")
	}
}

func TestDeleteGenre(t *testing.T) {
	repo := setupTestDB()
	defer teardownTestDB(repo)

	id, err := repo.CreateGenre(&models.Genre{Name: "Sci-Fi"})
	if err != nil {
		t.Fatal(err)
	}

	err = repo.DeleteGenre(id)
	if err != nil {
		t.Fatal(err)
	}

	genre, _ := repo.GetGenreById(id)
	if genre != nil {
		t.Fatal("genre should be nil")
	}

}
func TestUpdateGenre(t *testing.T) {
	repo := setupTestDB()
	defer teardownTestDB(repo)

	g, err := repo.GetGenreByName("Drama")
	if err != nil {
		t.Fatal(err)
	}

	err = repo.UpdateGenre(g.ID, &models.Genre{Name: "Comedy"})
	if err != nil {
		t.Fatal(err)
	}

	testGenre, _ := repo.GetGenreById(g.ID)
	if testGenre.Name != "Comedy" {
		t.Fatal("genre name should be updated")
	}

	err = repo.UpdateGenre(g.ID, &models.Genre{Name: "Fantasy"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestReportPopularGenres(t *testing.T) {
	repo := setupTestDB()
	defer teardownTestDB(repo)

	u, err := repo.GetUserByEmail("admin@example.com")
	if err != nil {
		t.Fatal(err)
	}

	b, _, err := repo.GetBookByIsbn("9780143124672")
	if err != nil {
		t.Fatal(err)
	}
	b2, _, err := repo.GetBookByIsbn("9780307958934")
	if err != nil {
		t.Fatal(err)
	}

	l, err := repo.GetLibraryByName("Central Library")
	if err != nil {
		t.Fatal(err)
	}

	l2, err := repo.GetLibraryById(2)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := repo.BorrowBook(u.ID, b.ID, l.ID); err != nil {
		t.Logf("error: %v", err)
	}

	if _, err := repo.BorrowBook(u.ID, b2.ID, l2.ID); err != nil {
		t.Fatal(err)
	}

	genres, err := repo.ReportPopularGenres()
	if err != nil {
		t.Fatalf("Failed to retrieve popular genres report: %s", err)
	}

	if len(genres) == 0 {
		t.Fatalf("expected at least one genre, got none")
	}

	for _, genre := range genres {
		t.Logf(genre.Name)
	}
}
