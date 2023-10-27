package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
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
	Type      BookingType
	Title     string `json:"title"`
	User      int
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

func auth(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		http.ServeFile(w, r, "login.html")
		return
	} else if r.Method == "POST" {
		session, _ := cookiestore.Get(r, AuthRealm)

		memberid := r.FormValue("memberid")
		if memberid == "" {
			http.Redirect(w, r, "/auth", http.StatusFound)
			return
		}

		var err error
		session.Values["memberid"], err = strconv.Atoi(memberid)
		if err != nil {
			panic(err)
		}
		session.Save(r, w)
		http.Redirect(w, r, "/", http.StatusFound)

	}
	fmt.Fprint(w, "hmmm")
}

func AuthHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/auth" {
			h.ServeHTTP(w, r)
			return
		}

		session, _ := cookiestore.Get(r, AuthRealm)
		if auth, ok := session.Values["memberid"].(int); !ok || auth == 0 {
			// redirect to auth
			http.Redirect(w, r, "/auth", http.StatusFound)
		} else {
			ctx := context.WithValue(context.Background(), UserCtxKey, auth)
			h.ServeHTTP(w, r.WithContext(ctx))
		}

	})
}

func main() {
	initDB()

	var err error
	myrSession, err = myradio.NewSessionFromKeyFile()
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
	r.HandleFunc("/canDelete/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		eventID := vars["id"]
		id, err := strconv.Atoi(eventID)
		if err != nil {
			// TODO
			panic(err)
		}

		w.Write([]byte(strconv.FormatBool(hasPermissionToDelete(r.Context().Value(UserCtxKey).(int), id))))
	}).Methods("GET")
	r.HandleFunc("/get", getEvents).Methods("GET")
	r.HandleFunc("/userCreateTypes", func(w http.ResponseWriter, r *http.Request) {
		json, err := json.Marshal(bookingsUserCanCreate(r.Context().Value(UserCtxKey).(int)))
		if err != nil {
			// TODO
			panic(err)
		}
		w.Header().Add("Content-Type", "application/json")
		w.Write(json)
	}).Methods("GET")
	r.HandleFunc("/auth", auth)
	r.HandleFunc("/name", func(w http.ResponseWriter, r *http.Request) {
		name := GetNameOfUser(r.Context().Value(UserCtxKey).(int))
		w.Write([]byte(name))
	})
	r.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		session, _ := cookiestore.Get(r, AuthRealm)
		session.Values["memberid"] = 0
		session.Save(r, w)
		http.Redirect(w, r, "https://ury.org.uk/myradio/MyRadio/logout", http.StatusFound)
	})

	http.Handle("/", AuthHandler(r))
	if err = http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
