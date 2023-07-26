package db

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"

	"git.neds.sh/matty/entain/racing/proto/racing"
)

// RacesRepo provides repository access to races.
type RacesRepo interface {
	// Init will initialise our races repository.
	Init() error

	// List will return a list of races.
	List(filter *racing.ListRacesRequestFilter, orderBy []*racing.ListRacesRequestOrderBy, currentDate time.Time) ([]*racing.Race, error)
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

func (r *racesRepo) List(filter *racing.ListRacesRequestFilter, orderBy []*racing.ListRacesRequestOrderBy, currentDate time.Time) ([]*racing.Race, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getRaceQueries()[racesList]

	query, args = r.applyFilter(query, filter)

	query = r.applyOrderBy(query, orderBy)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return r.scanRaces(rows, currentDate)
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

	switch filter.VisibilityStatus {
	case racing.VisibilityStatus_VISIBLE:
		clauses = append(clauses, "visible = 1")
	case racing.VisibilityStatus_HIDDEN:
		clauses = append(clauses, "visible = 0")
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	return query, args
}

func (r *racesRepo) applyOrderBy(query string, orderBy []*racing.ListRacesRequestOrderBy) string {
	var (
		clauses []string
	)

	if orderBy == nil {
		return query
	}

	for _, orderByClause := range orderBy {
		if strings.ToLower(orderByClause.FieldName) == "advertisedstarttime" {
			if orderByClause.Direction == racing.OrderByDirection_DESC {
				clauses = append(clauses, "advertised_start_time desc")
			} else {
				clauses = append(clauses, "advertised_start_time")
			}
		}
	}

	if len(clauses) != 0 {
		query += " ORDER BY " + strings.Join(clauses, ",")
	}

	return query
}

func (r *racesRepo) scanRaces(rows *sql.Rows, currentDate time.Time) ([]*racing.Race, error) {
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
		if advertisedStart.Before(currentDate) {
			race.Status = "CLOSED"
		} else {
			race.Status = "OPEN"
		}
		races = append(races, &race)
	}

	return races, nil
}
