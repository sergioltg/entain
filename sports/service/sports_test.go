package service

import (
	"context"
	"git.neds.sh/matty/entain/sports/proto/sports"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
	"sort"
	"testing"
	"time"
)

// MockEventsRepo is a mock implementation of the db.EventsRepo interface.
type MockEventsRepo struct{}

func (m *MockEventsRepo) Init() error {
	return nil
}

func (m *MockEventsRepo) List(filter *sports.ListEventsRequestFilter, orderBy []*sports.ListEventsRequestOrderBy, currentDate time.Time) ([]*sports.Event, error) {
	// Mock the behavior here and return a predefined response.
	// For simplicity, we'll return a predefined list of events.
	events := getAllTestData()

	var filteredEvents []*sports.Event
	for _, event := range events {
		if filter != nil {
			if filter.GetVisibilityStatus() == sports.VisibilityStatus_VISIBLE && !event.GetVisible() {
				// Skip events that are not visible
				continue
			}

			if filter.GetVisibilityStatus() == sports.VisibilityStatus_HIDDEN && event.GetVisible() {
				// Skip events that are visible
				continue
			}

			if len(filter.MeetingIds) > 0 {
				// Skip events that don't match meeting ids
				var result = false
				for _, x := range filter.MeetingIds {
					if x == event.MeetingId {
						result = true
						break
					}
				}
				if !result {
					continue
				}
			}
		}

		// Add the event to the filtered list if it passes the filter criteria
		filteredEvents = append(filteredEvents, event)
	}

	if orderBy != nil {
		// Sort the filtered events considering the request order by
		sort.Slice(filteredEvents, func(i, j int) bool {
			for _, orderByClause := range orderBy {
				condition := compareField(orderByClause, events[i], events[j])
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

	return filteredEvents, nil
}

// Compare a field in the event using sports.ListEventsRequestOrderBy considering the direction ASC or DESC
func compareField(orderBy *sports.ListEventsRequestOrderBy, a *sports.Event, b *sports.Event) int {
	if orderBy.FieldName == "advertisedStartTime" {
		time1 := a.AdvertisedStartTime.AsTime()
		time2 := b.AdvertisedStartTime.AsTime()
		if time1.Equal(time2) {
			return 0
		}
		if orderBy.Direction == sports.OrderByDirection_DESC {
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

func (m *MockEventsRepo) Get(id int64, currentDate time.Time) (*sports.Event, error) {
	events := getAllTestData()
	for _, event := range events {
		if event.Id == id {
			return event, nil
		}
	}
	return nil, nil
}

func TestSportsService_ListEvents(t *testing.T) {
	// Define test cases with different inputs and expected outputs
	testCases := []struct {
		name           string
		filter         *sports.ListEventsRequestFilter
		orderBy        []*sports.ListEventsRequestOrderBy
		expectedEvents []*sports.Event
		expectedErr    bool
	}{
		{
			name:   "NoFilter",
			filter: &sports.ListEventsRequestFilter{
				// empty filter
			},
			expectedEvents: getAllTestData(),
			expectedErr:    false,
		},
		{
			name: "FilterByMeetingIDs",
			filter: &sports.ListEventsRequestFilter{
				MeetingIds: []int64{1},
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
			expectedErr: false,
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
			expectedErr: false,
		},
		{
			name: "FilterByVisibilityStatusHidden",
			filter: &sports.ListEventsRequestFilter{
				VisibilityStatus: sports.VisibilityStatus_HIDDEN,
			},
			expectedEvents: []*sports.Event{
				{Id: 1,
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
			expectedErr: false,
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
				{Id: 1,
					MeetingId:           5,
					Name:                "North Dakota foes",
					Visible:             false,
					Status:              "CLOSED",
					AdvertisedStartTime: timestamppb.New(time.Date(2022, 7, 15, 12, 0, 0, 0, time.UTC)),
				},
			},
			expectedErr: false,
		},
	}

	// Create a mock EventsRepo and pass it to the sportsService
	eventsRepo := &MockEventsRepo{}
	sportsSvc := NewSportsService(eventsRepo)

	// Run the test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare the request
			request := &sports.ListEventsRequest{
				Filter:  tc.filter,
				OrderBy: tc.orderBy,
			}

			// Call the method being tested
			response, err := sportsSvc.ListEvents(context.Background(), request)

			// Check for errors
			if (err != nil) != tc.expectedErr {
				t.Fatalf("unexpected error: %v", err)
			}

			// Compare the response events to the expected events
			if len(response.Events) != len(tc.expectedEvents) {
				t.Errorf("unexpected number of events: got %d, want %d", len(response.Events), len(tc.expectedEvents))
			}

			assert.Equal(t, tc.expectedEvents, response.Events, "unexpected events")
		})
	}
}

func TestSportsService_GetEvent(t *testing.T) {
	t.Run("GetById", func(t *testing.T) {
		eventsRepo := &MockEventsRepo{}
		sportsSvc := NewSportsService(eventsRepo)

		// Prepare the request
		request := &sports.GetEventRequest{
			Id: 2,
		}

		// Call the method being tested
		response, err := sportsSvc.GetEvent(context.Background(), request)

		// Check for errors
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		assert.Equal(t, response.Event.Id, request.GetId())
	})
}

func getEvents() map[int64]*sports.Event {
	events := map[int64]*sports.Event{
		1: {
			Id:                  1,
			MeetingId:           5,
			Name:                "North Dakota foes",
			Visible:             false,
			Status:              "CLOSED",
			AdvertisedStartTime: timestamppb.New(time.Date(2022, 7, 15, 12, 0, 0, 0, time.UTC)),
		},
		2: {
			Id:                  2,
			MeetingId:           1,
			Name:                "Connecticut griffins",
			Visible:             true,
			Status:              "OPEN",
			AdvertisedStartTime: timestamppb.New(time.Date(2023, 7, 15, 12, 0, 0, 0, time.UTC)),
		},
		3: {
			Id:                  3,
			MeetingId:           8,
			Name:                "Rhode Island ghosts",
			Visible:             false,
			Status:              "OPEN",
			AdvertisedStartTime: timestamppb.New(time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)),
		},
	}
	return events
}

func getAllTestData() []*sports.Event {
	return []*sports.Event{
		{Id: 1,
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
	}
}
