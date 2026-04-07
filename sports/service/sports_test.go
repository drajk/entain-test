package service

import (
	"context"
	"testing"

	"git.neds.sh/matty/entain/sports/proto/sports"
	"github.com/stretchr/testify/assert"
)

type mockEventsRepo struct {
	events []*sports.Event
}

func (m *mockEventsRepo) Init() error { return nil }

func (m *mockEventsRepo) List(filter *sports.ListEventsRequestFilter) ([]*sports.Event, error) {
	return m.events, nil
}

func TestListEvents(t *testing.T) {
	repo := &mockEventsRepo{
		events: []*sports.Event{
			{Id: 1, Name: "A vs B", EventType: "football"},
			{Id: 2, Name: "C vs D", EventType: "tennis"},
		},
	}

	svc := NewSportsService(repo)
	resp, err := svc.ListEvents(context.Background(), &sports.ListEventsRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.Events, 2)
	assert.Equal(t, "A vs B", resp.Events[0].Name)
}
