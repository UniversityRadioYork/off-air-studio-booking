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

type Team int
type Officer int
type BookingType string
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

var typeOrdering map[BookingType]int = map[BookingType]int{
	TypeTraining:    1,
	TypeRecording:   2,
	TypeEngineering: 3,
	TypeMeeting:     4,
	TypeOther:       5,
}

type myRadioNameCacheObject struct {
	Name      string
	cacheTime time.Time
}

const cacheInvalidationTime = time.Duration(-2*24) * time.Hour

var myRadioNameCache map[int]myRadioNameCacheObject = make(map[int]myRadioNameCacheObject)

func GetNameOfUser(id int) string {
	if cacheObject, ok := myRadioNameCache[id]; ok {
		if !cacheObject.cacheTime.Before(time.Now().Add(cacheInvalidationTime)) {
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
		cacheTime: time.Now(),
	}

	return name
}

type myRadioOfficershipCacheObject struct {
	Officerships []myradio.Officership
	cacheTime    time.Time
}

var myRadioOfficershipsCache map[int]myRadioOfficershipCacheObject = make(map[int]myRadioOfficershipCacheObject)

func myRadioGetOfficerships(userID int) ([]myradio.Officership, error) {
	if cacheObject, ok := myRadioOfficershipsCache[userID]; ok {
		if !cacheObject.cacheTime.Before(time.Now().Add(cacheInvalidationTime)) {
			return cacheObject.Officerships, nil
		}
	}

	officerships, err := myrSession.GetUserOfficerships(userID)
	if err != nil {
		return nil, err
	}

	myRadioOfficershipsCache[userID] = myRadioOfficershipCacheObject{
		Officerships: officerships,
		cacheTime:    time.Now(),
	}
	return officerships, nil
}

type myRadioTrainingsCacheObject struct {
	Trainings []myradio.Training
	cacheTime time.Time
}

var myRadioTrainingsCache map[int]myRadioTrainingsCacheObject = make(map[int]myRadioTrainingsCacheObject)

func myRadioGetTrainings(userID int) ([]myradio.Training, error) {
	if cacheObject, ok := myRadioTrainingsCache[userID]; ok {
		if !cacheObject.cacheTime.Before(time.Now().Add(cacheInvalidationTime)) {
			return cacheObject.Trainings, nil
		}
	}

	trainings, err := myrSession.GetUserTraining(userID)
	if err != nil {
		return nil, err
	}

	myRadioTrainingsCache[userID] = myRadioTrainingsCacheObject{
		Trainings: trainings,
		cacheTime: time.Now(),
	}
	return trainings, nil
}

func isManagement(userID int) bool {
	officerships, err := myRadioGetOfficerships(userID)
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

func hasPermissionToDelete(userID int, eventID int) bool {
	var event Event
	db.QueryRow("SELECT * FROM events WHERE event_id = $1", eventID).Scan(
		&event.ID, &event.Type, &event.Title, &event.User, &event.StartTime, &event.EndTime)

	if userID == event.User {
		return true
	}

	if isManagement(userID) {
		return true
	}

	officerships, err := myRadioGetOfficerships(userID)
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

	return strings.HasSuffix(event.Title, fmt.Sprintf("- %s", GetNameOfUser(event.User)))

}

func bookingsUserCanCreate(userID int) []BookingType {
	bookingTypes := []BookingType{TypeOther}

	// If Studio Trained -> Recording
	// If Trainer -> Training
	trainings, err := myRadioGetTrainings(userID)
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
	officerships, err := myRadioGetOfficerships(userID)
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

func getWeekNames() map[string]string {
	if len(weekNamesCache) == 0 {
		updateWeekNamesCache()
	}

	if weekNameCacheSetTime.Before(time.Now().Add(cacheInvalidationTime)) {
		go updateWeekNamesCache()
	}

	return weekNamesCache
}

func hasComputingPermission(userID int) bool {
	officerships, err := myRadioGetOfficerships(userID)
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
