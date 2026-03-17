package web

import (
	"log/slog"
	"micro-rest-events/internal/repository"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const defaultPerPage = 20

type dashboardData struct {
	Events     []repository.Event
	UserID     string
	Status     string
	DateFrom   string
	Total      int
	New        int
	Seen       int
	Page       int
	PerPage    int
	TotalPages int
	PrevPage   int
	NextPage   int
}

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	data, err := s.loadDashboardData(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.tmpl.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		slog.Error("render dashboard", "err", err)
	}
}

func (s *Server) eventsTable(w http.ResponseWriter, r *http.Request) {
	data, err := s.loadDashboardData(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := s.tmpl.ExecuteTemplate(w, "events-table.html", data); err != nil {
		slog.Error("render events-table", "err", err)
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
		slog.Error("render events-table", "err", err)
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
		slog.Error("render events-table", "err", err)
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
		slog.Error("render events-table", "err", err)
	}
}

func (s *Server) loadDashboardData(r *http.Request) (dashboardData, error) {
	r.ParseForm()
	userID := r.FormValue("user_id")
	status := r.FormValue("status")
	dateFrom := r.FormValue("date_from")

	page := 1
	if p, err := strconv.Atoi(r.FormValue("page")); err == nil && p > 1 {
		page = p
	}

	var statuses []string
	if status != "" {
		statuses = []string{status}
	}

	filterQ := repository.Query{Statuses: statuses, DateFrom: dateFrom}
	pageQ := repository.Query{
		Statuses: statuses,
		DateFrom: dateFrom,
		Limit:    defaultPerPage,
		Offset:   (page - 1) * defaultPerPage,
	}

	var events []repository.Event
	var total int

	if userID != "" {
		total, _ = s.StoreProvider.CountByUserId(userID, filterQ)
		events, _ = s.StoreProvider.GetAllByUserId(userID, pageQ)
	} else {
		total, _ = s.StoreProvider.Count(filterQ)
		events, _ = s.StoreProvider.GetAll(pageQ)
	}

	if events == nil {
		events = []repository.Event{}
	}

	totalPages := (total + defaultPerPage - 1) / defaultPerPage
	prevPage := page - 1
	if prevPage < 1 {
		prevPage = 0
	}
	nextPage := page + 1
	if nextPage > totalPages {
		nextPage = 0
	}

	data := dashboardData{
		Events:     events,
		UserID:     userID,
		Status:     status,
		DateFrom:   dateFrom,
		Total:      total,
		Page:       page,
		PerPage:    defaultPerPage,
		TotalPages: totalPages,
		PrevPage:   prevPage,
		NextPage:   nextPage,
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
