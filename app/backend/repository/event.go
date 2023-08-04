package repository

import (
    "database/sql"
    "log"
    "errors"
  // "fmt"
   "strconv"
)

type EventRepository struct {
    Connection  *sql.DB
}


type Event struct {
    Uuid string
    UserId int
	Type string
	Status string
    Message string
    IsSeen bool
}

func NewEventRepository(conn *sql.DB) *EventRepository {
	return &EventRepository{
		Connection: conn,
	}
}

func (repo EventRepository) Create(e Event) error {
     sql := `INSERT INTO "events"("uuid", "name", "description", "text") VALUES($1, $2, $3, $4)`
     _, err := repo.Connection.Exec(sql, e.Uuid, e.Name, e.Description, e.Text)
     if err != nil {
        return errors.New("Couldn't create event")
     }

     return nil
}

func (repo EventRepository) GetOne(uuid string) (Event, error) {
    event := Event{}

    sql := `SELECT uuid, name, description FROM "events" WHERE uuid = $1`
    row := repo.Connection.QueryRow(sql, uuid)

    err := row.Scan(&event.Uuid, &event.Name, &event.Description)

    if err != nil {
        return event, errors.New("Row Not Found")
    }

    return event, nil
}


func (repo EventRepository) Change(uuid string, e Event) (int64, error) {
    sql := `UPDATE "events" SET name = $1, description = $2, text = $3 WHERE uuid = $4`
    res, err := repo.Connection.Exec(sql, e.Name, e.Description, e.Text, uuid)
    if err != nil {
        return 0, errors.New("Can't update row")
    }
    count, err := res.RowsAffected()

    if err != nil {
        return 0, err
    }

    return count, nil
}
