package models

type BookLibrary struct {
	ID              int `db:"id" json:"id"`
	BookID          int `db:"book_id" json:"book_id"`
	LibraryID       int `db:"library_id" json:"library_id"`
	TotalCopies     int `db:"total_copies" json:"total_copies"`
	AvailableCopies int `db:"available_copies" json:"available_copies"`
	BorrowedCopies  int `db:"borrowed_copies" json:"borrowed_copies"`
}
