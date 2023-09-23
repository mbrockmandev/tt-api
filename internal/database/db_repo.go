package database

import (
	"database/sql"
	"time"
)

type PostgresDBRepo struct {
	DB        *sql.DB
	DbTimeout time.Duration
}

func NewPostgresDBRepo() PostgresDBRepo {
	t := PostgresDBRepo{}
	t.DbTimeout = time.Second * 3
	return t
}

func (m *PostgresDBRepo) Connection() *sql.DB {
	return m.DB
}
