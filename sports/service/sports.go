package service

import (
	"git.neds.sh/matty/entain/sports/db"
	"git.neds.sh/matty/entain/sports/proto/sports"
	"golang.org/x/net/context"
)

type Sports interface {
	// ListEvents will return a collection of events.
	ListEvents(ctx context.Context, in *sports.ListEventsRequest) (*sports.ListEventsResponse, error)
}

// sportService implements the Sports interface.
type sportService struct {
	eventsRepo db.EventsRepo
}

// NewSportsService instantiates and returns a new sportService.
func NewSportsService(eventsRepo db.EventsRepo) Sports {
	return &sportService{eventsRepo}
}

func (s *sportService) ListEvents(ctx context.Context, in *sports.ListEventsRequest) (*sports.ListEventsResponse, error) {
	events, err := s.eventsRepo.List(in.Filter)
	if err != nil {
		return nil, err
	}

	return &sports.ListEventsResponse{Events: events}, nil
}