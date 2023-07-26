package db

import (
	"database/sql"
	"git.neds.sh/matty/entain/racing/proto/racing"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

func TestRacesRepo_List(t *testing.T) {
	// Open an in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory database: %v", err)
	}
	defer db.Close()

	// Create a new RacesRepo using the test database
	racesRepo := NewRacesRepo(db)

	// Initialize the test database with dummy data
	if err := initTestDB(db); err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	// Define test cases with different filters and expected outcomes
	testCases := []struct {
		name          string
		filter        *racing.ListRacesRequestFilter
		expectedRaces []*racing.Race
	}{
		{
			name:          "NoFilter",
			filter:        nil,
			expectedRaces: getAllTestData(),
		},
		{
			name: "FilterByMeetingIDs",
			filter: &racing.ListRacesRequestFilter{
				MeetingIds: []int64{5, 8},
			},
			expectedRaces: []*racing.Race{
				{
					Id:                  1,
					MeetingId:           5,
					Name:                "North Dakota foes",
					Number:              12,
					Visible:             false,
					AdvertisedStartTime: timestamppb.New(time.Date(2022, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
				{
					Id:                  3,
					MeetingId:           8,
					Name:                "Rhode Island ghosts",
					Number:              3,
					Visible:             false,
					AdvertisedStartTime: timestamppb.New(time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
			},
		},
		{
			name: "FilterByMeetingIDsAndVisibility",
			filter: &racing.ListRacesRequestFilter{
				MeetingIds:       []int64{5, 1},
				VisibilityStatus: racing.VisibilityStatus_VISIBLE,
			},
			expectedRaces: []*racing.Race{
				{
					Id:                  2,
					MeetingId:           1,
					Name:                "Connecticut griffins",
					Number:              12,
					Visible:             true,
					AdvertisedStartTime: timestamppb.New(time.Date(2023, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
			},
		},

		{
			name: "FilterByVisibilityStatusVisible",
			filter: &racing.ListRacesRequestFilter{
				VisibilityStatus: racing.VisibilityStatus_VISIBLE,
			},
			expectedRaces: []*racing.Race{
				{
					Id:                  2,
					MeetingId:           1,
					Name:                "Connecticut griffins",
					Number:              12,
					Visible:             true,
					AdvertisedStartTime: timestamppb.New(time.Date(2023, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
			},
		},
		{
			name: "FilterByVisibilityStatusHidden",
			filter: &racing.ListRacesRequestFilter{
				VisibilityStatus: racing.VisibilityStatus_HIDDEN,
			},
			expectedRaces: []*racing.Race{
				{
					Id:                  1,
					MeetingId:           5,
					Name:                "North Dakota foes",
					Number:              12,
					Visible:             false,
					AdvertisedStartTime: timestamppb.New(time.Date(2022, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
				{
					Id:                  3,
					MeetingId:           8,
					Name:                "Rhode Island ghosts",
					Number:              3,
					Visible:             false,
					AdvertisedStartTime: timestamppb.New(time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
			},
		},
	}

	// Run the test cases using table-driven testing
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the List method with the filter
			races, err := racesRepo.List(tc.filter)
			if err != nil {
				t.Fatalf("failed to get races: %v", err)
			}

			// Compare the length of the returned races and the expected races
			if len(races) != len(tc.expectedRaces) {
				t.Fatalf("unexpected number of races: got %d, want %d", len(races), len(tc.expectedRaces))
			}

			// Compare each race returned with the expected races
			assert.ElementsMatch(t, tc.expectedRaces, races, "Unexpected races")
		})
	}
}

func initTestDB(db *sql.DB) error {
	statement, err := db.Prepare(`CREATE TABLE IF NOT EXISTS races (id INTEGER PRIMARY KEY, meeting_id INTEGER, name TEXT, number INTEGER, visible INTEGER, advertised_start_time DATETIME)`)
	if err == nil {
		_, err = statement.Exec()
	}

	races := getAllTestData()

	for _, s := range races {
		statement, err = db.Prepare(`INSERT OR IGNORE INTO races(id, meeting_id, name, number, visible, advertised_start_time) VALUES (?,?,?,?,?,?)`)
		if err == nil {
			_, err = statement.Exec(
				s.Id,
				s.MeetingId,
				s.Name,
				s.Number,
				s.Visible,
				s.AdvertisedStartTime.AsTime().Format(time.RFC3339),
			)
		}
	}

	return nil
}

func getAllTestData() []*racing.Race {
	return []*racing.Race{
		{
			Id:                  1,
			MeetingId:           5,
			Name:                "North Dakota foes",
			Number:              12,
			Visible:             false,
			AdvertisedStartTime: timestamppb.New(time.Date(2022, 7, 15, 12, 0, 0, 0, time.UTC)),
		},
		{
			Id:                  2,
			MeetingId:           1,
			Name:                "Connecticut griffins",
			Number:              12,
			Visible:             true,
			AdvertisedStartTime: timestamppb.New(time.Date(2023, 7, 15, 12, 0, 0, 0, time.UTC)),
		},
		{
			Id:                  3,
			MeetingId:           8,
			Name:                "Rhode Island ghosts",
			Number:              3,
			Visible:             false,
			AdvertisedStartTime: timestamppb.New(time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)),
		},
	}
}
