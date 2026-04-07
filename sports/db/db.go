package db

import (
	"math/rand"
	"time"

	"syreclabs.com/go/faker"
)

var eventTypes = []string{"football", "tennis", "basketball", "cricket", "rugby"}

func (e *eventsRepo) seed() error {
	statement, err := e.db.Prepare(`CREATE TABLE IF NOT EXISTS sport_events (id INTEGER PRIMARY KEY, name TEXT, event_type TEXT, visible INTEGER, advertised_start_time DATETIME)`)
	if err == nil {
		_, err = statement.Exec()
	}

	for i := 1; i <= 50; i++ {
		statement, err = e.db.Prepare(`INSERT OR IGNORE INTO sport_events(id, name, event_type, visible, advertised_start_time) VALUES (?,?,?,?,?)`)
		if err == nil {
			_, err = statement.Exec(
				i,
				faker.Team().Name()+" vs "+faker.Team().Name(),
				eventTypes[rand.Intn(len(eventTypes))],
				faker.Number().Between(0, 1),
				faker.Time().Between(time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 2)).Format(time.RFC3339),
			)
		}
	}

	return err
}
