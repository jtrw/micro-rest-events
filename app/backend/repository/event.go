package repository

import (
    "database/sql"
    //"log"
    "errors"
   "fmt"
   "time"
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
     sql := `INSERT INTO "events"("uuid", "user_id", "type", "status") VALUES($1, $2, $3, $4)`
        _, err := repo.Connection.Exec(sql, e.Uuid, e.UserId, e.Type, e.Status)

     if err != nil {
        fmt.Println(err)
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

func (repo EventRepository) GetByUserId(userId int) (Event, error) {
    event := Event{}
    sql := `SELECT uuid,
                   user_id,
                   type,
                   status,
                   message,
                   is_seen
            FROM "events"
            WHERE user_id = $1 AND is_seen = false and status != 'new'
            ORDER BY created_at ASC LIMIT 1`

    row := repo.Connection.QueryRow(sql, userId)
    err := row.Scan(&event.Uuid, &event.UserId, &event.Type, &event.Status, &event.Message, &event.IsSeen)

    if err != nil {
        return event, errors.New("Row Not Found")
    }

    return event, nil
}


func (repo EventRepository) ChangeStatus(uuid string, e Event) (int64, error) {
    sql := `UPDATE "events" SET status = $1, message = $2, updated_at = $3 WHERE uuid = $4`
    res, err := repo.Connection.Exec(sql, e.Status, e.Message, time.Now(), uuid)

    if err != nil {
        return 0, errors.New("Can't update row")
    }
    count, err := res.RowsAffected()

    if err != nil {
        return 0, err
    }

    return count, nil
}

func (repo EventRepository) ChangeIsSeen(uuid string) (int64, error) {
    sql := `UPDATE "events" SET is_seen = true, updated_at = $1 WHERE uuid = $2`
    res, err := repo.Connection.Exec(sql, time.Now(), uuid)

    if err != nil {
        return 0, errors.New("Can't update row")
    }
    count, err := res.RowsAffected()

    if err != nil {
        return 0, err
    }

    return count, nil
}
