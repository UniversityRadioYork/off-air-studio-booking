package main

import (
	"fmt"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/UniversityRadioYork/myradio-go"
)

// Team numbers come from MyRadio
type Team int

// Officer numbers come from MyRadio
type Officer int

// BookingType relates particular constants to their values in the DB
type BookingType string

// TrainingStatus numbers come from MyRadio
type TrainingStatus int

const (
	TeamManagement  Team = 1
	TeamEngineering      = 7
	TeamComputing        = 8

	OfficerTrainingCoordinator Officer = 107

	TrainingStudioTrained  TrainingStatus = 1
	TrainingPodcastTrained                = 28
	TrainingTrainer                       = 3

	TypeMeeting                      BookingType = "Meeting"
	TypeOther                                    = "Other"
	TypeEngineering                              = "Engineering"
	TypeTraining                                 = "Training"
	TypeRecording                                = "Recording"
	TypeTrainingAutoAddedFromMyRadio             = "ONLY_FOR_USE_IN_MYRADIO_TRAINING_SYNC"
)

// typeOrdering defines the order booking types will be sorted into for
// the user's selector
var typeOrdering map[BookingType]int = map[BookingType]int{
	TypeTraining:    1,
	TypeRecording:   2,
	TypeEngineering: 3,
	TypeMeeting:     4,
	TypeOther:       5,
}

// cacheInvalidationTime determines how long we'll keep things in the
// cache before checking MyRadio to see if it needs updating
const cacheInvalidationTime = time.Duration(-2*24) * time.Hour

type myRadioNameCacheObject struct {
	Name      string
	CacheTime time.Time
}

var myRadioNameCache map[int]myRadioNameCacheObject = make(map[int]myRadioNameCacheObject)

// getNameOfUser will take a user ID and return their name, looked up from the cache,
// or from MyRadio if it is not known in the cache, or is too old
func getNameOfUser(id int) string {
	if cacheObject, ok := myRadioNameCache[id]; ok {
		if !cacheObject.CacheTime.Before(time.Now().Add(cacheInvalidationTime)) {
			return cacheObject.Name
		}
	}

	name, err := myrSession.GetUserName(id)
	if err != nil {
		// TODO
		panic(err)
	}

	myRadioNameCache[id] = myRadioNameCacheObject{
		Name:      name,
		CacheTime: time.Now(),
	}

	return name
}

type myRadioOfficershipCacheObject struct {
	Officerships []myradio.Officership
	CacheTime    time.Time
}

var myRadioOfficershipsCache map[int]myRadioOfficershipCacheObject = make(map[int]myRadioOfficershipCacheObject)

// getOfficerships will return officerships for a user, either from the cache or MyRadio
func getOfficerships(userID int) ([]myradio.Officership, error) {
	if cacheObject, ok := myRadioOfficershipsCache[userID]; ok {
		if !cacheObject.CacheTime.Before(time.Now().Add(cacheInvalidationTime)) {
			return cacheObject.Officerships, nil
		}
	}

	officerships, err := myrSession.GetUserOfficerships(userID)
	if err != nil {
		return nil, err
	}

	myRadioOfficershipsCache[userID] = myRadioOfficershipCacheObject{
		Officerships: officerships,
		CacheTime:    time.Now(),
	}
	return officerships, nil
}

type myRadioTrainingsCacheObject struct {
	Trainings []myradio.Training
	CacheTime time.Time
}

var myRadioTrainingsCache map[int]myRadioTrainingsCacheObject = make(map[int]myRadioTrainingsCacheObject)

// getTrainings will return the user's trainings, either from cache or MyRadio
func getTrainings(userID int) ([]myradio.Training, error) {
	if cacheObject, ok := myRadioTrainingsCache[userID]; ok {
		if !cacheObject.CacheTime.Before(time.Now().Add(cacheInvalidationTime)) {
			return cacheObject.Trainings, nil
		}
	}

	trainings, err := myrSession.GetUserTraining(userID)
	if err != nil {
		return nil, err
	}

	myRadioTrainingsCache[userID] = myRadioTrainingsCacheObject{
		Trainings: trainings,
		CacheTime: time.Now(),
	}
	return trainings, nil
}

// isManagement will return if a user is on management
// this gives them permissions, such as deleting all events
func isManagement(userID int) bool {
	officerships, err := getOfficerships(userID)
	if err != nil {
		// TODO
		panic(err)
	}

	for _, officership := range officerships {
		if officership.TillDateRaw != "" {
			continue
		}

		if officership.Officer.Team.TeamID == uint(TeamManagement) {
			// Management
			return true
		}
	}

	return false
}

// isComputing returns if a user has computing officer permissions,
// for working with cache-related endpoints, such as flushing
func isComputing(userID int) bool {
	officerships, err := getOfficerships(userID)
	if err != nil {
		// TODO
		panic(err)
	}

	for _, officership := range officerships {
		if officership.TillDateRaw != "" {
			continue
		}

		if officership.Officer.Team.TeamID == TeamComputing {
			return true
		}

	}

	return false

}

// hasPermissionToDelete works out if a user can delete a particular event,
// such as the TC being able to delete all training events
func hasPermissionToDelete(userID int, eventID int) bool {
	var event Event
	db.QueryRow("SELECT * FROM events WHERE event_id = $1", eventID).Scan(
		&event.ID, &event.Type, &event.Title, &event.User, &event.StartTime, &event.EndTime)

	// you can delete your own
	if userID == event.User {
		return true
	}

	// management can delete all
	if isManagement(userID) {
		return true
	}

	officerships, err := getOfficerships(userID)
	if err != nil {
		// TODO
		panic(err)
	}

	for _, officership := range officerships {
		if officership.TillDateRaw != "" {
			continue
		}

		if event.Type == TypeEngineering && (officership.Officer.Team.TeamID == TeamEngineering || officership.Officer.Team.TeamID == TeamComputing) {
			// Engineering Type Events
			return true
		}

		if event.Type == TypeTraining && officership.Officer.OfficerID == int(OfficerTrainingCoordinator) {
			// Training
			return true
		}
	}

	return false
}

// canClaimEventForStation determins if a user can remove a personal name
// from an event, and turn it into a station-wide event
func canClaimEventForStation(userID int, eventID int) bool {
	if !isManagement(userID) {
		return false
	}

	var event Event
	db.QueryRow("SELECT * FROM events WHERE event_id = $1", eventID).Scan(
		&event.ID, &event.Type, &event.Title, &event.User, &event.StartTime, &event.EndTime)

	if event.Type != TypeOther {
		return false
	}

	return strings.HasSuffix(event.Title, fmt.Sprintf("- %s", getNameOfUser(event.User)))

}

// bookingUserCanCreate returns an ordered list for the types
// of booking a user can create, based on their officerships and
// trainings
func bookingsUserCanCreate(userID int) []BookingType {
	bookingTypes := []BookingType{TypeOther}

	// If Studio Trained -> Recording
	// If Trainer -> Training
	trainings, err := getTrainings(userID)
	if err != nil {
		// TODO
		panic(err)
	}

	for _, training := range trainings {
		if training.StatusID == int(TrainingStudioTrained) || training.StatusID == TrainingPodcastTrained {
			bookingTypes = append(bookingTypes, TypeRecording)
			continue
		}

		if training.StatusID == TrainingTrainer {
			bookingTypes = append(bookingTypes, TypeTraining)
			continue
		}
	}

	// If Committee -> Meeting
	// If Tech -> Engineering
	officerships, err := getOfficerships(userID)
	if err != nil {
		// TODO
		panic(err)
	}

	for _, officership := range officerships {
		if officership.TillDateRaw != "" {
			continue
		}

		bookingTypes = append(bookingTypes, TypeMeeting)

		if officership.Officer.Team.TeamID == TeamEngineering || officership.Officer.Team.TeamID == TeamComputing {
			bookingTypes = append(bookingTypes, TypeEngineering)
		}
	}

	sort.SliceStable(bookingTypes, func(i, j int) bool {
		return typeOrdering[bookingTypes[i]] < typeOrdering[bookingTypes[j]]
	})

	return bookingTypes

}

// getBuildCommit allows us to see the Git commit in the app,
// as a way of having version numbers
func getBuildCommit() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				return setting.Value[:7]
			}
		}
	}
	return ""
}

var weekNamesCache map[string]string = make(map[string]string)
var weekNameCacheSetTime time.Time = time.Now()

// updateWeekNamesCache will cache week names from MyRadio.
// it will need to format them in the way the calendar displays dates
// NOTE: this uses a weird dash character
// NOTE: it uses three letter abbreviations for months, except September, which is Sept
func updateWeekNamesCache() {
	terms, err := myrSession.GetAllTerms()
	if err != nil {
		panic(err)
	}

	for _, term := range terms {
		for weekNo, weekName := range term.WeekNames {
			weekMonday := term.StartTime().Add(time.Duration(weekNo*7*24*60) * time.Minute)
			weekSunday := weekMonday.Add(6 * 24 * 60 * time.Minute)

			weekString := strconv.Itoa(weekMonday.Day())
			if weekMonday.Month() != weekSunday.Month() {
				weekString = weekString + " " + weekMonday.Month().String()[:3]
				if weekMonday.Month() == time.September {
					weekString = weekString + "t"
				}
			}

			weekString = weekString + " – " + strconv.Itoa(weekSunday.Day()) + " " + weekSunday.Month().String()[:3]
			if weekSunday.Month() == time.September {
				weekString = weekString + "t"
			}
			weekString = weekString + " " + strconv.Itoa(weekMonday.Year())

			weekNamesCache[weekString] = weekName
		}
	}

	weekNameCacheSetTime = time.Now()
}

// getWeekNames will return the week names, and possibly update the cache if
// it is old
func getWeekNames() map[string]string {
	if len(weekNamesCache) == 0 {
		updateWeekNamesCache()
	}

	if weekNameCacheSetTime.Before(time.Now().Add(cacheInvalidationTime)) {
		go updateWeekNamesCache()
	}

	return weekNamesCache
}
