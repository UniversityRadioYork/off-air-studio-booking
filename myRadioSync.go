package main

import "time"

func myRadioTrainingSync() {
	/**

	The calendar should sync with MyRadio training sessions to make
	sure every training is booked on the calendar.

	It'll check every 15 minutes between 9am and 10pm.

	If it finds a training session on MyRadio in the next 2 weeks not on the calendar,
	it'll add it, though if there's a conflict, it'll...I guess email
	the person.

	**/

	for {
		if time.Now().Hour() < 9 || time.Now().Hour() >= 22 {
			time.Sleep(15 * time.Minute)
			continue
		}

		// Do the Sync
		// Ask MyRadio for training sessions in next 14 days

		// Iterate Over Training

		// Is training in the calendar? Yes good.
		// No - is it free? Yes - add the session.
		// No - email the person.

		time.Sleep(15 * time.Minute)
	}
}
