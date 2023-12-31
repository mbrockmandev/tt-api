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

func (p *PostgresDBRepo) CreateUser(user *models.User) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	currentTime := time.Now()

	query := `
		insert into
			users
				(email, first_name, last_name, password, role, created_at, updated_at)
			values
				($1, $2, $3, $4, $5, $6, $7)
		returning id
	`

	var userId int
	err := p.DB.QueryRowContext(ctx, query,
		user.Email,
		user.FirstName,
		user.LastName,
		user.Password,
		user.Role,
		currentTime,
		currentTime,
	).Scan(&userId)

	if err != nil {
		return 0, err
	}

	query = fmt.Sprintf(`
		insert into
			users_libraries
				(user_id, library_id)
			values
				(%v, %v)
	`, userId, 1)

	_, err = p.DB.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}

	return userId, nil
}

func (p *PostgresDBRepo) GetAllUsers() ([]*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select
			*
		from
			users
		order by
			last_name
	`

	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*models.User{}
	for rows.Next() {
		user := &models.User{}
		err = rows.Scan(
			&user.ID,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Password,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (p *PostgresDBRepo) GetUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select
			id, email, first_name, last_name, password, role, created_at, updated_at
	 	from
	 		users
	 	where
			email = $1
	`

	user := &models.User{}
	row := p.DB.QueryRowContext(ctx, query, email)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no user found with email %s", email)
	} else if err != nil {
		return nil, err
	}

	return user, nil
}

func (p *PostgresDBRepo) GetUserRoleByEmail(email string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select
			role
		from
			users
		where
			email = $1
	`

	row := p.DB.QueryRowContext(ctx, query, email)

	var role string
	err := row.Scan(&role)
	if err != nil {
		return "", err
	}

	return role, nil
}

func (p *PostgresDBRepo) GetUserById(id int) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		select
			id, email, first_name, last_name, role, created_at, updated_at
		from
			users
		where
			id = $1
	`

	user := &models.User{}
	row := p.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (p *PostgresDBRepo) GetBorrowedBooksByUserId(userId int) ([]*models.Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `
		select
			b.id, b.title, b.author, b.isbn, b.published_at, b.summary, b.thumbnail
		from
			books b
		inner join
			users_books ub on b.id = ub.book_id
		where
			ub.user_id = $1 and returned_at is null
		order by
			title
	`

	rows, err := p.DB.QueryContext(ctx, stmt, userId)
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

func (p *PostgresDBRepo) GetUserHomeLibrary(userId int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := fmt.Sprintf(`
		select
			library_id
		from
			users_libraries
		where
			user_id = %v
	`, userId)

	row := p.DB.QueryRowContext(ctx, stmt)
	if row == nil {
		return 0, errors.New("unable to obtain row from database")
	}

	var libraryId int
	err := row.Scan(&libraryId)
	if err != nil {
		return 0, err
	}

	return libraryId, nil
}

func (p *PostgresDBRepo) GetRecentlyReturnedBooksByUserId(userId int) ([]*models.Book, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	stmt := `
		select
			b.id, b.title, b.author, b.isbn, b.published_at, b.summary, b.thumbnail
		from
			books b
		inner join
			users_books ub on b.id = ub.book_id
		where
			ub.user_id = $1
      and returned_at is not null
      and returned_at >= current_date - interval '14 days'
		order by
			title
	`

	rows, err := p.DB.QueryContext(ctx, stmt, userId)
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

func buildUpdateUserQuery(id int, user *models.User) (string, []interface{}) {
	var setValues []string
	var args []interface{}
	var argId int = 1

	if user.Email != "" {
		setValues = append(setValues, fmt.Sprintf("email = $%d", argId))
		args = append(args, user.Email)
		argId++
	}
	if user.FirstName != "" {
		setValues = append(setValues, fmt.Sprintf("first_name = $%d", argId))
		args = append(args, user.FirstName)
		argId++
	}
	if user.LastName != "" {
		setValues = append(setValues, fmt.Sprintf("last_name = $%d", argId))
		args = append(args, user.LastName)
		argId++
	}
	if user.Password != "" {
		setValues = append(setValues, fmt.Sprintf("password = $%d", argId))
		args = append(args, user.Password)
		argId++
	}
	if user.Role != "" {
		setValues = append(setValues, fmt.Sprintf("role = $%d", argId))
		args = append(args, user.Role)
		argId++
	}
	if !user.UpdatedAt.IsZero() {
		setValues = append(setValues, fmt.Sprintf("updated_at = $%d", argId))
		args = append(args, user.UpdatedAt)
		argId++
	}

	if len(setValues) == 0 {
		return "", nil
	}

	query := fmt.Sprintf("update users set %s where id = $%d", strings.Join(setValues, ", "), argId)
	args = append(args, id)
	return query, args
}

func (p *PostgresDBRepo) UpdateUser(id int, user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	_, err := p.GetUserById(id)
	if err != nil {
		return fmt.Errorf("no user found with id %v", id)
	}

	query, args := buildUpdateUserQuery(id, user)
	if query == "" {
		return errors.New("no columns provided for update")
	}

	_, err = p.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}

	return nil
}

func (p *PostgresDBRepo) DeleteUser(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	query := `
		delete from
			users_libraries
		where
			user_id = $1
	`

	_, err := p.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	query = `
		delete from
			users
		where
			id = $1
	`

	_, err = p.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	return nil
}

func (p *PostgresDBRepo) UpdateUserLibrary(libraryId, userId int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	checkQuery := `
		select
			1
		from
			users_libraries
		where
			user_id = $1 and library_id = $2
	`

	var exists int
	if err := p.DB.QueryRowContext(ctx, checkQuery, userId, libraryId).Scan(&exists); err != nil &&
		err != sql.ErrNoRows {
		return err
	}

	if exists == 1 {
		return errors.New("user already has this set as their library")
	}

	deleteQuery := `
		delete
		from
			users_libraries
		where
			user_id = $1
	`

	if _, err := p.DB.ExecContext(ctx, deleteQuery, userId); err != nil {
		return err
	}

	insertQuery := `
		insert into
		users_libraries
			(user_id, library_id)
		values
			($1, $2)
	`

	_, err := p.DB.ExecContext(ctx, insertQuery, userId, libraryId)
	if err != nil {
		return err
	}

	return nil
}
