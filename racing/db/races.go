package db

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"

	"git.neds.sh/matty/entain/racing/proto/racing"
)

// sortable columns — anything not in here gets rejected
var allowedSortFields = map[string]bool{
	"id":                    true,
	"meeting_id":            true,
	"name":                  true,
	"number":                true,
	"advertised_start_time": true,
}

const (
	defaultSortField     = "advertised_start_time"
	defaultSortDirection = "ASC"
)

// RacesRepo provides repository access to races.
type RacesRepo interface {
	// Init will initialise our races repository.
	Init() error

	// List will return a list of races.
	List(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error)

	// Get returns a single race by ID.
	Get(id int64) (*racing.Race, error)
}

type racesRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewRacesRepo creates a new races repository.
func NewRacesRepo(db *sql.DB) RacesRepo {
	return &racesRepo{db: db}
}

// Init prepares the race repository dummy data.
func (r *racesRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy races.
		err = r.seed()
	})

	return err
}

func (r *racesRepo) List(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getRaceQueries()[racesList]

	query, args, err = r.applyFilter(query, filter)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return r.scanRaces(rows)
}

func (r *racesRepo) Get(id int64) (*racing.Race, error) {
	rows, err := r.db.Query(getRaceQueries()[racesGet], id)
	if err != nil {
		return nil, err
	}

	races, err := r.scanRaces(rows)
	if err != nil {
		return nil, err
	}

	if len(races) == 0 {
		return nil, fmt.Errorf("race with id %d not found", id)
	}

	return races[0], nil
}

func (r *racesRepo) applyFilter(query string, filter *racing.ListRacesRequestFilter) (string, []interface{}, error) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query + " ORDER BY " + defaultSortField + " " + defaultSortDirection, args, nil
	}

	if len(filter.MeetingIds) > 0 {
		clauses = append(clauses, "meeting_id IN ("+strings.Repeat("?,", len(filter.MeetingIds)-1)+"?)")

		for _, meetingID := range filter.MeetingIds {
			args = append(args, meetingID)
		}
	}

	if filter.VisibleOnly {
		clauses = append(clauses, "visible = 1")
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	query, err := applyOrdering(query, filter)
	if err != nil {
		return query, args, err
	}

	return query, args, nil
}

// applyOrdering appends ORDER BY to the query. Defaults to advertised_start_time ASC.
// Validation lives here for now, but in a larger system this would sit in the service layer.
func applyOrdering(query string, filter *racing.ListRacesRequestFilter) (string, error) {
	field := defaultSortField
	direction := defaultSortDirection

	if filter.SortBy != "" {
		if !allowedSortFields[filter.SortBy] {
			allowed := make([]string, 0, len(allowedSortFields))
			for k := range allowedSortFields {
				allowed = append(allowed, k)
			}
			return "", fmt.Errorf("invalid sort field: %s, allowed: %v", filter.SortBy, allowed)
		}
		field = filter.SortBy
	}

	if filter.SortDirection == racing.SortDirection_DESC {
		direction = "DESC"
	}

	return query + " ORDER BY " + field + " " + direction, nil
}

func (m *racesRepo) scanRaces(
	rows *sql.Rows,
) ([]*racing.Race, error) {
	var races []*racing.Race

	for rows.Next() {
		var race racing.Race
		var advertisedStart time.Time

		if err := rows.Scan(&race.Id, &race.MeetingId, &race.Name, &race.Number, &race.Visible, &advertisedStart); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		ts, err := ptypes.TimestampProto(advertisedStart)
		if err != nil {
			return nil, err
		}

		race.AdvertisedStartTime = ts

		// Derive status at read time, simplest and always accurate.
		// Could also live in the service layer or be stored in DB, this was the simplest option for this example.
		if advertisedStart.Before(time.Now()) {
			race.Status = racing.RaceStatus_CLOSED
		}

		races = append(races, &race)
	}

	return races, nil
}
