package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mbrockmandev/tometracker/internal/models"
)

// Genre Methods
func (p *PostgresDBRepo) CreateGenre(genre *models.Genre) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		insert into
			genres
				(name)
			values
				($1)
		returning id
	`

	row := p.DB.QueryRowContext(ctx, query, genre.Name)
	err := row.Scan(&genre.ID)

	return genre.ID, err
}

func (p *PostgresDBRepo) GetAllGenres() ([]*models.Genre, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select
			id, name
		from
			genres
		order by
			name
	`

	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var genres []*models.Genre
	for rows.Next() {
		var genre models.Genre
		if err := rows.Scan(&genre.ID, &genre.Name); err != nil {
			return nil, err
		}
		genres = append(genres, &genre)
	}
	return genres, nil
}

func (p *PostgresDBRepo) GetGenreById(id int) (*models.Genre, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select
			id, name
		from
			genres
		where
			id = $1
	`

	row := p.DB.QueryRowContext(ctx, query, id)
	var genre models.Genre
	if err := row.Scan(&genre.ID, &genre.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no genre found with id %d", id)
		}
		return nil, err
	}
	return &genre, nil
}

func (p *PostgresDBRepo) GetGenreByName(name string) (*models.Genre, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select
			id, name
		from
			genres
		where
			name = $1
	`

	row := p.DB.QueryRowContext(ctx, query, name)
	var genre models.Genre
	if err := row.Scan(&genre.ID, &genre.Name); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no genre found with name %s", name)
		}
		return nil, err
	}
	return &genre, nil
}

func (p *PostgresDBRepo) DeleteGenre(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		delete from
			genres
		where
			id = $1
	`

	_, err := p.DB.ExecContext(ctx, query, id)
	return err
}

func (p *PostgresDBRepo) UpdateGenre(id int, genre *models.Genre) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		update
			genres
		set
			name = $1
		where
			id = $2
	`

	_, err := p.DB.ExecContext(ctx, query, genre.Name, id)
	return err
}

func (p *PostgresDBRepo) ReportPopularGenres() ([]*models.Genre, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select
			g.id, g.name, count(ub.book_id) as borrow_count
		from
			users_books ub
		join
			books b on ub.book_id = b.id
		join
			books_genres bg on b.id = bg.book_id
		join
			genres g on bg.genre_id = g.id
		where
			ub.borrowed_at >= $1
		group by
			g.id
		order by
			borrow_count desc
		limit
			10;
	`

	rows, err := p.DB.QueryContext(ctx, query, time.Now().AddDate(0, -1, 0))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var genres []*models.Genre
	for rows.Next() {
		var genre models.Genre
		var borrowCount int
		if err := rows.Scan(
			&genre.ID,
			&genre.Name,
			&borrowCount,
		); err != nil {
			return nil, err
		}
		genres = append(genres, &genre)
	}

	return genres, nil
}
