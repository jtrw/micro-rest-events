package web

import (
	"log"
	"micro-rest-events/internal/repository"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type dashboardData struct {
	Events   []repository.Event
	UserID   string
	Status   string
	DateFrom string
	Total    int
	New      int
	Seen     int
}

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	data, err := s.loadDashboardData(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.tmpl.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		log.Printf("[ERROR] render dashboard: %v", err)
	}
}

func (s *Server) eventsTable(w http.ResponseWriter, r *http.Request) {
	data, err := s.loadDashboardData(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.tmpl.ExecuteTemplate(w, "events-table.html", data); err != nil {
		log.Printf("[ERROR] render events-table: %v", err)
	}
}

func (s *Server) createEvent(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	userID := r.FormValue("user_id")
	eventType := r.FormValue("type")
	caption := r.FormValue("caption")
	body := r.FormValue("body")

	if userID == "" || eventType == "" {
		http.Error(w, "user_id and type are required", http.StatusBadRequest)
		return
	}

	rec := repository.Event{
		Uuid:    uuid.New().String(),
		UserId:  userID,
		Type:    eventType,
		Status:  STATUS_NEW,
		Caption: caption,
		Body:    body,
	}

	if err := s.StoreProvider.Create(rec); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := s.loadDashboardData(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.tmpl.ExecuteTemplate(w, "events-table.html", data); err != nil {
		log.Printf("[ERROR] render events-table: %v", err)
	}
}

func (s *Server) changeStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "uuid")

	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	status := r.FormValue("status")
	if status == "" {
		http.Error(w, "status is required", http.StatusBadRequest)
		return
	}

	rec := repository.Event{Status: status}
	if _, err := s.StoreProvider.ChangeStatus(id, rec); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := s.loadDashboardData(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.tmpl.ExecuteTemplate(w, "events-table.html", data); err != nil {
		log.Printf("[ERROR] render events-table: %v", err)
	}
}

func (s *Server) markSeen(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "uuid")

	if _, err := s.StoreProvider.ChangeIsSeen(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := s.loadDashboardData(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.tmpl.ExecuteTemplate(w, "events-table.html", data); err != nil {
		log.Printf("[ERROR] render events-table: %v", err)
	}
}

func (s *Server) loadDashboardData(r *http.Request) (dashboardData, error) {
	r.ParseForm()
	userID := r.FormValue("user_id")
	status := r.FormValue("status")
	dateFrom := r.FormValue("date_from")

	var statuses []string
	if status != "" {
		statuses = []string{status}
	}

	q := repository.Query{Statuses: statuses, DateFrom: dateFrom}

	var events []repository.Event
	var err error

	if userID != "" {
		events, err = s.StoreProvider.GetAllByUserId(userID, q)
	} else {
		events, err = s.StoreProvider.GetAll(q)
	}

	if err != nil {
		events = []repository.Event{}
	}

	data := dashboardData{
		Events:   events,
		UserID:   userID,
		Status:   status,
		DateFrom: dateFrom,
		Total:    len(events),
	}

	for _, e := range events {
		if e.Status == "new" {
			data.New++
		}
		if e.IsSeen {
			data.Seen++
		}
	}

	return data, nil
}
