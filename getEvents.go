package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type renderEvent struct {
	Event
	Color     string `json:"color"`
	TextColor string `json:"textColor"`
}

var typeToColor map[string]string = map[string]string{
	"Training":    "red",
	"Recording":   "blue",
	"Engineering": "green",
	"Meeting":     "purple",
	"Other":       "yellow",
}

func getEvents(w http.ResponseWriter, r *http.Request) {
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

	var events []renderEvent
	for rows.Next() {
		var event renderEvent
		if err := rows.Scan(&event.ID, &event.Type, &event.Title, &event.User, &event.StartTime, &event.EndTime); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		event.Color = typeToColor[event.Type]
		event.TextColor = "white"
		if event.Type == "Other" {
			event.TextColor = "black"
		}

		event.Start = event.StartTime.Format("2006-01-02T15:04")
		event.End = event.EndTime.Format("2006-01-02T15:04")

		events = append(events, event)
	}

	// Marshal the events to JSON and respond
	// Implement JSON marshaling and response
	json, err := json.Marshal(events)
	if err != nil {
		panic(err)
		// TODO
	}
	w.Header().Add("content-type", "application/json")
	w.Write(json)
}
