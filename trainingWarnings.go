package main

import (
	"fmt"
	"time"
)

// trainingWarning represents the underlying structure of a warning, giving the affected user,
// time conflict and the ID of the conflicting booking
type trainingWarning struct {
	UserID       int
	TrainingTime time.Time
	ClashID      int
}

// trainingWarnings is the array of warnings. We use their index in this array
// as an identifier
var trainingWarnings []trainingWarning = []trainingWarning{}

// removeOutdatedWarnings removes any warnings with a conflict time more than one hour
// ago. This runs as part of the myRadioSync
func removeOutdatedWarnings() {
	var updatedTrainingWarnings []trainingWarning = []trainingWarning{}

	for _, warning := range trainingWarnings {
		if warning.TrainingTime.After(time.Now().Add(-time.Hour)) {
			updatedTrainingWarnings = append(updatedTrainingWarnings, warning)
		}
	}

	trainingWarnings = updatedTrainingWarnings
}

// findExistingWarning will return the index in the trainingWarnings array
// for a warning given by the userID and target training time is pertains to
func findExistingWarning(targetUserID int, targetTime time.Time) int {
	for idx, warning := range trainingWarnings {
		if warning.UserID == targetUserID && warning.TrainingTime == targetTime {
			return idx
		}
	}
	return -1
}

// deleteWarningByIndex works based on the index in the trainingWarnings array
func deleteWarningByIndex(idx int) {
	trainingWarnings = append(trainingWarnings[:idx], trainingWarnings[idx+1:]...)
}

// userFacingWarning is the information passed to the frontend, containing the viewable text
// of the warning, and the ID of the associated clashing event, so the user can see it
type userFacingWarning struct {
	WarningText string
	ClashID     int
}

// createUserFacingWarnings creates an array of userFacingWarning objects
// based on visiting user's permissions
func createUserFacingWarnings(userID int) []userFacingWarning {
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

	return warnings
}
