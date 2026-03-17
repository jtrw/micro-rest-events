package web

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	event "micro-rest-events/internal/repository"
	repository "micro-rest-events/internal/repository"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

const STATUS_NEW = "new"

type JSON map[string]interface{}

type Handler struct {
	StoreProvider repository.StoreProviderInterface
}

func (h Handler) OnCreateEvent(w http.ResponseWriter, r *http.Request) {
	var requestData JSON

	b, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("read body", "err", err)
		render.Status(r, http.StatusNotFound)
		return
	}

	err = json.Unmarshal(b, &requestData)
	if err != nil {
		slog.Error("decode request body", "err", err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"status": "error", "message": "Error while decoding the data"})
		return
	}

	if requestData["type"] == nil || requestData["user_id"] == nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"status": "error", "message": "Not found type or user_id"})
		return
	}

	id := uuid.New().String()
	err = h.createOneEvent(id, requestData)
	if err != nil {
		slog.Error("create event", "err", err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"status": "error", "message": err})
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, JSON{"status": "ok", "uuid": id})
}

func (h Handler) OnCreateBatchEvents(w http.ResponseWriter, r *http.Request) {
	var requestData JSON

	b, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("read body", "err", err)
		render.Status(r, http.StatusNotFound)
		return
	}

	err = json.Unmarshal(b, &requestData)
	if err != nil {
		slog.Error("decode request body", "err", err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"status": "error", "message": "Error while decoding the data"})
		return
	}

	if requestData["type"] == nil || requestData["users"] == nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"status": "error", "message": "Not found type or users"})
		return
	}

	for _, user := range requestData["users"].([]interface{}) {
		id := uuid.New().String()
		requestData["user_id"] = user
		err = h.createOneEvent(id, requestData)
		if err != nil {
			slog.Error("create event", "err", err)
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, JSON{"status": "error", "message": err})
			return
		}
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, JSON{"status": "ok"})
}

func (h Handler) OnChangeBatchEvents(w http.ResponseWriter, r *http.Request) {
	var requestData JSON

	b, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("read body", "err", err)
		render.Status(r, http.StatusNotFound)
		return
	}

	err = json.Unmarshal(b, &requestData)
	if err != nil {
		slog.Error("decode request body", "err", err)
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"status": "error", "message": "Error while decoding the data"})
		return
	}

	if requestData["uuids"] == nil || requestData["status"] == nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"status": "error", "message": "Not found uuids or status"})
		return
	}

	uuids, ok := requestData["uuids"].([]interface{})
	if !ok {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"status": "error", "message": "Invalid uuids format"})
		return
	}

	status, ok := requestData["status"].(string)
	if !ok {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"status": "error", "message": "Invalid status format"})
		return
	}

	var count int64 = 0
	for _, id := range uuids {
		rec := event.Event{Status: status}
		cnt, err := h.StoreProvider.ChangeStatus(id.(string), rec)
		if err != nil {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, JSON{"status": "error", "message": err})
			return
		}
		count += cnt
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, JSON{"status": "ok", "count": count})
}

func (h Handler) OnGetEventsByUserId(w http.ResponseWriter, r *http.Request) {
	userId := chi.URLParam(r, "id")

	r.ParseForm()
	statuses := r.Form["status"]
	dateFrom := r.Form.Get("date_from")

	query := event.Query{Statuses: statuses, DateFrom: dateFrom}
	rows, err := h.StoreProvider.GetAllByUserId(userId, query)
	if err != nil {
		render.Status(r, http.StatusOK)
		render.JSON(w, r, JSON{"status": "ok", "data": []string{}})
		return
	}

	render.JSON(w, r, JSON{"status": "ok", "data": rows})
}

func (h Handler) OnChangeEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "uuid")

	var requestData JSON

	b, err := io.ReadAll(r.Body)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"status": "error", "message": "Error read the data"})
		return
	}

	err = json.Unmarshal(b, &requestData)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"status": "error", "message": "Error while decoding the data"})
		return
	}

	if requestData["status"] == nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"status": "error", "message": "Status is required"})
		return
	}

	var message string
	if requestData["message"] != nil {
		message = requestData["message"].(string)
	}

	rec := repository.Event{
		Status:  requestData["status"].(string),
		Message: message,
	}

	count, err := h.StoreProvider.ChangeStatus(id, rec)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"status": "error", "message": err})
	}

	if count == 0 {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, JSON{"status": "error", "message": "Not Found"})
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, JSON{"status": "ok", "uuid": id})
}

func (h Handler) OnSetSeen(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "uuid")

	count, err := h.StoreProvider.ChangeIsSeen(id)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, JSON{"status": "error", "message": err})
	}

	if count == 0 {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, JSON{"status": "error", "message": "Not Found"})
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, JSON{"status": "ok", "uuid": id})
}

func (h Handler) createOneEvent(id string, requestData JSON) error {
	userId, ok := requestData["user_id"].(string)
	if !ok {
		return errors.New("Invalid user_id format")
	}

	if requestData["caption"] == nil {
		requestData["caption"] = ""
	}
	if requestData["body"] == nil {
		requestData["body"] = ""
	}
	if requestData["status"] == nil {
		requestData["status"] = STATUS_NEW
	}

	rec := repository.Event{
		Uuid:    id,
		UserId:  userId,
		Status:  requestData["status"].(string),
		Type:    requestData["type"].(string),
		Caption: requestData["caption"].(string),
		Body:    requestData["body"].(string),
	}

	return h.StoreProvider.Create(rec)
}
