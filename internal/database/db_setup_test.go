package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func execSqlScript(db *sql.DB, path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("Failed to read file %s: %v", path, err))
	}

	commands := strings.Split(string(content), ";")
	for _, command := range commands {
		if strings.TrimSpace(command) == "" {
			continue
		}

		_, err := db.Exec(command)
		if err != nil {
			panic(fmt.Sprintf("Failed to execute command %s: %v", command, err))
		}
	}
}

func setupTestDB() *PostgresDBRepo {
	err := godotenv.Load("../../.env")
	if err != nil {
		panic(err)
	}
	dbHost := os.Getenv("DB_ENDPOINT_DEBUG")

	dsn := fmt.Sprintf(
		"host=%s port=5432 dbname=library_db_test",
		dbHost,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(err)
	}

	execSqlScript(db, "../../../../scripts/create_tables.sql")
	execSqlScript(db, "../../../../scripts/insert_test_data_into_tables.sql")

	repo := &PostgresDBRepo{DB: db}
	return repo
}

func teardownTestDB(repo *PostgresDBRepo) {
	tables := []string{
		"activity_logs",
		"books_ratings",
		"users_genres",
		"books_genres",
		"books_libraries",
		"users_books",
		"users_libraries",
		"genres",
		"books",
		"users",
		"libraries",
	}

	for _, table := range tables {
		query := fmt.Sprintf("DROP TABLE IF EXISTS %s;", table)
		_, err := repo.DB.Exec(query)
		if err != nil {
			fmt.Printf("Failed to drop table %s: %v\n", table, err)
		}
	}

	repo.DB.Close()
}
