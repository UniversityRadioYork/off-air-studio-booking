package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

var ErrClash error = fmt.Errorf("clashing events")
var ErrPermission error = fmt.Errorf("permission denied")

func addEvent(event Event) error {
	// Check the user can make this type of event
	allowed := false
	for _, v := range bookingsUserCanCreate(event.User) {
		if v == event.Type {
			allowed = true
			break
		}
	}
	if !allowed {
		return ErrPermission
	}

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

	if !event.NoNameAttached {
		if event.Title == "" {
			event.Title = fmt.Sprintf("%s - %s", event.Type, GetNameOfUser(event.User))
		} else {
			event.Title = fmt.Sprintf("%s - %s", event.Title, GetNameOfUser(event.User))
		}
	}

	creatingUser := GetNameOfUser(event.User)

	if event.Type == TypeTrainingAutoAddedFromMyRadio {
		event.Type = TypeTraining
		creatingUser = "MyRadio Auto-Sync"
		event.Title = fmt.Sprintf("%s %s", event.Title, "📻") // add a radio emoji for things that are added because they're on myradio
	}

	// Insert the event into the database
	err = db.QueryRow(
		"INSERT INTO events(event_type, event_title, user_id, start_time, end_time) VALUES($1, $2, $3, $4, $5) RETURNING event_id",
		string(event.Type), event.Title, event.User, event.StartTime, event.EndTime).Scan(&event.ID)

	encodedEventsCache = ""

	log.Printf("%s created %s at %v (ID %v)\n", creatingUser, event.Title, event.Start, event.ID)

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

	var ok bool
	event.User, ok = r.Context().Value(UserCtxKey).(int)
	if !ok {
		// TODO
		panic(err)
	}

	event.parseTimes()

	if event.Type == TypeTrainingAutoAddedFromMyRadio {
		// this type is only for the automatic adding, not for users
		http.Error(w, "wrong training type", http.StatusBadRequest)
		return
	}

	firstStartTime := event.StartTime
	firstEndTime := event.EndTime

	for i := 0; i < event.Repeat; i++ {
		event.StartTime = firstStartTime.Add(time.Duration(i*24*7) * time.Hour)
		event.EndTime = firstEndTime.Add(time.Duration(i*24*7) * time.Hour)

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
	id, err := strconv.Atoi(eventID)
	if err != nil {
		// TODO
		panic(err)
	}

	if !hasPermissionToDelete(r.Context().Value(UserCtxKey).(int), id) {
		// TODO
		panic("TODO")
	}

	_, err = db.Exec("DELETE FROM events WHERE event_id=$1", id)

	encodedEventsCache = ""

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("%s deleted event %v", GetNameOfUser(r.Context().Value(UserCtxKey).(int)), id)

	w.WriteHeader(http.StatusNoContent)
}

func claimEventForStation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eventID := vars["id"]
	id, err := strconv.Atoi(eventID)
	if err != nil {
		// TODO
		panic(err)
	}

	if !canClaimEventForStation(r.Context().Value(UserCtxKey).(int), id) {
		panic("TODO")
	}

	var event Event
	db.QueryRow("SELECT * FROM events WHERE event_id = $1", eventID).Scan(
		&event.ID, &event.Type, &event.Title, &event.User, &event.StartTime, &event.EndTime)

	newTitle, _, _ := strings.Cut(event.Title, fmt.Sprintf("- %s", GetNameOfUser(event.User)))

	_, err = db.Exec("UPDATE events SET event_title = $1 WHERE event_id = $2", newTitle, id)
	if err != nil {
		panic(err)
	}

	encodedEventsCache = ""

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
