package repository

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func newTestDB(t *testing.T) *StoreProvider {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uuid TEXT NOT NULL,
		user_id TEXT NULL,
		type TEXT NULL,
		status TEXT NULL,
		caption TEXT NULL,
		message TEXT NULL,
		is_seen BOOLEAN DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("create table: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return &StoreProvider{db: db, dbType: "sqlite"}
}

func seedEvent(t *testing.T, repo *StoreProvider, e Event) {
	t.Helper()
	if err := repo.Create(e); err != nil {
		t.Fatalf("seed event: %v", err)
	}
}

// --- Create ---

func TestCreate(t *testing.T) {
	repo := newTestDB(t)
	err := repo.Create(Event{
		Uuid:   "uuid-1",
		UserId: "user-1",
		Type:   "info",
		Status: "new",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreate_DuplicateUuid(t *testing.T) {
	repo := newTestDB(t)
	e := Event{Uuid: "uuid-dup", UserId: "user-1", Type: "info", Status: "new"}
	seedEvent(t, repo, e)
	// SQLite allows duplicate TEXT, but if we add UNIQUE it would fail.
	// Just verify two rows can be inserted with same uuid (current schema has no UNIQUE on uuid).
	if err := repo.Create(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- GetOne ---

func TestGetOne(t *testing.T) {
	repo := newTestDB(t)
	seedEvent(t, repo, Event{Uuid: "uuid-2", UserId: "user-2", Type: "alert", Status: "new", Caption: "cap", Body: "body text"})

	got, err := repo.GetOne("uuid-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Uuid != "uuid-2" {
		t.Errorf("got uuid %q, want uuid-2", got.Uuid)
	}
	if got.UserId != "user-2" {
		t.Errorf("got user_id %q, want user-2", got.UserId)
	}
	if got.Caption != "cap" {
		t.Errorf("got caption %q, want cap", got.Caption)
	}
}

func TestGetOne_NotFound(t *testing.T) {
	repo := newTestDB(t)
	_, err := repo.GetOne("non-existent")
	if err == nil {
		t.Error("expected error for missing uuid, got nil")
	}
}

// --- GetOneByUserId ---

func TestGetOneByUserId(t *testing.T) {
	repo := newTestDB(t)
	// GetOneByUserId returns first unseen event with status != 'new'
	seedEvent(t, repo, Event{Uuid: "uuid-3", UserId: "user-3", Type: "info", Status: "delivered"})

	got, err := repo.GetOneByUserId("user-3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Uuid != "uuid-3" {
		t.Errorf("got uuid %q, want uuid-3", got.Uuid)
	}
}

func TestGetOneByUserId_SkipsNewStatus(t *testing.T) {
	repo := newTestDB(t)
	seedEvent(t, repo, Event{Uuid: "uuid-new", UserId: "user-4", Type: "info", Status: "new"})

	_, err := repo.GetOneByUserId("user-4")
	if err == nil {
		t.Error("expected Row Not Found for 'new' status event, got nil")
	}
}

func TestGetOneByUserId_NotFound(t *testing.T) {
	repo := newTestDB(t)
	_, err := repo.GetOneByUserId("no-such-user")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// --- GetAll ---

func TestGetAll_Empty(t *testing.T) {
	repo := newTestDB(t)
	events, err := repo.GetAll(Query{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

func TestGetAll_ReturnsAll(t *testing.T) {
	repo := newTestDB(t)
	seedEvent(t, repo, Event{Uuid: "a1", UserId: "u1", Type: "t", Status: "new"})
	seedEvent(t, repo, Event{Uuid: "a2", UserId: "u2", Type: "t", Status: "done"})

	events, err := repo.GetAll(Query{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

func TestGetAll_FilterByStatus(t *testing.T) {
	repo := newTestDB(t)
	seedEvent(t, repo, Event{Uuid: "b1", UserId: "u1", Type: "t", Status: "new"})
	seedEvent(t, repo, Event{Uuid: "b2", UserId: "u1", Type: "t", Status: "done"})
	seedEvent(t, repo, Event{Uuid: "b3", UserId: "u1", Type: "t", Status: "new"})

	events, err := repo.GetAll(Query{Statuses: []string{"new"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 'new' events, got %d", len(events))
	}
	for _, e := range events {
		if e.Status != "new" {
			t.Errorf("unexpected status %q", e.Status)
		}
	}
}

func TestGetAll_FilterByMultipleStatuses(t *testing.T) {
	repo := newTestDB(t)
	seedEvent(t, repo, Event{Uuid: "c1", UserId: "u1", Type: "t", Status: "new"})
	seedEvent(t, repo, Event{Uuid: "c2", UserId: "u1", Type: "t", Status: "done"})
	seedEvent(t, repo, Event{Uuid: "c3", UserId: "u1", Type: "t", Status: "failed"})

	events, err := repo.GetAll(Query{Statuses: []string{"new", "done"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events, got %d", len(events))
	}
}

// --- GetAllByUserId ---

func TestGetAllByUserId(t *testing.T) {
	repo := newTestDB(t)
	seedEvent(t, repo, Event{Uuid: "d1", UserId: "user-A", Type: "t", Status: "new"})
	seedEvent(t, repo, Event{Uuid: "d2", UserId: "user-A", Type: "t", Status: "done"})
	seedEvent(t, repo, Event{Uuid: "d3", UserId: "user-B", Type: "t", Status: "new"})

	events, err := repo.GetAllByUserId("user-A", Query{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("expected 2 events for user-A, got %d", len(events))
	}
	for _, e := range events {
		if e.UserId != "user-A" {
			t.Errorf("got event for wrong user %q", e.UserId)
		}
	}
}

func TestGetAllByUserId_FilterByStatus(t *testing.T) {
	repo := newTestDB(t)
	seedEvent(t, repo, Event{Uuid: "e1", UserId: "user-C", Type: "t", Status: "new"})
	seedEvent(t, repo, Event{Uuid: "e2", UserId: "user-C", Type: "t", Status: "done"})

	events, err := repo.GetAllByUserId("user-C", Query{Statuses: []string{"done"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
	if events[0].Uuid != "e2" {
		t.Errorf("got uuid %q, want e2", events[0].Uuid)
	}
}

func TestGetAllByUserId_Empty(t *testing.T) {
	repo := newTestDB(t)
	events, err := repo.GetAllByUserId("unknown-user", Query{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}

// --- ChangeStatus ---

func TestChangeStatus(t *testing.T) {
	repo := newTestDB(t)
	seedEvent(t, repo, Event{Uuid: "f1", UserId: "u1", Type: "t", Status: "new"})

	count, err := repo.ChangeStatus("f1", Event{Status: "done"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row affected, got %d", count)
	}

	got, _ := repo.GetOne("f1")
	if got.Status != "done" {
		t.Errorf("expected status 'done', got %q", got.Status)
	}
}

func TestChangeStatus_WithMessage(t *testing.T) {
	repo := newTestDB(t)
	seedEvent(t, repo, Event{Uuid: "f2", UserId: "u1", Type: "t", Status: "new"})

	count, err := repo.ChangeStatus("f2", Event{Status: "failed", Message: "timeout error"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row affected, got %d", count)
	}

	got, _ := repo.GetOne("f2")
	if got.Status != "failed" {
		t.Errorf("expected status 'failed', got %q", got.Status)
	}
	if got.Message != "timeout error" {
		t.Errorf("expected message 'timeout error', got %q", got.Message)
	}
}

func TestChangeStatus_NotFound(t *testing.T) {
	repo := newTestDB(t)
	count, err := repo.ChangeStatus("non-existent", Event{Status: "done"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 rows affected, got %d", count)
	}
}

// --- ChangeIsSeen ---

func TestChangeIsSeen(t *testing.T) {
	repo := newTestDB(t)
	seedEvent(t, repo, Event{Uuid: "g1", UserId: "u1", Type: "t", Status: "new"})

	count, err := repo.ChangeIsSeen("g1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row affected, got %d", count)
	}

	got, _ := repo.GetOne("g1")
	if !got.IsSeen {
		t.Error("expected is_seen = true after ChangeIsSeen")
	}
	if got.Status != "seen" {
		t.Errorf("expected status = 'seen' after ChangeIsSeen, got %q", got.Status)
	}
}

func TestChangeIsSeen_NotFound(t *testing.T) {
	repo := newTestDB(t)
	count, err := repo.ChangeIsSeen("non-existent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 rows affected, got %d", count)
	}
}
