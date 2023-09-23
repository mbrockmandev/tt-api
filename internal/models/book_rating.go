package models

type BookRating struct {
	ID     int `db:"id" json:"id"`
	UserID int `db:"user_id" json:"user_id"`
	BookID int `db:"book_id" json:"book_id"`
	Rating int `db:"rating" json:"rating"`
}
