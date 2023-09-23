package database

import (
	"database/sql"
	"time"

	"github.com/mbrockmandev/tometracker/internal/models"
)

type DatabaseRepo interface {
	Connection() *sql.DB

	// User Methods
	CreateUser(user *models.User) (int, error)
	GetAllUsers() ([]*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserRoleByEmail(email string) (string, error)
	GetUserById(id int) (*models.User, error)
	DeleteUser(id int) error
	UpdateUser(id int, user *models.User) error
	UpdateUserLibrary(libraryId, userId int) error

	// Book Methods
	CreateBook(book *models.Book, libraryId int) (int, error)
	GetAllBooks(genre ...int) ([]*models.Book, error)
	GetPopularBooks(limit, offset int) ([]*models.Book, int, error)
	GetBookById(id int) (*models.Book, *models.BookMetadata, error)
	GetBookByIdForLibrary(id, libraryId int) (*models.Book, *models.BookMetadata, error)
	GetBookByIsbn(ISBN string) (*models.Book, error)
	GetBooksByQuery(queryText, queryType string, limit, offset int) ([]*models.Book, int, error)
	GetBooksByAuthor(author string) ([]*models.Book, error)
	GetBorrowedBooksByUserId(id int) ([]*models.Book, error)
	GetRecentlyReturnedBooksByUserId(userId int) ([]*models.Book, error)
	GetUserHomeLibrary(id int) (int, error)

	BorrowBook(userId, bookId, libraryId int) (*models.BookMetadata, error)
	ReturnBook(userId, bookId, libraryId int) (*models.BookMetadata, error)
	DeleteBook(id int) error
	UpdateBook(id int, book *models.Book) error
	ReportPopularBooks() ([]*models.Book, error)

	// Genre Methods
	CreateGenre(genre *models.Genre) (int, error)
	GetAllGenres() ([]*models.Genre, error)
	GetGenreById(id int) (*models.Genre, error)
	DeleteGenre(id int) error
	UpdateGenre(id int, genre *models.Genre) error
	ReportPopularGenres() ([]*models.Genre, error)

	// Library Methods
	CreateLibrary(library *models.Library) (int, error)
	GetAllLibraries(limit, offset int) ([]*models.Library, int, error)
	GetLibraryById(id int) (*models.Library, error)
	GetLibraryByName(name string) (*models.Library, error)
	UpdateLibrary(id int, library *models.Library) error
	DeleteLibrary(id int) error

	// ActivityLog Methods
	CreateActivityLog(activity *models.ActivityLog) (int, error)
	GetAllActivityLogs() ([]*models.ActivityLog, error)
	GetActivityLogById(id int) (*models.ActivityLog, error)
	DeleteActivityLog(id int) error
	UpdateActivityLog(id int, activity *models.ActivityLog) error
	ReportBusyTimes() ([]*BusyTime, error)
	LogActivity(activity *models.ActivityLog) error

	// Report Methods
	GetBooksByLibrary(libraryId int) ([]*models.Book, []*models.BookMetadata, error)
}

const dbTimeout = time.Second * 3
