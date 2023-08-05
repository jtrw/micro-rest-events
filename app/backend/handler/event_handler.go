package handler

import (
    "io"
    "fmt"
    "net/http"
    "database/sql"
    "strconv"
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/render"
    event "micro-rest-events/v1/app/backend/repository"
    "encoding/json"
    "github.com/google/uuid"
)

const STATUS_NEW = "new"

type JSON map[string]interface{}

type Handler struct {
    Connection *sql.DB
}

func NewHandler(conn *sql.DB) Handler {
    return Handler{Connection: conn}
}


func (h Handler) OnGetEventsByUserId(w http.ResponseWriter, r *http.Request) {
    userId, _ := strconv.Atoi(chi.URLParam(r, "id"))

    eventRepository := event.NewEventRepository(h.Connection)
    row, err := eventRepository.GetByUserId(userId)

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
        render.Status(r, http.StatusNotFound)
        return
    }

    err = json.Unmarshal(b, &requestData)

     if err != nil {
        fmt.Println("Error while decoding the data", err.Error())
        return
     }
     uuid := uuid.New().String()

     userId := int(requestData["user_id"].(float64))
     rec := event.Event{
        Uuid: uuid,
        UserId: userId,
        Type: STATUS_NEW,
        Status: requestData["status"].(string),
     }

     eventRepository := event.NewEventRepository(h.Connection)
     err = eventRepository.Create(rec)
     if err != nil {
        render.Status(r, http.StatusBadRequest)
        render.JSON(w, r, JSON{"status": "error", "message": err})
        return
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
     var message string = ""
     if requestData["message"] != nil {
        message = requestData["message"].(string)
     }

     rec := event.Event{
        Status: requestData["status"].(string),
        Message: message,
     }
     eventRepository := event.NewEventRepository(h.Connection)
     count, err := eventRepository.Change(uuid, rec)

     if count == 0 {
       // Check if the row not exists
        render.Status(r, http.StatusNotFound)
        render.JSON(w, r, JSON{"status": "error", "message": "Not Found"})
        return
     }

     if err != nil {
        render.Status(r, http.StatusBadRequest)
        render.JSON(w, r, JSON{"status": "error", "message": err})
     }

     render.Status(r, http.StatusOK)
     render.JSON(w, r, JSON{"status": "ok", "uuid": uuid})
}

