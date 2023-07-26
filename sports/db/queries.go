package db

const (
	eventsList = "list"
)

func getEventsQueries() map[string]string {
	return map[string]string{
		eventsList: `
			SELECT 
				id, 
				meeting_id, 
				name, 
				visible, 
				advertised_start_time 
			FROM events
		`,
	}
}
