package main

import "strconv"

type Team int
type BookingType string

const (
	TeamManagement  Team = 1
	TeamEngineering      = 2
	TeamComputing        = 3

	TypeMeeting     BookingType = "Meeting"
	TypeOther                   = "Other"
	TypeEngineering             = "Engineering"
	TypeTraining                = "Training"
	TypeRecording               = "Recording"
)

func GetNameOfUser(id int) string {
	name, err := myrSession.GetUserName(id)
	if err != nil {
		// TODO
		panic(err)
	}

	return name
}

func hasPermissionToDelete(userID int, event Event) bool {
	if strconv.Itoa(userID) == event.User {
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

		if officership.TeamId == 1 {
			// Management
			return true
		}

		if event.Type == TypeEngineering && (officership.TeamId == TeamEngineering || officership.TeamId == TeamComputing) {
			// Engineering Type Events
			return true
		}
	}

	return false
}

func bookingsUserCanCreate(userID int) []BookingType {
	bookingTypes := []BookingType{TypeOther}

	// If Studio Trained -> Recording
	// If Trainer -> Training

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

		if officership.TeamId == TeamEngineering || officership.TeamId == TeamComputing {
			bookingTypes = append(bookingTypes, TypeEngineering)
		}
	}

	return bookingTypes

}
