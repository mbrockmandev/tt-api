package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mbrockmandev/tometracker/internal/models"
)

func (p *PostgresDBRepo) CreateBook(book *models.Book, libraryId int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var publishedAt time.Time
	if book.PublishedAt.IsZero() {
		publishedAt = time.Now()
	}

	stmt := `
		insert into
			books
				(title, author, isbn, published_at, summary, thumbnail)
			values
				($1, $2, $3, $4, $5, $6)
		returning id
	`

	var newId int
	err := p.DB.QueryRowContext(ctx, stmt,
		book.Title,
		book.Author,
		book.ISBN,
		publishedAt,
		book.Summary,
		book.Thumbnail,
	).Scan(&newId)
	if err != nil {
		return 0, fmt.Errorf(
			"failed to scan new book in: %v -- title: %v, isbn: %v",
			err,
			book.Title,
			book.ISBN,
		)
	}

	// default library association if not passed in (using sentinel value, 0)
	if libraryId == 0 {
		stmt = `
			select
				id
			from
				libraries
			limit
				1
		`

		err = p.DB.QueryRowContext(ctx, stmt).Scan(&libraryId)
		if err != nil {
			return 0, fmt.Errorf("failed to get library Id: %v", err)
		}
	}

	checkStmt := `
		select
			count(*)
		from
			books_libraries
		where
			book_id = $1 and library_id = $2
	`

	var count int
	err = p.DB.QueryRowContext(ctx, checkStmt, newId, libraryId).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get association between book and library: %v", err)
	}

	if count == 0 {
		insertStmt := `
			insert into
				books_libraries
					(book_id, library_id, total_copies, available_copies)
				values
					($1, $2, 1, 1)
		`

		_, err = p.DB.ExecContext(ctx, insertStmt, newId, libraryId)
		if err != nil {
			return 0, fmt.Errorf("failed to insert into books_libraries: %v", err)
		}
	} else {
		updateStmt := `
			update
				books_libraries
			set
				total_copies = total_copies + 1,
				available_copies = available_copies + 1
			where
				book_id = $1 and library_id = $2
		`

		_, err = p.DB.ExecContext(ctx, updateStmt, newId, libraryId)
		if err != nil {
			return 0, fmt.Errorf("failed to update books_libraries: %v", err)
		}
	}

	return newId, nil
}

func (p *PostgresDBRepo) GetAllBooks(genre ...int) ([]*models.Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	where := ""
	if len(genre) > 0 {
		where = fmt.Sprintf(
			"where id in (select book_id from books_genres where genre_id = %d)",
			genre[0],
		)
	}

	query := fmt.Sprintf(`
		select
		  id, title, author, isbn, published_at, summary, thumbnail
		from
		  books %s
		order by
			title
	`, where)

	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []*models.Book

	for rows.Next() {
		book := &models.Book{}
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
			return nil, err
		}

		books = append(books, book)
	}
	return books, nil
}

func (p *PostgresDBRepo) GetBookByIsbn(isbn string) (*models.Book, *models.BookMetadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select
			id, title, author, isbn, published_at, summary, thumbnail
		from
			books
		where
			isbn = $1
	`

	row := p.DB.QueryRowContext(ctx, query, isbn)
	book := &models.Book{}
	err := row.Scan(
		&book.ID,
		&book.Title,
		&book.Author,
		&book.ISBN,
		&book.PublishedAt,
		&book.Summary,
		&book.Thumbnail,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan book with isbn %v", isbn)
	}

	query = `
	select
		book_id, total_copies, available_copies, borrowed_copies
	from
		books_libraries
	where
		book_id = $1
`

	bookMetadata := &models.BookMetadata{}
	row = p.DB.QueryRowContext(ctx, query, book.ID)
	err = row.Scan(
		&bookMetadata.BookId,
		&bookMetadata.TotalCopies,
		&bookMetadata.AvailableCopies,
		&bookMetadata.BorrowedCopies,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan book metadata with isbn %v", isbn)
	}

	return book, bookMetadata, nil
}

func (p *PostgresDBRepo) GetBookById(id int) (*models.Book,
	*models.BookMetadata, error,
) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select
			id, title, author, isbn, published_at, summary, thumbnail
		from
			books
		where
			id = $1
	`

	row := p.DB.QueryRowContext(ctx, query, id)

	book := &models.Book{}
	err := row.Scan(
		&book.ID,
		&book.Title,
		&book.Author,
		&book.ISBN,
		&book.PublishedAt,
		&book.Summary,
		&book.Thumbnail,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan book with id %v", id)
	}

	query = `
		select
		  book_id, total_copies, available_copies, borrowed_copies
		from
		  books_libraries
		where
		  book_id = $1
  `

	bookMetadata := &models.BookMetadata{}
	row = p.DB.QueryRowContext(ctx, query, id)
	err = row.Scan(
		&bookMetadata.BookId,
		&bookMetadata.TotalCopies,
		&bookMetadata.AvailableCopies,
		&bookMetadata.BorrowedCopies,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan book metadata with id %v", id)
	}

	return book, bookMetadata, nil
}

func (p *PostgresDBRepo) GetBookByIdForLibrary(id, libraryId int) (*models.Book,
	*models.BookMetadata, error,
) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select
			id, title, author, isbn, published_at, summary, thumbnail
		from
			books
		where
			id = $1
	`

	row := p.DB.QueryRowContext(ctx, query, id)

	book := &models.Book{}
	err := row.Scan(
		&book.ID,
		&book.Title,
		&book.Author,
		&book.ISBN,
		&book.PublishedAt,
		&book.Summary,
		&book.Thumbnail,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan book with id %v", id)
	}

	query = `
		select
		  book_id, total_copies, available_copies, borrowed_copies
		from
		  books_libraries
		where
		  book_id = $1 and library_id = $2
  `

	bookMetadata := &models.BookMetadata{}
	row = p.DB.QueryRowContext(ctx, query, id, libraryId)
	err = row.Scan(
		&bookMetadata.BookId,
		&bookMetadata.TotalCopies,
		&bookMetadata.AvailableCopies,
		&bookMetadata.BorrowedCopies,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to scan book metadata with id %v for library_id %v",
			id,
			libraryId,
		)
	}

	return book, bookMetadata, nil
}

func (p *PostgresDBRepo) GetBookByIsbnForLibrary(
	isbn string,
	libraryId int,
) (*models.Book, *models.BookMetadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select
			id, title, author, isbn, published_at, summary, thumbnail
		from
			books
		where
			isbn = $1
	`

	row := p.DB.QueryRowContext(ctx, query, isbn)

	book := &models.Book{}
	err := row.Scan(
		&book.ID,
		&book.Title,
		&book.Author,
		&book.ISBN,
		&book.PublishedAt,
		&book.Summary,
		&book.Thumbnail,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to scan book with isbn %v", isbn)
	}

	query = `
		select
		  book_id, total_copies, available_copies, borrowed_copies
		from
		  books_libraries
		where
		  book_id = $1 and library_id = $2
  `

	bookMetadata := &models.BookMetadata{}
	row = p.DB.QueryRowContext(ctx, query, book.ID, libraryId)
	err = row.Scan(
		&bookMetadata.BookId,
		&bookMetadata.TotalCopies,
		&bookMetadata.AvailableCopies,
		&bookMetadata.BorrowedCopies,
	)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"failed to scan book metadata with isbn %v for library_id %v",
			isbn,
			libraryId,
		)
	}

	return book, bookMetadata, nil
}

func (p *PostgresDBRepo) GetPopularBooks(limit, offset int) ([]*models.Book, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select
		  id, title, author, isbn, published_at, summary, thumbnail
		from
		  books
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

	var books []*models.Book
	var total int

	for rows.Next() {
		book := &models.Book{}
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
			return nil, 0, err
		}

		books = append(books, book)
	}

	err = p.DB.QueryRow(`select count(*) from books`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

func (p *PostgresDBRepo) GetBooksByQuery(
	queryText, queryType string,
	limit, offset int,
) ([]*models.Book, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	queryCol := ""
	switch queryType {
	case "title":
		queryCol = "title"
	case "author":
		queryCol = "author"
	case "isbn":
		queryCol = "isbn"
	default:
		return nil, 0, fmt.Errorf("invalid query type: %s", queryType)
	}

	searchQuery := "%" + queryText + "%"

	stmt := fmt.Sprintf(`
		select
		  id, title, author, isbn, published_at, summary, thumbnail
		from
		  books
		where
			%s ilike '%s'
		order by
			title
    limit
			%v
		offset
			%v
	`, queryCol, searchQuery, limit, offset)

	rows, err := p.DB.QueryContext(ctx, stmt)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var books []*models.Book
	var total int

	for rows.Next() {
		book := &models.Book{}
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
			fmt.Println("here we are? 1")
			return nil, 0, err
		}

		books = append(books, book)
	}

	err = p.DB.QueryRow(`select count(*) from books`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	return books, total, nil
}

func (p *PostgresDBRepo) GetBooksByAuthor(author string) ([]*models.Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `
		select
			*
		from
			books
		where
			author ilike $1
		order by
			title
	`

	searchQuery := "%" + author + "%"

	rows, err := p.DB.QueryContext(ctx, stmt, searchQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []*models.Book

	for rows.Next() {
		book := &models.Book{}
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
			return nil, err
		}

		books = append(books, book)
	}
	return books, nil
}

func (p *PostgresDBRepo) BorrowBook(userId, bookId, libraryId int) (*models.BookMetadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `
		select
			count(*)
		from
			books_libraries
		where
			book_id = $1 and library_id = $2
	`

	var count int
	err := p.DB.QueryRowContext(ctx, stmt, bookId, libraryId).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count < 1 {
		return nil, fmt.Errorf("book is not available at this library")
	}

	stmt = `
	select
		count(*)
	from
		users_books
	where
		user_id = $1 and book_id = $2 and returned_at is null
	`

	err = p.DB.QueryRowContext(ctx, stmt, userId, bookId).Scan(&count)
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, fmt.Errorf(
			"user has already borrowed book with id %v and this has not been returned yet",
			bookId,
		)
	}

	stmt = `
		select
			available_copies
		from
			books_libraries
		where
			id = $1
	`

	var availableCopies int
	err = p.DB.QueryRowContext(ctx, stmt, bookId).Scan(&availableCopies)
	if err != nil {
		return nil, err
	}

	if availableCopies < 1 {
		return nil, fmt.Errorf("no available copies of this book at this library for borrowing")
	}

	stmt = `
		update
			books_libraries
		set
			available_copies = available_copies - 1
		where
			book_id = $1 and library_id = $2
	`

	_, err = p.DB.ExecContext(ctx, stmt, bookId, libraryId)
	if err != nil {
		return nil, err
	}

	stmt = `
		update
			books_libraries
		set
			borrowed_copies = borrowed_copies + 1
		where
			book_id = $1 and library_id = $2
	`

	_, err = p.DB.ExecContext(ctx, stmt, bookId, libraryId)
	if err != nil {
		return nil, err
	}

	stmt = `
		insert into
			users_books
				(user_id, book_id, due_date, borrowed_at)
			values
				($1, $2, $3, $4)
			returning id
	`

	var borrowId int
	dueDate := time.Now().AddDate(0, 0, 14)

	err = p.DB.QueryRowContext(ctx, stmt, userId, bookId, dueDate, time.Now()).Scan(&borrowId)
	if err != nil {
		return nil, err
	}

	stmt = `
		select
		  book_id, total_copies, available_copies, borrowed_copies
		from
		  books_libraries
		where
		  book_id = $1
  `

	bookMetadata := &models.BookMetadata{}
	row := p.DB.QueryRowContext(ctx, stmt, bookId)
	err = row.Scan(
		&bookMetadata.BookId,
		&bookMetadata.TotalCopies,
		&bookMetadata.AvailableCopies,
		&bookMetadata.BorrowedCopies,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan book metadata with id %v", bookId)
	}

	return bookMetadata, nil
}

func (p *PostgresDBRepo) ReturnBook(userId, bookId, libraryId int) (*models.BookMetadata, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	checkStmt := `
	select
		id
	from
		users_books
	where
		book_id = $1 and user_id = $2 and returned_at is null
	`

	var borrowId int
	err := p.DB.QueryRowContext(ctx, checkStmt, bookId, userId).Scan(&borrowId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("the user hasn't borrowed this book or has already returned it")
		}
		return nil, err
	}

	// check if borrowed
	checkStmt = `
		select
			borrowed_copies
		from
			books_libraries
		where
			book_id = $1 and library_id = $2
	`

	var borrowedCopies int
	err = p.DB.QueryRowContext(ctx, checkStmt, bookId, libraryId).Scan(&borrowedCopies)
	if err != nil {
		return nil, err
	}
	if borrowedCopies < 1 {
		return nil, fmt.Errorf("no copies of this book at this library for returning")
	}

	returnStmt := `
		update
			users_books
		set
			returned_at = $1
		where
			id = $2
	`

	_, err = p.DB.ExecContext(ctx, returnStmt, time.Now(), borrowId)
	if err != nil {
		return nil, err
	}

	updateCountStmt := `
		update
			books_libraries
		set
			available_copies = available_copies + 1
		where
			book_id = $1 and library_id = $2
	`

	_, err = p.DB.ExecContext(ctx, updateCountStmt, bookId, libraryId)
	if err != nil {
		return nil, err
	}

	updateCountStmt = `
		update
			books_libraries
		set
			borrowed_copies = borrowed_copies - 1
		where
			book_id = $1 and library_id = $2
		`

	_, err = p.DB.ExecContext(ctx, updateCountStmt, bookId, libraryId)
	if err != nil {
		return nil, err
	}

	queryStmt := `
		select
		  book_id, total_copies, available_copies, borrowed_copies
		from
		  books_libraries
		where
		  book_id = $1
  `

	bookMetadata := &models.BookMetadata{}
	row := p.DB.QueryRowContext(ctx, queryStmt, bookId)
	err = row.Scan(
		&bookMetadata.BookId,
		&bookMetadata.TotalCopies,
		&bookMetadata.AvailableCopies,
		&bookMetadata.BorrowedCopies,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan book metadata with id %v", bookId)
	}

	return bookMetadata, nil
}

func buildUpdateBookQuery(id int, book *models.Book) (string, []interface{}) {
	const defaultUpdateFieldsCount = 3
	var setValues []string
	var args []interface{}
	var argId int = 1

	if book.Title != "" {
		setValues = append(setValues, fmt.Sprintf("title = $%d", argId))
		args = append(args, book.Title)
		argId++
	}
	if book.Author != "" {
		setValues = append(setValues, fmt.Sprintf("author = $%d", argId))
		args = append(args, book.Author)
		argId++
	}
	if book.ISBN != "" {
		setValues = append(setValues, fmt.Sprintf("isbn = $%d", argId))
		args = append(args, book.ISBN)
		argId++
	}
	if !book.PublishedAt.IsZero() {
		setValues = append(setValues, fmt.Sprintf("published_at = $%d", argId))
		args = append(args, book.PublishedAt)
		argId++
	} else {
		setValues = append(setValues, "published_at = now()")
	}
	if book.Summary != "" {
		setValues = append(setValues, fmt.Sprintf("summary = $%d", argId))
		args = append(args, book.Summary)
		argId++
	}
	if book.Thumbnail != "" {
		setValues = append(setValues, fmt.Sprintf("thumbnail = $%d", argId))
		args = append(args, book.Thumbnail)
		argId++
	}
	if book.Edition != "" {
		setValues = append(setValues, fmt.Sprintf("edition = $%d", argId))
		args = append(args, book.Edition)
		argId++
	}
	if !book.CreatedAt.IsZero() {
		setValues = append(setValues, fmt.Sprintf("created_at = $%d", argId))
		args = append(args, book.CreatedAt)
		argId++
	} else {
		setValues = append(setValues, "created_at = now()")
	}
	setValues = append(setValues, "updated_at = now()")

	if len(setValues) <= defaultUpdateFieldsCount {
		return "", nil
	}

	query := fmt.Sprintf("update books set %s where id = $%d", strings.Join(setValues, ", "), argId)
	args = append(args, id)

	return query, args
}

func (p *PostgresDBRepo) UpdateBook(id int, book *models.Book) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	_, _, err := p.GetBookById(id)
	if err != nil {
		return fmt.Errorf("no book found with id %v", id)
	}

	query, args := buildUpdateBookQuery(id, book)
	if query == "" {
		return errors.New("no columns provided for update")
	}
	_, err = p.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update book: %v -- query: %v -- args: %v", err, query, args)
	}

	return nil
}

func (p *PostgresDBRepo) DeleteBook(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `
		delete from
			books
		where
		id = $1
	`

	_, err := p.DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return err
	}

	stmt = `
		delete from
			books_libraries
		where
			book_id = $1
	`

	_, err = p.DB.ExecContext(ctx, stmt, id)
	return err
}

func (p *PostgresDBRepo) ReportPopularBooks() ([]*models.Book, error) {
	query := `
		select
			b.id, b.title, b.author, b.isbn, b.published_at, b.summary, b.thumbnail, count(ub.book_id) as borrow_count
		from
			users_books ub
		join
			books b on ub.book_id = b.id
		where
			ub.borrowed_at >= $1
		group by
			b.id
		order by
			borrow_count desc
		limit
			10;
	`

	oneMonthAgo := time.Now().AddDate(0, -1, 0)

	rows, err := p.DB.Query(query, oneMonthAgo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []*models.Book
	var borrowCount int
	for rows.Next() {
		var book models.Book
		if err := rows.Scan(
			&book.ID,
			&book.Title,
			&book.Author,
			&book.ISBN,
			&book.PublishedAt,
			&book.Summary,
			&book.Thumbnail,
			&borrowCount,
		); err != nil {
			return nil, err
		}
		books = append(books, &book)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return books, nil
}
