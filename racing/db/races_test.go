package db

import (
	"testing"

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
