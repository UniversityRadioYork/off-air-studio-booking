package main

import (
	"runtime/debug"
	"sort"
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

	TrainingStudioTrained TrainingStatus = 1
	TrainingTrainer                      = 3

	TypeMeeting     BookingType = "Meeting"
	TypeOther                   = "Other"
	TypeEngineering             = "Engineering"
	TypeTraining                = "Training"
	TypeRecording               = "Recording"
)

var typeOrdering map[BookingType]int = map[BookingType]int{
	TypeTraining:    1,
	TypeRecording:   2,
	TypeEngineering: 3,
	TypeMeeting:     4,
	TypeOther:       5,
}

func GetNameOfUser(id int) string {
	name, err := myrSession.GetUserName(id)
	if err != nil {
		// TODO
		panic(err)
	}

	return name
}

func isManagement(userID int) bool {
	officerships, err := myrSession.GetUserOfficerships(userID)
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

	officerships, err := myrSession.GetUserOfficerships(userID)
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

func bookingsUserCanCreate(userID int) []BookingType {
	bookingTypes := []BookingType{TypeOther}

	// If Studio Trained -> Recording
	// If Trainer -> Training
	trainings, err := myrSession.GetUserTraining(userID)
	if err != nil {
		// TODO
		panic(err)
	}

	for _, training := range trainings {
		if training.StatusID == int(TrainingStudioTrained) {
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
	officerships, err := myrSession.GetUserOfficerships(userID)
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
