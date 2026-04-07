package db

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"git.neds.sh/matty/entain/sports/proto/sports"
	"github.com/stretchr/testify/assert"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)
	_, err = db.Exec(`CREATE TABLE sport_events (id INTEGER PRIMARY KEY, name TEXT, event_type TEXT, visible INTEGER, advertised_start_time DATETIME)`)
	assert.NoError(t, err)
	return db
}

func insertEvent(t *testing.T, db *sql.DB, id int, name, eventType string, visible int, start time.Time) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO sport_events VALUES (?,?,?,?,?)`, id, name, eventType, visible, start.Format(time.RFC3339))
	assert.NoError(t, err)
}

func TestList_ReturnsAll(t *testing.T) {
	db := newTestDB(t)
	insertEvent(t, db, 1, "A vs B", "football", 1, time.Now().Add(1*time.Hour))
	insertEvent(t, db, 2, "C vs D", "tennis", 0, time.Now().Add(2*time.Hour))

	repo := &eventsRepo{db: db}
	events, err := repo.List(&sports.ListEventsRequestFilter{})
	assert.NoError(t, err)
	assert.Len(t, events, 2)
}

func TestList_VisibleOnly(t *testing.T) {
	db := newTestDB(t)
	insertEvent(t, db, 1, "A vs B", "football", 1, time.Now().Add(1*time.Hour))
	insertEvent(t, db, 2, "C vs D", "tennis", 0, time.Now().Add(2*time.Hour))

	repo := &eventsRepo{db: db}
	events, err := repo.List(&sports.ListEventsRequestFilter{VisibleOnly: true})
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, "A vs B", events[0].Name)
}
