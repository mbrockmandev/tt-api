package app

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func (app *Application) ConnectDB() (*sql.DB, error) {
	defaultDsn := getDefaultDsn()
	db, err := sql.Open("pgx", defaultDsn)
	if err != nil {
		log.Println("Failed to connect to database")
		return nil, err
	}

	// check if database exists
	exists, err := checkDatabaseExists(db, "library_db")
	if err != nil {
		db.Close()
		log.Println("Failed to check if database exists")
		return nil, err
	}

	// create database
	if !exists {
		if err := createDatabase(db, "library_db"); err != nil {
			db.Close()
			log.Println("Failed to create database")
			return nil, err
		}
		log.Println("Database library_db created")
	}

	db.Close()

	// actual connection
	app.DSN = getDsn()
	db, err = sql.Open("pgx", app.DSN)
	if err != nil {
		return nil, err
	}

	log.Println("Connected to Library Database.")

	// create tables
	if err := runSQLScript(db, "./create_tables.sql"); err != nil {
		db.Close()
		log.Println("Failed to create tables")
		return nil, err
	}
	log.Println("Tables created")

	// insert book data, if needed
	err = batchInsertFromCsvToPostgres(db, "./books_table.csv", 1000)
	if err != nil {
		db.Close()
		log.Println("Failed to insert book data")
		return nil, err
	}
	log.Println("Book data inserted")

	return db, nil
}

func getDefaultDsn() string {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s database=postgres sslmode=require timezone=UTC connect_timeout=5", dbHost, dbPort, dbUsername, dbPassword)

	return dsn
}

func checkDatabaseExists(db *sql.DB, dbName string) (bool, error) {
	const query = `
			select datname
			from pg_database
			where datname = $1;
		`
	var name string
	if err := db.QueryRow(query, dbName).Scan(&name); err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func createDatabase(db *sql.DB, dbName string) error {
	query := fmt.Sprintf(`
		create database %s;
	`, dbName)
	_, err := db.Exec(query)

	return err
}

func runSQLScript(db *sql.DB, path string) error {
	content, err := os.ReadFile(path)

	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", path, err)
	}

	commands := strings.Split(string(content), ";")
	for i, command := range commands {
		if strings.TrimSpace(command) == "" {
			continue
		}

		_, err := db.Exec(command)
		if err != nil {
			return fmt.Errorf("failed to execute command %s: %v", command, err)
		}
		log.Println("Executed command No.", i)
		// log.Println("Executed command", command)
	}
	return nil
}

func batchInsertFromCsvToPostgres(db *sql.DB, path string, batchSize int) error {
	log.Println("Inserting data...")
	var count int
	err := db.QueryRow(`select count(*) from books;`).Scan(&count)
	if err != nil {
		log.Println(err)
		return err
	}

	// silently return, we don't want to insert more data
	if count > 1000 {
		log.Println("Data already exists, skipping insert...")
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		log.Println("failed to open csv file", err)
		return err
	}
	defer file.Close()
	r := csv.NewReader(file)
	records, err := r.ReadAll()
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Inserting", len(records), "records")
	for i := 1; i < len(records); i += batchSize {
		log.Println("Inserting batch", i, "to", i+batchSize)
		end := i + batchSize
		if end > len(records) {
			end = len(records)
		}

		batch := records[i:end]

		valueStrings := []string{}
		valueArgs := []interface{}{}

		for _, record := range batch {
			// invalid book data, skip
			if !validateBookEntry(record) {
				fmt.Printf("invalid record: %v\n", record)
				continue
			}

			requiredData := append(record[1:6], record[7:]...)

			placeholderOffset := len(valueArgs) + 1
			placeholder := []string{}
			for _, value := range requiredData {
				if value == "NULL" {
					placeholder = append(placeholder, "NULL")
				} else {
					placeholder = append(placeholder, fmt.Sprintf("$%d", placeholderOffset))
					valueArgs = append(valueArgs, value)
					placeholderOffset++
				}
			}
			valueStrings = append(valueStrings, "("+strings.Join(placeholder, ", ")+")")
		}

		stmt := fmt.Sprintf("insert into books (title, author, isbn, published_at, summary, thumbnail, created_at, updated_at) values %s;", strings.Join(valueStrings, ","))
		stmt = replacePlaceholders(stmt, len(valueArgs))

		log.Printf(stmt, valueArgs...)
		_, err = db.Exec(stmt, valueArgs...)
		if err != nil {
			log.Printf("failed to insert batch %d to %d: %v", i, end, err)
			return err
		}
		log.Println("inserted batch", i, "to", end)
	}

	return nil
}

func replacePlaceholders(stmt string, argCount int) string {
	counter := 1
	for i := 0; i < argCount; i++ {
		placeholder := fmt.Sprintf("$%d", counter)
		stmt = strings.Replace(stmt, "?", placeholder, 1)
		counter++
	}
	return stmt
}

func validateBookEntry(record []string) bool {
	if len(record) != 10 {
		return false
	}

	for _, field := range []int{1, 2, 3, 4, 8, 9} {
		if strings.TrimSpace(record[field]) == "" {
			return false
		}
	}

	timeFormat := "2006-01-02 15:04:05-07"
	for _, field := range []int{4, 8, 9} {
		_, err := time.Parse(timeFormat, record[field])
		if err != nil {
			return false
		}
	}

	return true
}

func getDsn() string {
	env := os.Getenv("ENV")
	var dbHost, dbPort, dbUsername, dbPassword, dbName string

	// debug mode (local postgres)
	if env == "debug" {
		dbHost = os.Getenv("DB_HOST_DEBUG")
		dbPort = os.Getenv("DB_PORT_DEBUG")
		dbUsername = os.Getenv("DB_USERNAME_DEBUG")
		dbPassword = os.Getenv("DB_PASSWORD_DEBUG")
		return fmt.Sprintf("host=%s port=%s user=%s password=%s database=%s sslmode=disable timezone=UTC connect_timeout=5", dbHost, dbPort, dbUsername, dbPassword, dbName)
	}

	// prod mode (AWS RDS Postgres)
	dbHost = os.Getenv("DB_HOST")
	dbPort = os.Getenv("DB_PORT")
	dbUsername = os.Getenv("DB_USERNAME")
	dbName = os.Getenv("DB_DBNAME")
	dbPassword = os.Getenv("DB_PASSWORD")
	return fmt.Sprintf("host=%s port=%s user=%s password=%s database=%s sslmode=require timezone=UTC connect_timeout=5", dbHost, dbPort, dbUsername, dbPassword, dbName)

}
