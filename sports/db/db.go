package db

import (
	"time"

	"syreclabs.com/go/faker"
)

func (r *eventsRepo) seed() error {
	statement, err := r.db.Prepare(`CREATE TABLE IF NOT EXISTS events (id INTEGER PRIMARY KEY, meeting_id INTEGER, name TEXT, visible INTEGER, advertised_start_time DATETIME)`)
	if err == nil {
		_, err = statement.Exec()
	}

	for i := 1; i <= 100; i++ {
		statement, err = r.db.Prepare(`INSERT OR IGNORE INTO events(id, meeting_id, name, visible, advertised_start_time) VALUES (?,?,?,?,?)`)
		if err == nil {
			_, err = statement.Exec(
				i,
				faker.Number().Between(1, 10),
				faker.Lorem().Sentence(20),
				faker.Number().Between(0, 1),
				faker.Time().Between(time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 2)).Format(time.RFC3339),
			)
		}
	}

	return err
}
