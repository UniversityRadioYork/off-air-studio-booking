package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// renderEvent takes the Event and adds details we only need for the calendar
// NOTE: the calendar expects Color spelling in the JSON
type renderEvent struct {
	Event
	Colour     string `json:"color"`
	TextColour string `json:"textColor"`
}

// typeToColour maps event types to their colours. The calendar expects the
// colours to be in the API data, so we insert them here. These colours
// came from the colours in the original studio booking spreadsheet.
// NOTE: it is expected to match the key in the HTML
var typeToColour map[BookingType]string = map[BookingType]string{
	TypeTraining:    "red",
	TypeRecording:   "blue",
	TypeEngineering: "green",
	TypeMeeting:     "purple",
	TypeOther:       "yellow",
}

// encodedEventsCache saves us bugging the DB each time the page is loaded
// by saving the JSON response as a string
// IMPORTANT: whenever DB data is changed, this must be cleared
var encodedEventsCache string = ""

// getEventsHandler is the main handler for returning JSON containing the bookings information
// It looks +/- 3 months from present (when it calls the DB, not if it just calls
// encodedEventsCache)
func getEventsHandler(w http.ResponseWriter, r *http.Request) {
	if encodedEventsCache != "" {
		w.Header().Add("content-type", "application/json")
		w.Write([]byte(encodedEventsCache))
		return
	}

	// Calculate the date range for the past 3 and next 3 months
	now := time.Now()
	threeMonthsAgo := now.AddDate(0, -3, 0)
	threeMonthsFromNow := now.AddDate(0, 3, 0)

	// Query the events from the database based on the date range
	rows, err := db.Query("SELECT * FROM events WHERE start_time >= $1 AND end_time <= $2", threeMonthsAgo, threeMonthsFromNow)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}
	defer rows.Close()

	// We will need to turn each Event into a renderEvent by including the colours
	var events []renderEvent = []renderEvent{}
	for rows.Next() {
		var event renderEvent
		if err := rows.Scan(&event.ID, &event.Type, &event.Title, &event.User, &event.StartTime, &event.EndTime); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		event.Colour = typeToColour[BookingType(event.Type)]

		// TypeOther uses a yellow background, so we need contrasting text
		event.TextColour = "white"
		if event.Type == TypeOther {
			event.TextColour = "black"
		}

		event.parseTimes()

		events = append(events, event)
	}

	// Marshal the events to JSON and respond
	json, err := json.Marshal(events)
	if err != nil {
		panic(err)
		// TODO
	}

	encodedEventsCache = string(json)

	w.Header().Add("content-type", "application/json")
	w.Write(json)
}
