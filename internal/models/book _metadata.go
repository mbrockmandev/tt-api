package models

type BookMetadata struct {
	BookId          int `json:"book_id"`
	TotalCopies     int `json:"total_copies"`
	AvailableCopies int `json:"available_copies"`
	BorrowedCopies  int `json:"borrowed_copies"`
}
