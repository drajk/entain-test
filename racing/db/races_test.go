package db

import (
	"testing"

	"git.neds.sh/matty/entain/racing/proto/racing"
	"github.com/stretchr/testify/assert"
)

func TestApplyFilter(t *testing.T) {
	repo := &racesRepo{}
	baseQuery := "SELECT id FROM races"

	tests := []struct {
		name            string
		filter          *racing.ListRacesRequestFilter
		expectedQuery   string
		expectedNumArgs int
	}{
		{
			name:            "nil filter returns base query",
			filter:          nil,
			expectedQuery:   baseQuery,
			expectedNumArgs: 0,
		},
		{
			name:            "empty filter returns base query",
			filter:          &racing.ListRacesRequestFilter{},
			expectedQuery:   baseQuery,
			expectedNumArgs: 0,
		},
		{
			name:            "visible_only adds visible clause",
			filter:          &racing.ListRacesRequestFilter{VisibleOnly: true},
			expectedQuery:   baseQuery + " WHERE visible = 1",
			expectedNumArgs: 0,
		},
		{
			name:            "visible_only false returns base query",
			filter:          &racing.ListRacesRequestFilter{VisibleOnly: false},
			expectedQuery:   baseQuery,
			expectedNumArgs: 0,
		},
		{
			name:            "meeting_ids filter",
			filter:          &racing.ListRacesRequestFilter{MeetingIds: []int64{1, 2}},
			expectedQuery:   baseQuery + " WHERE meeting_id IN (?,?)",
			expectedNumArgs: 2,
		},
		{
			name:            "both filters combined",
			filter:          &racing.ListRacesRequestFilter{MeetingIds: []int64{1}, VisibleOnly: true},
			expectedQuery:   baseQuery + " WHERE meeting_id IN (?) AND visible = 1",
			expectedNumArgs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQuery, gotArgs := repo.applyFilter(baseQuery, tt.filter)
			assert.Equal(t, tt.expectedQuery, gotQuery)
			assert.Len(t, gotArgs, tt.expectedNumArgs)
		})
	}
}
