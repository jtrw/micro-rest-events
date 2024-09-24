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

func NewStoreProvider(conn string) (*StoreProvider, error) {
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
		    id INT PRIMARY KEY,
		    uuid varchar(50) NOT NULL,
		    user_id varchar(50) NULL,
		    "type" varchar(50) NULL,
		    status varchar(50) NULL,
		    caption varchar(155) NULL,
		    message text NULL,
		    is_seen bool DEFAULT false,
		    created_at timestamp(0) NOT NULL DEFAULT now(),
		    updated_at timestamp(0) NULL DEFAULT now()
		);`)
	if err != nil {
		return nil, err
	}
	log.Printf("[INFO] secrets provider: using %s database, type: %s", conn, dbt)
	return &StoreProvider{db: db, dbType: dbt}, nil
}
