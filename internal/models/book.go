package models

import "time"

type Book struct {
	ID          int       `db:"id" json:"id"`
	Title       string    `db:"title" json:"title"`
	Author      string    `db:"author" json:"author"`
	ISBN        string    `db:"isbn" json:"isbn"`
	PublishedAt time.Time `db:"published_at" json:"published_at"`
	Summary     string    `db:"summary" json:"summary"`
	Thumbnail   string    `db:"thumbnail" json:"thumbnail"`
	Edition     string    `db:"edition" json:"edition"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
