package service

import (
	"context"
	"git.neds.sh/matty/entain/racing/proto/racing"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

// MockRacesRepo is a mock implementation of the db.RacesRepo interface.
type MockRacesRepo struct{}

func (m *MockRacesRepo) Init() error {
	return nil
}

func (m *MockRacesRepo) List(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error) {
	// Mock the behavior here and return a predefined response.
	// For simplicity, we'll return a predefined list of races.
	races := []*racing.Race{
		{Id: 1,
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

	var filteredRaces []*racing.Race
	for _, race := range races {
		if filter.GetVisibilityStatus() == racing.VisibilityStatus_VISIBLE && !race.GetVisible() {
			// Skip races that are not visible
			continue
		}

		if filter.GetVisibilityStatus() == racing.VisibilityStatus_HIDDEN && race.GetVisible() {
			// Skip races that are visible
			continue
		}

		if len(filter.MeetingIds) > 0 {
			// Skip races that don't match meeting ids
			var result bool = false
			for _, x := range filter.MeetingIds {
				if x == race.MeetingId {
					result = true
					break
				}
			}
			if !result {
				continue
			}
		}

		// Add the race to the filtered list if it passes the filter criteria
		filteredRaces = append(filteredRaces, race)
	}

	return filteredRaces, nil
}

func TestRacingService_ListRaces(t *testing.T) {
	// Define test cases with different inputs and expected outputs
	testCases := []struct {
		name          string
		filter        *racing.ListRacesRequestFilter
		expectedRaces []*racing.Race
		expectedErr   bool
	}{
		{
			name:   "NoFilter",
			filter: &racing.ListRacesRequestFilter{
				// empty filter
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
			},
			expectedErr: false,
		},
		{
			name: "FilterByMeetingIDs",
			filter: &racing.ListRacesRequestFilter{
				MeetingIds: []int64{1},
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
			expectedErr: false,
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
			expectedErr: false,
		},
		{
			name: "FilterByVisibilityStatusHidden",
			filter: &racing.ListRacesRequestFilter{
				VisibilityStatus: racing.VisibilityStatus_HIDDEN,
			},
			expectedRaces: []*racing.Race{
				{Id: 1,
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
			expectedErr: false,
		},
	}

	// Create a mock RacesRepo and pass it to the racingService
	racesRepo := &MockRacesRepo{}
	racingSvc := NewRacingService(racesRepo)

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare the request
			request := &racing.ListRacesRequest{
				Filter: tc.filter,
			}

			// Call the method being tested
			response, err := racingSvc.ListRaces(context.Background(), request)

			// Check for errors
			if (err != nil) != tc.expectedErr {
				t.Fatalf("unexpected error: %v", err)
			}

			// Compare the response races to the expected races
			if len(response.Races) != len(tc.expectedRaces) {
				t.Errorf("unexpected number of races: got %d, want %d", len(response.Races), len(tc.expectedRaces))
			}

			assert.ElementsMatch(t, tc.expectedRaces, response.Races, "unexpected races")
		})
	}
}
