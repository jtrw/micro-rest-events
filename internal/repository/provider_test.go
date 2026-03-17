package repository

import (
	"os"
	"path/filepath"
	"testing"
)

// --- dbType detection (via NewStoreProvider) ---

func TestNewStoreProvider_SQLiteFilePrefix(t *testing.T) {
	p, err := NewStoreProvider("file:///tmp/test_provider.db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sp := p.(*StoreProvider)
	if sp.dbType != "sqlite" {
		t.Errorf("expected dbType sqlite, got %q", sp.dbType)
	}
}

func TestNewStoreProvider_SQLiteDotSqlite(t *testing.T) {
	f := filepath.Join(t.TempDir(), "test.sqlite")
	p, err := NewStoreProvider(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sp := p.(*StoreProvider)
	if sp.dbType != "sqlite" {
		t.Errorf("expected dbType sqlite, got %q", sp.dbType)
	}
}

func TestNewStoreProvider_SQLiteDotDb(t *testing.T) {
	f := filepath.Join(t.TempDir(), "test.db")
	p, err := NewStoreProvider(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sp := p.(*StoreProvider)
	if sp.dbType != "sqlite" {
		t.Errorf("expected dbType sqlite, got %q", sp.dbType)
	}
}

func TestNewStoreProvider_UnsupportedConnString(t *testing.T) {
	_, err := NewStoreProvider("redis://localhost:6379")
	if err == nil {
		t.Fatal("expected error for unsupported connection string, got nil")
	}
}

func TestNewStoreProvider_EmptyConnString(t *testing.T) {
	_, err := NewStoreProvider("")
	if err == nil {
		t.Fatal("expected error for empty connection string, got nil")
	}
}

func TestNewStoreProvider_PostgresPrefix(t *testing.T) {
	// postgres:// is detected correctly; sql.Open fails to connect but the
	// dbType detection itself is correct — error comes from CREATE TABLE, not dbType.
	_, err := NewStoreProvider("postgres://user:pass@localhost:5432/db?sslmode=disable")
	// We expect an error (no real postgres running), but NOT "unsupported database type"
	if err == nil {
		t.Skip("unexpected success — postgres may be running locally")
	}
	if err.Error() == "can't determine database type: unsupported database type in connection string" {
		t.Errorf("postgres:// should be detected as postgres, not unsupported: %v", err)
	}
}

func TestNewStoreProvider_MySQLPrefix(t *testing.T) {
	_, err := NewStoreProvider("user:pass@tcp(localhost:3306)/db")
	if err == nil {
		t.Skip("unexpected success — mysql may be running locally")
	}
	if err.Error() == "can't determine database type: unsupported database type in connection string" {
		t.Errorf("@tcp( should be detected as mysql, not unsupported: %v", err)
	}
}

// --- table creation ---

func TestNewStoreProvider_CreatesEventsTable(t *testing.T) {
	f := filepath.Join(t.TempDir(), "schema.db")
	p, err := NewStoreProvider(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sp := p.(*StoreProvider)

	// Verify the table exists by querying sqlite_master
	var name string
	row := sp.db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name='events'`)
	if err := row.Scan(&name); err != nil {
		t.Fatalf("events table not found: %v", err)
	}
	if name != "events" {
		t.Errorf("expected table name 'events', got %q", name)
	}
}

func TestNewStoreProvider_IdempotentTableCreation(t *testing.T) {
	f := filepath.Join(t.TempDir(), "idempotent.db")

	// First call creates the table
	_, err := NewStoreProvider(f)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}

	// Second call with the same file must not fail (CREATE TABLE IF NOT EXISTS)
	_, err = NewStoreProvider(f)
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
}

// --- implements interface ---

func TestNewStoreProvider_ImplementsInterface(t *testing.T) {
	f := filepath.Join(t.TempDir(), "iface.db")
	p, err := NewStoreProvider(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// compile-time check is enough, but also assert at runtime
	var _ StoreProviderInterface = p
}

// --- file actually created on disk ---

func TestNewStoreProvider_FileCreatedOnDisk(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "disk.db")

	if _, err := os.Stat(f); !os.IsNotExist(err) {
		t.Fatal("file should not exist before NewStoreProvider")
	}

	_, err := NewStoreProvider(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(f); os.IsNotExist(err) {
		t.Error("expected db file to be created on disk")
	}
}
