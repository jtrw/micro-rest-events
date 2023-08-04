package repository

import (
    "database/sql"
    //"log"
    "errors"
  // "fmt"
  // "strconv"
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
     sql := `INSERT INTO "events"("uuid", "user_id", "type", "status", "message", "is_seen") VALUES($1, $2, $3, $4, $5, $6)`
        _, err := repo.Connection.Exec(sql, e.Uuid, e.UserId, e.Type, e.Status, e.Message, e.IsSeen)

     if err != nil {
        return errors.New("Couldn't create event")
     }

     return nil
}

func (repo EventRepository) GetOne(uuid string) (Event, error) {
    event := Event{}
    sql := `SELECT uuid, user_id, type, status, message, is_seen FROM "events" WHERE uuid = $1`
    row := repo.Connection.QueryRow(sql, uuid)
    err := row.Scan(&event.Uuid, &event.UserId, &event.Type, &event.Status, &event.Message, &event.IsSeen)

    if err != nil {
        return event, errors.New("Row Not Found")
    }

    return event, nil
}


func (repo EventRepository) Change(uuid string, e Event) (int64, error) {
    sql := `UPDATE "events" SET user_id = $1, type = $2, status = $3, message = $4, is_seen = $5 WHERE uuid = $6`
    res, err := repo.Connection.Exec(sql, e.UserId, e.Type, e.Status, e.Message, e.IsSeen, uuid)

    if err != nil {
        return 0, errors.New("Can't update row")
    }
    count, err := res.RowsAffected()

    if err != nil {
        return 0, err
    }

    return count, nil
}
