package main

import (
	"errors"
	"time"
)

func myRadioTrainingSync() {
	/**

	The calendar should sync with MyRadio training sessions to make
	sure every training is booked on the calendar.

	It'll check every 15 minutes between 9am and 10pm.

	If it finds a training session on MyRadio not on the calendar,
	it'll add it, though if there's a conflict, it'll...I guess email
	the person.

	**/

	for {
		if time.Now().Hour() < 9 || time.Now().Hour() >= 22 {
			time.Sleep(15 * time.Minute)
			continue
		}

		// Do the Sync
		// Ask MyRadio for training sessions
		trainings, err := myrSession.GetFutureTrainingSessions()
		if err != nil {
			// TODO
			panic(err)
		}

		// Iterate Over Training
		for _, trainingSession := range trainings {
			// Is training in the calendar? Yes good.
			// No - is it free? Yes - add the session.
			// No - email the person.

			var count int
			err = db.QueryRow("SELECT COUNT(*) FROM events WHERE start_time = $1 AND event_type = 'Training' AND user_id = $2",
				trainingSession.StartTime(), trainingSession.HostMemberID).Scan(&count)
			if err != nil {
				// TODO
				panic(err)
			}
			if count != 0 {
				continue
			}

			err = addEvent(Event{
				Type:      TypeTraining,
				User:      trainingSession.HostMemberID,
				StartTime: trainingSession.StartTime(),
				EndTime:   trainingSession.StartTime().Add(time.Hour),
			})

			if err == nil {
				continue
			}

			if !errors.Is(err, ErrClash) {
				// TODO
				panic(err)
			}

			// TODO Inform the User

		}

		time.Sleep(15 * time.Minute)
	}
}
