package database

import "context"

func (repo *PostgresDBRepo) AssignBookToGenre(bookId, genreId int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		insert into
			book_genre (book_id, genre_id)
		values
			($1, $2)
	`

	_, err := repo.DB.ExecContext(ctx, query, bookId, genreId)
	return err
}
