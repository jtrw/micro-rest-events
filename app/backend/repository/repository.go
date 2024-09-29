package repository

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

type StoreProvider struct {
	db     *sql.DB
	dbType string
}

type StoreProviderInterface interface {
	Create(e Event) error
	GetOne(uuid string) (Event, error)
	GetOneByUserId(userId string) (Event, error)
	GetAllByUserId(userId string, q Query) ([]Event, error)
	ChangeStatus(uuid string, e Event) (int64, error)
	ChangeIsSeen(uuid string) (int64, error)
}

func NewStoreProvider(conn string) (StoreProviderInterface, error) {
	dbType := func(c string) (string, error) {
		if strings.HasPrefix(c, "postgres://") {
			return "postgres", nil
		}
		if strings.Contains(c, "@tcp(") {
			return "mysql", nil
		}
		if strings.HasPrefix(c, "file:/") || strings.HasSuffix(c, ".sqlite") || strings.HasSuffix(c, ".db") {
			return "sqlite", nil
		}
		return "", fmt.Errorf("unsupported database type in connection string")
	}

	dbt, err := dbType(conn)
	if err != nil {
		return nil, fmt.Errorf("can't determine database type: %w", err)
	}

	db, err := sql.Open(dbt, conn)
	if err != nil {
		return nil, fmt.Errorf("error opening secrets database: %w", err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS events (
	    id INTEGER PRIMARY KEY AUTOINCREMENT,
	    uuid TEXT NOT NULL,
	    user_id TEXT NULL,
	    type TEXT NULL,
	    status TEXT NULL,
	    caption TEXT NULL,
	    message TEXT NULL,
	    is_seen BOOLEAN DEFAULT 0,
	    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	    updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		return nil, err
	}
	log.Printf("[INFO] secrets provider: using %s database, type: %s", conn, dbt)
	return &StoreProvider{db: db, dbType: dbt}, nil
}
