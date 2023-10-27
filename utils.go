package main

import "fmt"

type Team int
type BookingType string

const (
	TeamManagement  Team = 1
	TeamEngineering      = 7
	TeamComputing        = 8

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

func hasPermissionToDelete(userID int, eventID int) bool {
	var event Event
	db.QueryRow("SELECT * FROM events WHERE event_id = $1", eventID).Scan(
		&event.ID, &event.Type, &event.Title, &event.User, &event.StartTime, &event.EndTime)

	fmt.Println(event)
	if userID == event.User {
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

		if officership.Officer.Team.TeamID == 1 {
			// Management
			return true
		}

		if event.Type == TypeEngineering && (officership.Officer.Team.TeamID == TeamEngineering || officership.Officer.Team.TeamID == TeamComputing) {
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

		bookingTypes = append(bookingTypes, TypeMeeting)

		if officership.Officer.Team.TeamID == TeamEngineering || officership.Officer.Team.TeamID == TeamComputing {
			bookingTypes = append(bookingTypes, TypeEngineering)
		}
	}

	return bookingTypes

}
