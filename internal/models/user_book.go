package models

import "time"

type UserBook struct {
	ID         int       `db:"id" json:"id"`
	UserID     int       `db:"user_id" json:"user_id"`
	BookID     int       `db:"book_id" json:"book_id"`
	DueDate    time.Time `db:"due_date" json:"due_date"`
	ReturnedAt time.Time `db:"returned_at" json:"returned_at"`
	BorrowedAt time.Time `db:"borrowed_at" json:"borrowed_at"`
}
