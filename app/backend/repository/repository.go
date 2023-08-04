package repository

import (
    "database/sql"
    "os"
)

type Repository struct {
    Connection  *sql.DB
}

func ConnectDB() Repository {
    db, err := sql.Open("postgres", os.Getenv("POSTGRES_DSN"))
    if err != nil {
        panic(err)
    }
    rep := Repository{
        Connection: db,
    }
    return rep
}
