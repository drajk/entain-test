package db

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"git.neds.sh/matty/entain/racing/proto/racing"
	"github.com/stretchr/testify/assert"
)

func TestApplyFilter_VisibleFlag(t *testing.T) {
	repo := &racesRepo{}
	baseQuery := "SELECT id FROM races"
	defaultOrder := " ORDER BY advertised_start_time ASC"

	tests := []struct {
		name            string
		filter          *racing.ListRacesRequestFilter
		expectedQuery   string
		expectedNumArgs int
	}{
		{
			name:            "nil filter defaults to order by start time",
			filter:          nil,
			expectedQuery:   baseQuery + defaultOrder,
			expectedNumArgs: 0,
		},
		{
			name:            "empty filter defaults to order by start time",
			filter:          &racing.ListRacesRequestFilter{},
			expectedQuery:   baseQuery + defaultOrder,
			expectedNumArgs: 0,
		},
		{
			name:            "visible_only adds visible clause",
			filter:          &racing.ListRacesRequestFilter{VisibleOnly: true},
			expectedQuery:   baseQuery + " WHERE visible = 1" + defaultOrder,
			expectedNumArgs: 0,
		},
		{
			name:            "visible_only false returns all",
			filter:          &racing.ListRacesRequestFilter{VisibleOnly: false},
			expectedQuery:   baseQuery + defaultOrder,
			expectedNumArgs: 0,
		},
		{
			name:            "meeting_ids filter",
			filter:          &racing.ListRacesRequestFilter{MeetingIds: []int64{1, 2}},
			expectedQuery:   baseQuery + " WHERE meeting_id IN (?,?)" + defaultOrder,
			expectedNumArgs: 2,
		},
		{
			name:            "both filters combined",
			filter:          &racing.ListRacesRequestFilter{MeetingIds: []int64{1}, VisibleOnly: true},
			expectedQuery:   baseQuery + " WHERE meeting_id IN (?) AND visible = 1" + defaultOrder,
			expectedNumArgs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuery, gotArgs, err := repo.applyFilter(baseQuery, tt.filter)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedQuery, gotQuery)
			assert.Len(t, gotArgs, tt.expectedNumArgs)
		})
	}
}

func TestApplyFilter_Ordering(t *testing.T) {
	repo := &racesRepo{}
	baseQuery := "SELECT id FROM races"

	tests := []struct {
		name          string
		filter        *racing.ListRacesRequestFilter
		expectedQuery string
		expectErr     bool
	}{
		{
			name:          "default sort by start time asc",
			filter:        &racing.ListRacesRequestFilter{},
			expectedQuery: baseQuery + " ORDER BY advertised_start_time ASC",
		},
		{
			name:          "custom sort field",
			filter:        &racing.ListRacesRequestFilter{SortBy: "name"},
			expectedQuery: baseQuery + " ORDER BY name ASC",
		},
		{
			name:          "custom sort field with desc",
			filter:        &racing.ListRacesRequestFilter{SortBy: "number", SortDirection: racing.SortDirection_DESC},
			expectedQuery: baseQuery + " ORDER BY number DESC",
		},
		{
			name:          "desc on default field",
			filter:        &racing.ListRacesRequestFilter{SortDirection: racing.SortDirection_DESC},
			expectedQuery: baseQuery + " ORDER BY advertised_start_time DESC",
		},
		{
			name:      "invalid sort field rejected",
			filter:    &racing.ListRacesRequestFilter{SortBy: "DROP TABLE races"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, _, err := repo.applyFilter(baseQuery, tt.filter)
			if tt.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedQuery, q)
		})
	}
}

// newTestDB spins up an in-memory SQLite with the races schema.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)
	_, err = db.Exec(`CREATE TABLE races (id INTEGER PRIMARY KEY, meeting_id INT, name TEXT, number INT, visible INT, advertised_start_time DATETIME)`)
	assert.NoError(t, err)
	return db
}

func insertRace(t *testing.T, db *sql.DB, id int, name string, startTime time.Time) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO races VALUES (?,1,?,1,1,?)`, id, name, startTime.Format(time.RFC3339))
	assert.NoError(t, err)
}

func TestList_DerivedStatus(t *testing.T) {
	db := newTestDB(t)
	insertRace(t, db, 1, "Past Race", time.Now().Add(-1*time.Hour))
	insertRace(t, db, 2, "Future Race", time.Now().Add(1*time.Hour))

	repo := &racesRepo{db: db}
	races, err := repo.List(&racing.ListRacesRequestFilter{})
	assert.NoError(t, err)
	assert.Len(t, races, 2)

	// Default order is advertised_start_time ASC, so past comes first.
	assert.Equal(t, racing.RaceStatus_CLOSED, races[0].Status)
	assert.Equal(t, racing.RaceStatus_OPEN, races[1].Status)
}

func TestGet_Found(t *testing.T) {
	db := newTestDB(t)
	insertRace(t, db, 1, "Test Race", time.Now().Add(1*time.Hour))

	repo := &racesRepo{db: db}
	race, err := repo.Get(1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), race.Id)
	assert.Equal(t, "Test Race", race.Name)
}

func TestGet_NotFound(t *testing.T) {
	db := newTestDB(t)

	repo := &racesRepo{db: db}
	race, err := repo.Get(999)
	assert.Nil(t, race)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
