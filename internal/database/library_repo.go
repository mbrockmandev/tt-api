package database

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/lib/pq"
	"github.com/mbrockmandev/tometracker/internal/models"
)

func (p *PostgresDBRepo) CreateLibrary(library *models.Library) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `
		insert into
			libraries
				(name, city, street_address, postal_code, country, phone)
			values
				($1, $2, $3, $4, $5, $6)
			returning id
	`

	var newId int
	err := p.DB.QueryRowContext(ctx, stmt,
		library.Name,
		library.City,
		library.StreetAddress,
		library.PostalCode,
		library.Country,
		library.Phone,
	).Scan(&newId)
	if err != nil {
		return 0, fmt.Errorf("failed to create library: %v", err)
	}

	go p.populateBooksForLibrary(ctx, newId)

	return newId, nil
}

func (p *PostgresDBRepo) GetLibraryByName(name string) (*models.Library, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select
			id, name, city, street_address, postal_code, country, phone
		from
			libraries
		where
			name ilike $1
	`

	searchName := "%" + name + "%"

	row := p.DB.QueryRowContext(ctx, query, searchName)
	var library models.Library

	err := row.Scan(
		&library.ID,
		&library.Name,
		&library.City,
		&library.StreetAddress,
		&library.PostalCode,
		&library.Country,
		&library.Phone,
	)
	if err != nil {
		return nil, err
	}

	return &library, nil
}

func (p *PostgresDBRepo) GetAllLibraries(limit, offset int) ([]*models.Library, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select
			id, name, city, street_address, postal_code, country, phone
		from
			libraries
		order by
			name
		limit
			$1
		offset
			$2
	`

	rows, err := p.DB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	var libraries []*models.Library
	var count int
	for rows.Next() {
		var library models.Library
		err := rows.Scan(
			&library.ID,
			&library.Name,
			&library.City,
			&library.StreetAddress,
			&library.PostalCode,
			&library.Country,
			&library.Phone,
		)
		if err != nil {
			return nil, 0, err
		}
		err = p.DB.QueryRow(`select count(*) from libraries`).Scan(&count)
		if err != nil {
			return nil, 0, err
		}

		if count == 0 {
			return nil, 0, fmt.Errorf("unable to find any libraries")
		}

		libraries = append(libraries, &library)
	}

	return libraries, count, nil
}

func (p *PostgresDBRepo) GetLibraryById(id int) (*models.Library, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select
			id, name, city, street_address, postal_code, country, phone
		from
			libraries
		where
			id = $1
	`

	row := p.DB.QueryRowContext(ctx, query, id)
	var library models.Library
	err := row.Scan(
		&library.ID,
		&library.Name,
		&library.City,
		&library.StreetAddress,
		&library.PostalCode,
		&library.Country,
		&library.Phone,
	)
	if err != nil {
		return nil, err
	}

	return &library, nil
}

func (p *PostgresDBRepo) UpdateLibrary(id int, library *models.Library) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `
		update
			libraries
		set
			name = $1,
			city = $2,
			street_address = $3,
			postal_code = $4,
			country = $5,
			phone = $6
	`

	_, err := p.DB.ExecContext(ctx, stmt,
		library.Name,
		library.City,
		library.StreetAddress,
		library.PostalCode,
		library.Country,
		library.Phone,
	)

	return err
}

func (p *PostgresDBRepo) DeleteLibrary(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `
		delete from
			libraries
		where
			id = $1
	`

	_, err := p.DB.ExecContext(ctx, stmt, id)
	return err
}

func (p *PostgresDBRepo) GetBooksByLibrary(libraryId int) ([]*models.Book,
	[]*models.BookMetadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := fmt.Sprintf(`
		select
			book_id, total_copies, available_copies, borrowed_copies
		from
			books_libraries
		where
			library_id = %v
	`, libraryId)

	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve inventory info: %s", err)
	}
	defer rows.Close()

	var libraryInventory []*models.BookMetadata
	var bookIds []int
	for rows.Next() {
		var item models.BookMetadata
		err := rows.Scan(
			&item.BookId,
			&item.TotalCopies,
			&item.AvailableCopies,
			&item.BorrowedCopies,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan inventory info: %s", err)
		}
		libraryInventory = append(libraryInventory, &item)
		bookIds = append(bookIds, item.BookId)
	}

	query = `
		select
			id, title, author, isbn, published_at, summary, thumbnail
		from
			books
		where
			id = any($1)
	`

	rows, err = p.DB.QueryContext(ctx, query, pq.Array(bookIds))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve books for libraryId %v: %s", libraryId, err)
	}
	defer rows.Close()

	var books []*models.Book
	for rows.Next() {
		var book models.Book
		err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Author,
			&book.ISBN,
			&book.PublishedAt,
			&book.Summary,
			&book.Thumbnail,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan book: %s", err)
		}
		books = append(books, &book)
	}

	return books, libraryInventory, nil
}

func (p *PostgresDBRepo) populateBooksForLibrary(ctx context.Context, libraryId int) {
	rows, err := p.DB.QueryContext(ctx, "select id from books")
	if err != nil {
		log.Printf("failed to retrieve ids of books: %s", err)
		return
	}
	defer rows.Close()

	var bookIds []int
	for rows.Next() {
		var bookId int
		if err := rows.Scan(&bookId); err != nil {
			log.Printf("failed to scan book id: %s", err)
			return
		}
		bookIds = append(bookIds, bookId)
	}
	if err := rows.Err(); err != nil {
		log.Printf("failed to go through rows: %s", err)
		return
	}

	for _, bookId := range bookIds {
		totalCopies := rand.Intn(10) + 1
		borrowedCopies := 0
		availableCopies := totalCopies

		_, err = p.DB.ExecContext(ctx, `
			insert into
				books_libraries
					(book_id, library_id, total_copies, borrowed_copies, available_copies)
				values
					($1, $2, $3, $4, $5)
		`, bookId, libraryId, totalCopies, borrowedCopies, availableCopies)
		if err != nil {
			log.Printf("failed to insert book into library: %s bookId: %v -- \nlibraryId: %v", err, bookId, libraryId)
			return
		}
	}
}
