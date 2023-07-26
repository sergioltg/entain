package service

import (
	"context"
	"git.neds.sh/matty/entain/racing/proto/racing"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sort"
	"testing"
	"time"
)

// MockRacesRepo is a mock implementation of the db.RacesRepo interface.
type MockRacesRepo struct{}

func (m *MockRacesRepo) Init() error {
	return nil
}

func (m *MockRacesRepo) List(filter *racing.ListRacesRequestFilter, orderBy []*racing.ListRacesRequestOrderBy) ([]*racing.Race, error) {
	// Mock the behavior here and return a predefined response.
	// For simplicity, we'll return a predefined list of races.
	races := getAllTestData()

	var filteredRaces []*racing.Race
	for _, race := range races {
		if filter != nil {
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
				var result = false
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
		}

		// Add the race to the filtered list if it passes the filter criteria
		filteredRaces = append(filteredRaces, race)
	}

	if orderBy != nil {
		// Sort the filtered races considering the request order by
		sort.Slice(filteredRaces, func(i, j int) bool {
			for _, orderByClause := range orderBy {
				condition := compareField(orderByClause, races[i], races[j])
				if condition < 0 {
					return false
				} else if condition > 0 {
					return true
				}
			}
			// If all fields are equal, preserve the original order
			return i < j
		})
	}

	return filteredRaces, nil
}

// Compare a field in the race using racing.ListRacesRequestOrderBy considering the direction ASC or DESC
func compareField(orderBy *racing.ListRacesRequestOrderBy, a *racing.Race, b *racing.Race) int {
	if orderBy.FieldName == "advertisedStartTime" {
		time1 := a.AdvertisedStartTime.AsTime()
		time2 := b.AdvertisedStartTime.AsTime()
		if time1.Equal(time2) {
			return 0
		}
		if orderBy.Direction == racing.OrderByDirection_DESC {
			if time2.After(time1) {
				return -1
			} else {
				return 1
			}
		} else {
			if time1.After(time2) {
				return -1
			} else {
				return 1
			}
		}
	}
	return 0
}

func TestRacingService_ListRaces(t *testing.T) {
	// Define test cases with different inputs and expected outputs
	testCases := []struct {
		name          string
		filter        *racing.ListRacesRequestFilter
		orderBy       []*racing.ListRacesRequestOrderBy
		expectedRaces []*racing.Race
		expectedErr   bool
	}{
		{
			name:   "NoFilter",
			filter: &racing.ListRacesRequestFilter{
				// empty filter
			},
			expectedRaces: getAllTestData(),
			expectedErr:   false,
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
		{
			name: "OrderByAdvertisedStartTimeDescending",
			orderBy: []*racing.ListRacesRequestOrderBy{
				{FieldName: "advertisedStartTime",
					Direction: racing.OrderByDirection_DESC,
				},
			},
			expectedRaces: []*racing.Race{
				{
					Id:                  3,
					MeetingId:           8,
					Name:                "Rhode Island ghosts",
					Number:              3,
					Visible:             false,
					AdvertisedStartTime: timestamppb.New(time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
				{
					Id:                  2,
					MeetingId:           1,
					Name:                "Connecticut griffins",
					Number:              12,
					Visible:             true,
					AdvertisedStartTime: timestamppb.New(time.Date(2023, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
				{Id: 1,
					MeetingId:           5,
					Name:                "North Dakota foes",
					Number:              12,
					Visible:             false,
					AdvertisedStartTime: timestamppb.New(time.Date(2022, 7, 15, 12, 0, 0, 0, time.UTC)),
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
				Filter:  tc.filter,
				OrderBy: tc.orderBy,
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

			assert.Equal(t, tc.expectedRaces, response.Races, "unexpected races")
		})
	}
}

func getAllTestData() []*racing.Race {
	return []*racing.Race{
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
}
