package models

type BookGenre struct {
	ID      int `db:"id" json:"id"`
	BookID  int `db:"book_id" json:"book_id"`
	GenreID int `db:"genre_id" json:"genre_id"`
}
