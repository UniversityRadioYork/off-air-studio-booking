package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

func addEvent(event Event) error {
	// Find clashes

	// Insert the event into the database
	err := db.QueryRow("INSERT INTO events(event_type, event_title, user_id, start_time, end_time) VALUES($1, $2, $3, $4, $5) RETURNING event_id",
		event.Type, event.Title, 1, event.StartTime, event.EndTime).Scan(&event.ID)

	return err
}

func createEvent(w http.ResponseWriter, r *http.Request) {
	// Check user's permissions and validate their identity
	// Implement your authorization logic here

	// Parse the request body and create a new event
	// var event Event
	// Parse event data from the request body
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
		// TODO
	}
	var event Event
	json.Unmarshal(body, &event)
	// fmt.Println(string(body))

	event.parseTimes()

	err = addEvent(event)

	if err != nil {
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
