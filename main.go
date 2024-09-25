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

const AuthRealm string = "ury-off-air-bookings"

var cookiestore = sessions.NewCookieStore([]byte(AuthRealm))

type CtxKey string

const UserCtxKey CtxKey = "user"

type userFacingWarning struct {
	WarningText string
	ClashID     int
}

// initDB will create our connection to the database
// this uses the environment variables as described in the README
func initDB() {
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

// Event is the main representation of a booking
// NOTE: Start and End are strings, StartTime and EndTime are
// time.Time objects. The strings are expected by the JS, but
// for backend convenience, we use time.Time. Use event.parseTimes()
// to convert.
type Event struct {
	ID        int `json:"id"`
	Type      BookingType
	Title     string `json:"title"`
	User      int
	Start     string `json:"start"`
	End       string `json:"end"`
	StartTime time.Time
	EndTime   time.Time
}

const CalendarTimeFormat string = "2006-01-02T15:04"

// parseTimes will create StartTime/EndTime if Start/End
// is populated, or vice versa. If you create an Event object,
// you should call event.parseTimes()
func (e *Event) parseTimes() {
	if e.Start == "" {
		e.Start = e.StartTime.Format(CalendarTimeFormat)
		e.End = e.EndTime.Format(CalendarTimeFormat)
		return
	}

	var err error
	e.StartTime, err = time.Parse(CalendarTimeFormat, e.Start)
	if err != nil {
		panic(err)
		// TODO
	}

	e.EndTime, err = time.Parse(CalendarTimeFormat, e.End)
	if err != nil {
		panic(err)
		// TODO
	}

}

// indexPageHandler serves the calendar HTML
func indexPageHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

// jsHandler serves the calendar JS
func jsHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "main.js")
}

// faviconHandler serves the calendar icon
func faviconHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "favicon.ico")
}

// infoHandler serves information used by the calendar app, including
// what types of booking the user can create, and whether the user
// has permission to create events without attaching their personal
// name to it
func infoHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserCtxKey).(int)
	createTypes := bookingsUserCanCreate(userID)
	name := getNameOfUser(userID)
	commit := getBuildCommit()

	// Create Warnings
	var warnings []userFacingWarning = []userFacingWarning{}

	// 1. Warnings about you
	if isTrainer(userID) {
		for _, warning := range trainingWarnings {
			if warning.UserID == userID {
				warnings = append(warnings, userFacingWarning{WarningText: fmt.Sprintf(
					"You have a training session booked on MyRadio on %v, however there is a conflict on the calendar.",
					warning.TrainingTime.Format("Mon 02/01 at 15:04")),
					ClashID: warning.ClashID})
			}
		}
	}

	// 2. Is management or TC?
	if isManagement(userID) || isTrainingCoordinator(userID) {
		for _, warning := range trainingWarnings {
			if warning.UserID == userID {
				continue
			}

			warnings = append(warnings, userFacingWarning{WarningText: fmt.Sprintf(
				"%s has a training session booked on MyRadio on %v, however there is a conflict on the calendar.",
				getNameOfUser(warning.UserID), warning.TrainingTime.Format("Mon 02/01 at 15:04")), ClashID: warning.ClashID})
		}
	}

	json, err := json.Marshal(struct {
		CreateTypes                []BookingType
		Name                       string
		CommitHash                 string
		UserCanCreateUnnamedEvents bool
		WeekNames                  map[string]string
		Warnings                   []userFacingWarning
	}{
		CreateTypes:                createTypes,
		Name:                       name,
		CommitHash:                 commit,
		UserCanCreateUnnamedEvents: isManagement(r.Context().Value(UserCtxKey).(int)),
		WeekNames:                  getWeekNames(),
		Warnings:                   warnings,
	})

	if err != nil {
		// TODO
		panic(err)
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(json)
}

// canModifyHandler will be called when a user goes to modify a booking
// and will return if they can delete that booking, or they can
// claim the event as a station event (essentially, removing the user's
// personal name from it)
func canModifyHandler(w http.ResponseWriter, r *http.Request) {
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
}

// cacheFlushHandler is for people with computing officership
// permissions to wipe clear all the caches within the app
func cacheFlushHandler(w http.ResponseWriter, r *http.Request) {

	if !isComputing(r.Context().Value(UserCtxKey).(int)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	encodedEventsCache = ""
	myRadioNameCache = make(map[int]myRadioNameCacheObject)
	myRadioOfficershipsCache = make(map[int]myRadioOfficershipCacheObject)
	myRadioTrainingsCache = make(map[int]myRadioTrainingsCacheObject)
	weekNamesCache = make(map[string]string)
	trainingWarnings = make([]trainingWarning, 0)

}

// cacheViewHandler allows people with computing officer permissions
// to view all the information cached within the app
func cacheViewHandler(w http.ResponseWriter, r *http.Request) {

	if !isComputing(r.Context().Value(UserCtxKey).(int)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
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
		weekNameCacheSetTime,
		trainingWarnings,
	})
	if err != nil {
		// TODO
		panic(err)
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(d)
}

func main() {
	initDB()

	var err error
	myrSession, err = myradio.NewSessionFromKeyFileForServer(os.Getenv("MYRADIO_API_SERVER"))
	if err != nil {
		// TODO
		panic(err)
	}

	// start the MyRadio training session sync daemon
	go myRadioTrainingSync()

	r := mux.NewRouter()

	// define all our endpoints
	r.HandleFunc("/", indexPageHandler).Methods("GET")
	r.HandleFunc("/main.js", jsHandler).Methods("GET")
	r.HandleFunc("/create", createEventHandler).Methods("POST")
	r.HandleFunc("/delete/{id}", deleteEventHandler).Methods("DELETE")
	r.HandleFunc("/claim/{id}", claimEventForStationHandler).Methods("PUT")
	r.HandleFunc("/canModify/{id}", canModifyHandler).Methods("GET")
	r.HandleFunc("/get", getEventsHandler).Methods("GET")
	r.HandleFunc("/auth", auth)
	r.HandleFunc("/info", infoHandler).Methods("GET")
	r.HandleFunc("/logout", logout)
	r.HandleFunc("/favicon.ico", faviconHandler).Methods("GET")
	r.HandleFunc("/flush", cacheFlushHandler).Methods("GET")
	r.HandleFunc("/cacheview", cacheViewHandler).Methods("GET")

	// route all our endpoints through the authentication
	http.Handle("/", AuthHandler(r))
	if err = http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
