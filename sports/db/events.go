package db

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"

	"git.neds.sh/matty/entain/sports/proto/sports"
)

// EventsRepo provides repository access to events.
type EventsRepo interface {
	// Init will initialise our events repository.
	Init() error

	// List will return a list of events.
	List(filter *sports.ListEventsRequestFilter, orderBy []*sports.ListEventsRequestOrderBy, currentDate time.Time) ([]*sports.Event, error)

	// Get will return a single event. It will return an error if no event is found
	Get(id int64, currentDate time.Time) (*sports.Event, error)
}

type eventsRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewEventsRepo creates a new sports repository.
func NewEventsRepo(db *sql.DB) EventsRepo {
	return &eventsRepo{db: db}
}

// Init prepares the event repository dummy data.
func (r *eventsRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy events.
		err = r.seed()
	})

	return err
}

// List Returns a list of events
func (r *eventsRepo) List(filter *sports.ListEventsRequestFilter, orderBy []*sports.ListEventsRequestOrderBy, currentDate time.Time) ([]*sports.Event, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getEventsQueries()[eventsList]

	query, args = r.applyFilter(query, filter)

	query = r.applyOrderBy(query, orderBy)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return r.scanEvents(rows, currentDate)
}

func (r *eventsRepo) applyFilter(query string, filter *sports.ListEventsRequestFilter) (string, []interface{}) {
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
	case sports.VisibilityStatus_VISIBLE:
		clauses = append(clauses, "visible = 1")
	case sports.VisibilityStatus_HIDDEN:
		clauses = append(clauses, "visible = 0")
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	return query, args
}

func (r *eventsRepo) applyOrderBy(query string, orderBy []*sports.ListEventsRequestOrderBy) string {
	var (
		clauses []string
	)

	if orderBy == nil {
		return query
	}

	for _, orderByClause := range orderBy {
		if strings.ToLower(orderByClause.FieldName) == "advertisedstarttime" {
			if orderByClause.Direction == sports.OrderByDirection_DESC {
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

// Get Return a single event by id
func (r *eventsRepo) Get(id int64, currentDate time.Time) (*sports.Event, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getEventsQueries()[eventsList]
	query += " WHERE Id = ?"
	args = append(args, id)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	events, err := r.scanEvents(rows, currentDate)

	if len(events) == 1 {
		return events[0], err
	} else {
		// in case a event is not found return an error for no rows
		return nil, sql.ErrNoRows
	}
}

func (r *eventsRepo) scanEvents(rows *sql.Rows, currentDate time.Time) ([]*sports.Event, error) {
	var events []*sports.Event

	for rows.Next() {
		var event sports.Event
		var advertisedStart time.Time

		if err := rows.Scan(&event.Id, &event.MeetingId, &event.Name, &event.Visible, &advertisedStart); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		ts, err := ptypes.TimestampProto(advertisedStart)
		if err != nil {
			return nil, err
		}

		event.AdvertisedStartTime = ts
		if advertisedStart.Before(currentDate) {
			event.Status = "CLOSED"
		} else {
			event.Status = "OPEN"
		}
		events = append(events, &event)
	}

	return events, nil
}
