package db

import (
	"database/sql"
	"git.neds.sh/matty/entain/sports/proto/sports"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

func TestSportsRepo_List(t *testing.T) {
	// Open an in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	// Create a new EventsRepo using the test database
	eventsRepo := NewEventsRepo(db)

	// Initialize the test database with dummy data
	if err := initTestDB(db); err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	// Define test cases with different filters and expected outcomes
	testCases := []struct {
		name           string
		filter         *sports.ListEventsRequestFilter
		orderBy        []*sports.ListEventsRequestOrderBy
		expectedEvents []*sports.Event
	}{
		{
			name:   "NoFilter",
			filter: nil,
			expectedEvents: []*sports.Event{
				{
					Id:                  1,
					MeetingId:           5,
					Name:                "North Dakota foes",
					Visible:             false,
					Status:              "CLOSED",
					AdvertisedStartTime: timestamppb.New(time.Date(2022, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
				{
					Id:                  2,
					MeetingId:           1,
					Name:                "Connecticut griffins",
					Visible:             true,
					Status:              "OPEN",
					AdvertisedStartTime: timestamppb.New(time.Date(2023, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
				{
					Id:                  3,
					MeetingId:           8,
					Name:                "Rhode Island ghosts",
					Visible:             false,
					Status:              "OPEN",
					AdvertisedStartTime: timestamppb.New(time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
			},
		},
		{
			name: "FilterByMeetingIDs",
			filter: &sports.ListEventsRequestFilter{
				MeetingIds: []int64{5, 8},
			},
			expectedEvents: []*sports.Event{
				{
					Id:                  1,
					MeetingId:           5,
					Name:                "North Dakota foes",
					Visible:             false,
					Status:              "CLOSED",
					AdvertisedStartTime: timestamppb.New(time.Date(2022, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
				{
					Id:                  3,
					MeetingId:           8,
					Name:                "Rhode Island ghosts",
					Visible:             false,
					Status:              "OPEN",
					AdvertisedStartTime: timestamppb.New(time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
			},
		},
		{
			name: "FilterByMeetingIDsAndVisibility",
			filter: &sports.ListEventsRequestFilter{
				MeetingIds:       []int64{5, 1},
				VisibilityStatus: sports.VisibilityStatus_VISIBLE,
			},
			expectedEvents: []*sports.Event{
				{
					Id:                  2,
					MeetingId:           1,
					Name:                "Connecticut griffins",
					Visible:             true,
					Status:              "OPEN",
					AdvertisedStartTime: timestamppb.New(time.Date(2023, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
			},
		},

		{
			name: "FilterByVisibilityStatusVisible",
			filter: &sports.ListEventsRequestFilter{
				VisibilityStatus: sports.VisibilityStatus_VISIBLE,
			},
			expectedEvents: []*sports.Event{
				{
					Id:                  2,
					MeetingId:           1,
					Name:                "Connecticut griffins",
					Visible:             true,
					Status:              "OPEN",
					AdvertisedStartTime: timestamppb.New(time.Date(2023, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
			},
		},
		{
			name: "FilterByVisibilityStatusHidden",
			filter: &sports.ListEventsRequestFilter{
				VisibilityStatus: sports.VisibilityStatus_HIDDEN,
			},
			expectedEvents: []*sports.Event{
				{
					Id:                  1,
					MeetingId:           5,
					Name:                "North Dakota foes",
					Visible:             false,
					Status:              "CLOSED",
					AdvertisedStartTime: timestamppb.New(time.Date(2022, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
				{
					Id:                  3,
					MeetingId:           8,
					Name:                "Rhode Island ghosts",
					Visible:             false,
					Status:              "OPEN",
					AdvertisedStartTime: timestamppb.New(time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
			},
		},
		{
			name: "OrderByAdvertisedStartTimeDescending",
			orderBy: []*sports.ListEventsRequestOrderBy{
				{FieldName: "advertisedStartTime",
					Direction: sports.OrderByDirection_DESC,
				},
			},
			expectedEvents: []*sports.Event{
				{
					Id:                  3,
					MeetingId:           8,
					Name:                "Rhode Island ghosts",
					Visible:             false,
					Status:              "OPEN",
					AdvertisedStartTime: timestamppb.New(time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
				{
					Id:                  2,
					MeetingId:           1,
					Name:                "Connecticut griffins",
					Visible:             true,
					Status:              "OPEN",
					AdvertisedStartTime: timestamppb.New(time.Date(2023, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
				{
					Id:                  1,
					MeetingId:           5,
					Name:                "North Dakota foes",
					Visible:             false,
					Status:              "CLOSED",
					AdvertisedStartTime: timestamppb.New(time.Date(2022, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
			},
		},
	}

	// Run the test cases using table-driven testing
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the List method with the filter
			events, err := eventsRepo.List(tc.filter, tc.orderBy, getDateNow())
			if err != nil {
				t.Fatalf("failed to get events: %v", err)
			}

			// Compare the length of the returned events and the expected events
			if len(events) != len(tc.expectedEvents) {
				t.Fatalf("unexpected number of events: got %d, want %d", len(events), len(tc.expectedEvents))
			}

			// Compare each event returned with the expected events
			assert.Equal(t, tc.expectedEvents, events, "Unexpected events")
		})
	}
}

func TestSportsRepo_Get(t *testing.T) {
	// Open an in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	// Create a new EventsRepo using the test database
	eventsRepo := NewEventsRepo(db)

	// Initialize the test database with dummy data
	if err := initTestDB(db); err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	t.Run("GetById", func(t *testing.T) {
		// Call the List method with the filter
		event, err := eventsRepo.Get(2, getDateNow())
		if err != nil {
			t.Fatalf("failed to get events: %v", err)
		}

		// Compare if the right event was returned
		expectedRace := sports.Event{
			Id:                  2,
			MeetingId:           1,
			Name:                "Connecticut griffins",
			Visible:             true,
			Status:              "OPEN",
			AdvertisedStartTime: timestamppb.New(time.Date(2023, 7, 15, 12, 0, 0, 0, time.UTC)),
		}
		assert.Equal(t, event, &expectedRace)
	})

	t.Run("GetByIdNotFound", func(t *testing.T) {
		// Call the List method with the filter
		_, err := eventsRepo.Get(999, getDateNow())
		if err != sql.ErrNoRows {
			t.Fatalf("failed to get events: %v", err)
		}
	})
}

func initTestDB(db *sql.DB) error {
	statement, err := db.Prepare(`CREATE TABLE IF NOT EXISTS events (id INTEGER PRIMARY KEY, meeting_id INTEGER, name TEXT, visible INTEGER, advertised_start_time DATETIME)`)
	if err == nil {
		_, err = statement.Exec()
	}

	events := getAllTestData()

	for _, s := range events {
		statement, err = db.Prepare(`INSERT OR IGNORE INTO events(id, meeting_id, name, visible, advertised_start_time) VALUES (?,?,?,?,?)`)
		if err == nil {
			_, err = statement.Exec(
				s.Id,
				s.MeetingId,
				s.Name,
				s.Visible,
				s.AdvertisedStartTime.AsTime().Format(time.RFC3339),
			)
		}
	}

	return nil
}

func getAllTestData() []*sports.Event {
	return []*sports.Event{
		{
			Id:                  1,
			MeetingId:           5,
			Name:                "North Dakota foes",
			Visible:             false,
			AdvertisedStartTime: timestamppb.New(time.Date(2022, 7, 15, 12, 0, 0, 0, time.UTC)),
		},
		{
			Id:                  2,
			MeetingId:           1,
			Name:                "Connecticut griffins",
			Visible:             true,
			AdvertisedStartTime: timestamppb.New(time.Date(2023, 7, 15, 12, 0, 0, 0, time.UTC)),
		},
		{
			Id:                  3,
			MeetingId:           8,
			Name:                "Rhode Island ghosts",
			Visible:             false,
			AdvertisedStartTime: timestamppb.New(time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)),
		},
	}
}

func getDateNow() time.Time {
	return time.Date(2023, 7, 15, 12, 0, 0, 0, time.UTC)
}
