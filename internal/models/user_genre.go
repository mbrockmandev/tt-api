package models

type UserGenre struct {
	ID      int `db:"id" json:"id"`
	UserID  int `db:"user_id" json:"user_id"`
	GenreID int `db:"genre_id" json:"genre_id"`
}
