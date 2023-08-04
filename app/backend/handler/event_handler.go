package handler

import (
    "io"
    "fmt"
    "net/http"
    "database/sql"
    //"strconv"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/render"
    event "micro-rest-events/v1/app/backend/repository"
    "encoding/json"
    "github.com/google/uuid"
)

type JSON map[string]interface{}

type Handler struct {
    Connection *sql.DB
}

func NewHandler(conn *sql.DB) Handler {
    return Handler{Connection: conn}
}



func (h Handler) OnGetEventsByUserId(w http.ResponseWriter, r *http.Request) {
    uuid := chi.URLParam(r, "id")

    eventRepository := event.NewEventRepository(h.Connection)
    row, err := eventRepository.GetOne(uuid)

    if err != nil {
         render.Status(r, http.StatusNotFound)
         render.JSON(w, r, JSON{"status": "error"})
         return
    }

    render.JSON(w, r, JSON{"status": "ok", "data": row})
}


func (h Handler) OnCreateEvent(w http.ResponseWriter, r *http.Request) {
    var requestData JSON

    b, err := io.ReadAll(r.Body)
    if err != nil {
        fmt.Printf("[ERROR] %s", err)
    }

    err = json.Unmarshal(b, &requestData)

     if err != nil {
        fmt.Println("Error while decoding the data", err.Error())
     }
     uuid := uuid.New().String()

     rec := event.Event{
        Uuid: uuid,
        UserId: requestData["user_id"].(int),
        Type: requestData["type"].(string),
        Status: requestData["status"].(string),
        Message: requestData["message"].(string),
        IsSeen: requestData["is_seen"].(bool),
     }
     eventRepository := event.NewEventRepository(h.Connection)
     err = eventRepository.Create(rec)
     if err != nil {
        render.Status(r, http.StatusBadRequest)
        render.JSON(w, r, JSON{"status": "error", "message": err})
     }

     render.Status(r, http.StatusCreated)
     render.JSON(w, r, JSON{"status": "ok", "uuid": uuid})
}

func (h Handler) OnChangeEvent(w http.ResponseWriter, r *http.Request) {
    uuid := chi.URLParam(r, "uuid")

    var requestData JSON

    b, err := io.ReadAll(r.Body)
    if err != nil {
        fmt.Printf("[ERROR] %s", err)
    }

    err = json.Unmarshal(b, &requestData)

     if err != nil {
        fmt.Println("Error while decoding the data", err.Error())
     }

     rec := event.Event{
        Status: requestData["status"].(string),
        Message: requestData["message"].(string),
        IsSeen: requestData["is_seen"].(bool),
     }
     eventRepository := event.NewEventRepository(h.Connection)
     _, err = eventRepository.Change(uuid, rec)
     if err != nil {
        render.Status(r, http.StatusBadRequest)
        render.JSON(w, r, JSON{"status": "error", "message": err})
     }

     render.Status(r, http.StatusOK)
     render.JSON(w, r, JSON{"status": "ok", "uuid": uuid})
}

