package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

var ErrClash error = fmt.Errorf("clashing events")

func addEvent(event Event) error {
	// Find clashes
	rows, err := db.Query(
		"SELECT * FROM events WHERE (start_time >= $1 AND start_time < $2) OR (end_time > $1 AND end_time <= $2) OR (start_time < $1 AND end_time > $2)",
		event.StartTime, event.EndTime)
	if err != nil {
		return err
	}

	defer rows.Close()

	if rows.Next() {
		return ErrClash
	}

	if event.Title == "" {
		event.Title = fmt.Sprintf("%s - %s", event.Type, event.User)
	} else {
		event.Title = fmt.Sprintf("%s - %s", event.Title, event.User)
	}

	// Insert the event into the database
	err = db.QueryRow(
		"INSERT INTO events(event_type, event_title, user_id, start_time, end_time) VALUES($1, $2, $3, $4, $5) RETURNING event_id",
		event.Type, event.Title, 1, event.StartTime, event.EndTime).Scan(&event.ID)

	return err
}

func createEvent(w http.ResponseWriter, r *http.Request) {
	// Check user's permissions and validate their identity
	// Implement your authorization logic here

	// Parse the request body and create a new event
	// Parse event data from the request body
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
		// TODO
	}
	var event Event
	json.Unmarshal(body, &event)

	event.parseTimes()

	err = addEvent(event)

	if err != nil {
		if errors.Is(err, ErrClash) {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, err)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, err)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("{\"status\": \"OK\"}"))
}

func deleteEvent(w http.ResponseWriter, r *http.Request) {
	// Check user's permissions and validate their identity

	// Delete the event from the database
	vars := mux.Vars(r)
	eventID := vars["id"]
	_, err := db.Exec("DELETE FROM events WHERE event_id=$1", eventID)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
