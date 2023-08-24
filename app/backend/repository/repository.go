package repository

import (
    "database/sql"
    "os"
    "log"
    _ "github.com/lib/pq"
)

type Repository struct {
    Connection  *sql.DB
}

func ConnectDB() Repository {
    db, err := sql.Open("postgres", os.Getenv("POSTGRES_DSN"))
    if err != nil {
        log.Println(os.Getenv("POSTGRES_DSN"))
        panic(err)
    }
    rep := Repository{
        Connection: db,
    }
    return rep
}
