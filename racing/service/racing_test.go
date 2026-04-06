package service

import (
	"context"
	"fmt"
	"testing"

	"git.neds.sh/matty/entain/racing/proto/racing"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// mockRacesRepo is a minimal stub that implements RacesRepo.
type mockRacesRepo struct {
	races []*racing.Race
}

func (m *mockRacesRepo) Init() error { return nil }

func (m *mockRacesRepo) List(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error) {
	return m.races, nil
}

func (m *mockRacesRepo) Get(id int64) (*racing.Race, error) {
	for _, r := range m.races {
		if r.Id == id {
			return r, nil
		}
	}
	return nil, fmt.Errorf("race with id %d not found", id)
}

func TestListRaces(t *testing.T) {
	repo := &mockRacesRepo{
		races: []*racing.Race{
			{Id: 1, Name: "Race A"},
			{Id: 2, Name: "Race B"},
		},
	}

	svc := NewRacingService(repo)
	resp, err := svc.ListRaces(context.Background(), &racing.ListRacesRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.Races, 2)
	assert.Equal(t, "Race A", resp.Races[0].Name)
}

func TestGetRace_Found(t *testing.T) {
	repo := &mockRacesRepo{
		races: []*racing.Race{
			{Id: 1, Name: "Race A"},
		},
	}

	svc := NewRacingService(repo)
	race, err := svc.GetRace(context.Background(), &racing.GetRaceRequest{Id: 1})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), race.Id)
}

func TestGetRace_NotFound_Returns404(t *testing.T) {
	repo := &mockRacesRepo{races: []*racing.Race{}}

	svc := NewRacingService(repo)
	race, err := svc.GetRace(context.Background(), &racing.GetRaceRequest{Id: 999})
	assert.Nil(t, race)

	// Verify the service wraps the error as gRPC NotFound.
	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}
