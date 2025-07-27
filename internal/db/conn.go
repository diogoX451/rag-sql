package db

import (
	"database/sql"
	"log"
	"rag-sql/internal/config"

	_ "github.com/lib/pq"
)

func Connect(cfg config.DatabaseConfig) *sql.DB {
	db, err := sql.Open("postgres", cfg.ConnString())
	if err != nil {
		log.Fatalf("erro ao abrir conexão com banco: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("erro ao pingar banco: %v", err)
	}

	return db
}
