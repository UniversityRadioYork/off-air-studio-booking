package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var db *sql.DB

func initDB() {
	// Replace with your PostgreSQL connection string
	connectionString := "user=bookings password=bookings dbname=offairbooking sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		panic(err)
	}

	// Check if the database connection is successful
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to the database")
}

type Event struct {
	ID        int `json:"id"`
	Type      string
	Title     string `json:"title"`
	User      string
	Start     string `json:"start"`
	End       string `json:"end"`
	StartTime time.Time
	EndTime   time.Time
}

func (e *Event) parseTimes() {
	var err error
	e.StartTime, err = time.Parse("2006-01-02T15:04", e.Start)
	if err != nil {
		panic(err)
		// TODO
	}

	e.EndTime, err = time.Parse("2006-01-02T15:04", e.End)
	if err != nil {
		panic(err)
		// TODO
	}

}

func indexPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func js(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "main.js")
}

func main() {
	initDB()

	go myRadioTrainingSync()

	r := mux.NewRouter()
	r.HandleFunc("/", indexPage).Methods("GET")
	r.HandleFunc("/main.js", js).Methods("GET")
	r.HandleFunc("/create", createEvent).Methods("POST")
	r.HandleFunc("/delete/{id}", deleteEvent).Methods("DELETE")
	r.HandleFunc("/get", getEvents).Methods("GET")

	http.Handle("/", r)
	http.ListenAndServe(":8080", nil)
}
