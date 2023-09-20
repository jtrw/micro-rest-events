package repository

import (
    "database/sql"
    "errors"
   //"fmt"
   "time"
   "log"
   "strings"
  // "strconv"
)

type Query struct {
    Statuses []string
}

type EventRepository struct {
    Connection  *sql.DB
}

type EventRepositoryInterface interface {
    Create(e Event) error
    GetOne(uuid string) (Event, error)
    GetOneByUserId(userId string) (Event, error)
    GetAllByUserId(userId string, q Query) ([]Event, error)
    ChangeStatus(uuid string, e Event) (int64, error)
    ChangeIsSeen(uuid string) (int64, error)
}

type Event struct {
    Uuid string
    UserId string
	Type string
	Status string
    Caption string
    Body string
    Message string
    IsSeen bool
}

func NewEventRepository(conn *sql.DB) EventRepositoryInterface {
	return &EventRepository{
		Connection: conn,
	}
}

func (repo EventRepository) Create(e Event) error {
     sql := `INSERT INTO "events"
        ("uuid", "user_id", "type", "status", "caption", "message")
        VALUES($1, $2, $3, $4, $5, $6)`
     _, err := repo.Connection.Exec(sql, e.Uuid, e.UserId, e.Type, e.Status, e.Caption, e.Body)

     if err != nil {
        log.Println("[ERROR] Error while creating the event", err.Error())
        return errors.New("Couldn't create event")
     }

     return nil
}

func (repo EventRepository) GetOne(uuid string) (Event, error) {
    event := Event{}
    sql := `SELECT uuid,
                   user_id,
                   type,
                   status,
                   caption,
                   message,
                   is_seen
            FROM "events" WHERE uuid = $1`
    row := repo.Connection.QueryRow(sql, uuid)
    err := row.Scan(&event.Uuid, &event.UserId, &event.Type, &event.Status, &event.Caption, &event.Message, &event.IsSeen)

    if err != nil {
        return event, errors.New("Row Not Found")
    }

    return event, nil
}

func (repo EventRepository) GetOneByUserId(userId string) (Event, error) {
    event := Event{}
    sql := `SELECT uuid,
                   user_id,
                   type,
                   status,
                   caption,
                   message,
                   is_seen
            FROM "events"
            WHERE user_id = $1 AND is_seen = false and status != 'new'
            ORDER BY created_at ASC LIMIT 1`

    row := repo.Connection.QueryRow(sql, userId)
    err := row.Scan(&event.Uuid, &event.UserId, &event.Type, &event.Status, &event.Caption, &event.Message, &event.IsSeen)

    if err != nil {
        return event, errors.New("Row Not Found")
    }

    return event, nil
}

func (repo EventRepository) GetAllByUserId(userId string, q Query) ([]Event, error) {
    sql := `SELECT uuid,
                   user_id,
                   type,
                   status,
                   caption,
                   message,
                   is_seen
            FROM "events"
            WHERE user_id = $1`

    if q.Statuses != nil {
        sql += ` AND status IN ('` + strings.Join(q.Statuses, `', '`) + `')`
    }

    sql += ` ORDER BY created_at ASC`

    rows, err := repo.Connection.Query(sql, userId)

    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var events []Event
    for rows.Next() {
        event := Event{}
        err := rows.Scan(&event.Uuid, &event.UserId, &event.Type, &event.Status, &event.Caption, &event.Message, &event.IsSeen)
        if err != nil {
            return nil, err
        }
        events = append(events, event)
    }
    if err = rows.Err(); err != nil {
        return nil, err
    }

//     if err != nil {
//         return event, errors.New("Row Not Found")
//     }

    return events, nil
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
