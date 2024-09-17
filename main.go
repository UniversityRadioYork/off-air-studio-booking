/**
URY Off Air Studio Booking App
*/

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/UniversityRadioYork/myradio-go"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
)

var db *sql.DB
var myrSession *myradio.Session
var cookiestore = sessions.NewCookieStore([]byte("key"))

type CtxKey string

const UserCtxKey CtxKey = "user"

const AuthRealm string = "ury-off-air-bookings"

func initDB() {
	// Replace with your PostgreSQL connection string
	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		os.Getenv("DBUSER"), os.Getenv("DBPASS"), os.Getenv("DBNAME"), os.Getenv("DBHOST"), os.Getenv("DBPORT"))
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
	ID             int `json:"id"`
	Type           BookingType
	Title          string `json:"title"`
	User           int
	Start          string `json:"start"`
	End            string `json:"end"`
	StartTime      time.Time
	EndTime        time.Time
	NoNameAttached bool `json:"noNameAttached"`
	Repeat         int  `json:"repeat"`
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

func favicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "favicon.ico")
}

func info(w http.ResponseWriter, r *http.Request) {
	createTypes := bookingsUserCanCreate(r.Context().Value(UserCtxKey).(int))
	name := GetNameOfUser(r.Context().Value(UserCtxKey).(int))
	commit := getBuildCommit()

	json, err := json.Marshal(struct {
		CreateTypes                []BookingType
		Name                       string
		CommitHash                 string
		UserCanCreateUnnamedEvents bool
		WeekNames                  map[string]string
	}{
		CreateTypes:                createTypes,
		Name:                       name,
		CommitHash:                 commit,
		UserCanCreateUnnamedEvents: isManagement(r.Context().Value(UserCtxKey).(int)),
		WeekNames:                  getWeekNames(),
	})

	if err != nil {
		// TODO
		panic(err)
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(json)
}

func main() {
	initDB()

	var err error
	myrSession, err = myradio.NewSessionFromKeyFileForServer(os.Getenv("MYRADIO_API_SERVER"))
	if err != nil {
		// TODO
		panic(err)
	}

	go myRadioTrainingSync()

	r := mux.NewRouter()

	r.HandleFunc("/", indexPage).Methods("GET")
	r.HandleFunc("/main.js", js).Methods("GET")
	r.HandleFunc("/create", createEvent).Methods("POST")
	r.HandleFunc("/delete/{id}", deleteEvent).Methods("DELETE")
	r.HandleFunc("/claim/{id}", claimEventForStation).Methods("PUT")
	r.HandleFunc("/canModify/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		eventID := vars["id"]
		id, err := strconv.Atoi(eventID)
		if err != nil {
			// TODO
			panic(err)
		}

		userID := r.Context().Value(UserCtxKey).(int)

		json, err := json.Marshal(struct {
			Delete          bool
			ClaimForStation bool
		}{
			Delete:          hasPermissionToDelete(userID, id),
			ClaimForStation: canClaimEventForStation(userID, id),
		})

		if err != nil {
			// TODO
			panic(err)
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(json)
	}).Methods("GET")
	r.HandleFunc("/get", getEvents).Methods("GET")
	r.HandleFunc("/auth", auth)
	r.HandleFunc("/info", info).Methods("GET")
	r.HandleFunc("/logout", logout)
	r.HandleFunc("/favicon.ico", favicon).Methods("GET")

	r.HandleFunc("/flush", func(w http.ResponseWriter, r *http.Request) {

		if !hasComputingPermission(r.Context().Value(UserCtxKey).(int)) {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}

		encodedEventsCache = ""
		myRadioNameCache = make(map[int]myRadioNameCacheObject)
		myRadioOfficershipsCache = make(map[int]myRadioOfficershipCacheObject)
		myRadioTrainingsCache = make(map[int]myRadioTrainingsCacheObject)
		weekNamesCache = make(map[string]string)

	}).Methods("GET")

	r.HandleFunc("/cacheview", func(w http.ResponseWriter, r *http.Request) {

		if !hasComputingPermission(r.Context().Value(UserCtxKey).(int)) {
			http.Error(w, "Forbidden", http.StatusForbidden)
		}

		var events interface{}

		toDecode := encodedEventsCache
		if toDecode == "" {
			toDecode = "{}"
		}

		err := json.Unmarshal([]byte(toDecode), &events)
		if err != nil {
			// TODO
			panic(err)
		}

		d, err := json.Marshal([]interface{}{
			events,
			myRadioNameCache,
			myRadioOfficershipsCache,
			myRadioTrainingsCache,
			weekNamesCache,
		})
		if err != nil {
			// TODO
			panic(err)
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(d)
	}).Methods("GET")

	http.Handle("/", AuthHandler(r))
	if err = http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
