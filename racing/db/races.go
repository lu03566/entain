package db

import (
	"database/sql"
	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"
	"strings"
	"sync"
	"time"
	"strconv"

	"git.neds.sh/matty/entain/racing/proto/racing"
)

// RacesRepo provides repository access to races.
type RacesRepo interface {
	// Init will initialise our races repository.
	Init() error

	// List will return a list of races.
	List(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error)

	// Get will allows us to fetch a single race by its ID.
	Get(filter *racing.GetRaceRequest) (*racing.Race, error)
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

	query, args = r.applyFilter(query, filter)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return r.scanRaces(rows)
}

// Get race by ID
func (r *racesRepo) Get(filter *racing.GetRaceRequest) (*racing.Race, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getRaceQueries()[racesList]

	query += " WHERE id=" + strconv.FormatInt(filter.Id, 10)

	rows, err := r.db.Query(query, args...)

	if err != nil {
		return nil, err
	}

	races, err := r.scanRaces(rows)

	// If no record found return nil (Server will crash if trying to return races[0] when there is no result)
	if len(races) == 0 {
		return nil, err
	}
	
	return races[0], err
}

func (r *racesRepo) applyFilter(query string, filter *racing.ListRacesRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}

	if len(filter.MeetingIds) > 0 {
		clauses = append(clauses, "meeting_id IN ("+strings.Repeat("?,", len(filter.MeetingIds)-1)+"?)")

		for _, meetingID := range filter.MeetingIds {
			args = append(args, meetingID)
		}
	}
	
	// Filter races that are visible only
	if filter.VisibleOnly {
		clauses = append(clauses, "visible=true")
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	// Order by advertised_start_time with option to specify to sort by ascending or descending order.
	query += " ORDER BY advertised_start_time" 
	if len(filter.OrderBy) > 0 && (filter.OrderBy == "ASC" || filter.OrderBy == "DESC")  {
		query += " " + filter.OrderBy;
	}

	return query, args
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

		// Check and compare `advertised_start_time`, if it's in the past status should reflect `CLOSED` ELSE `OPEN`
		if advertisedStart.Before(time.Now()) {
			race.Status = "CLOSED"
		} else {
			race.Status = "OPEN"
		}

		races = append(races, &race)
	}

	return races, nil
}
