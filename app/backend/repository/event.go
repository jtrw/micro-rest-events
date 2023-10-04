package repository

import (
    "database/sql"
    "errors"
    "time"
    "log"
    "strings"
)

type Query struct {
    Statuses []string
    Limit int
    Offset int
    DateFrom string
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
    CreatedAt string
    UpdatedAt string
}

func NewEventRepository(conn *sql.DB) EventRepositoryInterface {
	return &EventRepository{
		Connection: conn,
	}
}

func (repo EventRepository) Create(e Event) error {

    timeNow := time.Now().Format("2006-01-02 15:04:05")
     sql := `INSERT INTO "events"
        ("uuid", "user_id", "type", "status", "caption", "message", "created_at", "updated_at")
        VALUES($1, $2, $3, $4, $5, $6, $7, $8)`
     _, err := repo.Connection.Exec(sql, e.Uuid, e.UserId, e.Type, e.Status, e.Caption, e.Body, timeNow, timeNow)

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
                   is_seen,
                   created_at,
                   updated_at
            FROM "events"
            WHERE user_id = $1`

    if q.Statuses != nil {
        sql += ` AND status IN ('` + strings.Join(q.Statuses, `', '`) + `')`
    }

    if len(q.DateFrom) > 0 {
        sql += ` AND updated_at >= '` + q.DateFrom + `'`
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
        err := rows.Scan(&event.Uuid,
            &event.UserId,
            &event.Type,
            &event.Status,
            &event.Caption,
            &event.Message,
            &event.IsSeen,
            &event.CreatedAt,
            &event.UpdatedAt)
        if err != nil {
            return nil, err
        }
        events = append(events, event)
    }
    if err = rows.Err(); err != nil {
        return nil, err
    }

    return events, nil
}


func (repo EventRepository) ChangeStatus(uuid string, e Event) (int64, error) {
    sql := `UPDATE "events" SET status = $1, updated_at = $2`


    if len(e.Message) > 0 {
        sql += `, message = '`+e.Message+`'`
    }

    sql += ` WHERE uuid = $3`

    res, err := repo.Connection.Exec(sql, e.Status, time.Now(), uuid)

    if err != nil {
        log.Println("[ERROR] Error while updating the event", err.Error())
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
