package main

import (
	"log"
	"net/http"
	"github.com/jmoiron/sqlx"
)

// Does all the heavy lifting of populating the local db with everything from Fitocracy
func PopulateDB(db *sqlx.DB, username string, password string){
	ensureSchema(db)

	err, fitocracyUserId, client := getClient(username, password)
	if nil != err {
		log.Fatal(err)
	}
	log.Printf("Looking up user %s (%d)\n", username, fitocracyUserId)
	err, user := GetUserByFitocracyId(db, fitocracyUserId)
	if nil != err {
		log.Println(err)
		user.FitocracyId = fitocracyUserId
		user.FitocracyUsername = username
		_, err := db.NamedExec("INSERT INTO users(fitocracy_id, fitocracy_username) VALUES(:fitocracy_id, :fitocracy_username)", &user)
		if nil != err {
			log.Fatal(err)
		}
	}

	err, activities := getActivities(client, fitocracyUserId)
	if nil != err {
		log.Fatal(err)
	}

	SyncActivities(db, user, activities)
	err = SyncUserActivities(db, client, user)
	if nil != err {
		log.Fatal(err)
	}
}

// Given activities from the API, insert them in the database
func SyncActivities(db *sqlx.DB, user User, activities []ApiActivity) (err error) {
	for _, apiActivity := range activities {
		_, err := db.Exec("INSERT INTO activities(id, name) VALUES($1, $2) ON CONFLICT(id) DO UPDATE SET name=excluded.name", apiActivity.Id, apiActivity.Name)
		if nil != err {
			log.Fatal(err)
		}
		_, err = db.Exec("INSERT INTO user_activity_counts(user_id, activity_id, count) VALUES($1, $2, $3) ON CONFLICT(user_id, activity_id) DO UPDATE SET count=excluded.count", user.Id, apiActivity.Id, apiActivity.Count)
		if nil != err {
			log.Fatal(err)
		}
	}
	return
}

// Given the activities we know the user has performed, fetch them from the API
// and insert them into the database
func SyncUserActivities(db *sqlx.DB, client http.Client, user User) (err error) {
	allUserActivityCounts := []UserActivityCount{}
	rows, err := db.Queryx("SELECT * FROM user_activity_counts WHERE user_id=$1", user.Id)
	for rows.Next() {
		userActivityCount := UserActivityCount{}
		err := rows.StructScan(&userActivityCount)
		if err != nil {
			return err
		}
		allUserActivityCounts = append(allUserActivityCounts, userActivityCount)
	}
	err = rows.Close()
	if err != nil {
		return err
	}

	//create a channel that can buffer everything if needed
	c := make(chan []ApiActivityHistory, len(allUserActivityCounts))

	go InsertActivityHistory(db, user, c)
	FetchUserActivities(client, allUserActivityCounts, c)

	//wait for the channel to finish
	<-c
	return
}

// Get detailed user activities from the API and send them to a channel
func FetchUserActivities(client http.Client, allUserActivityCounts []UserActivityCount, ch chan<- []ApiActivityHistory) {
	for _, userActivityCount := range allUserActivityCounts {
		log.Printf("Fetching activity history for activity %d\n", userActivityCount.ActivityId)
		err, apiActivityHistoryArray := getActivityHistory(client, userActivityCount.ActivityId)
		if err != nil {
			log.Fatal(err)
		}
		ch <- apiActivityHistoryArray
	}
	log.Println("Completed reading all activities from the Fitocracy API")
	close(ch)
}

// Given a channel of detailed user activities, insert them into the db
func InsertActivityHistory(db *sqlx.DB, user User, ch <-chan []ApiActivityHistory) {
	for {
		apiActivityHistoryArray := <-ch
		for _, activityHistory := range apiActivityHistoryArray {
			log.Printf("Looping over sets for %s\n", activityHistory.Name)
			for _, apiActivityAction := range activityHistory.Actions {
				performedAt, err := apiActivityAction.PerformedAt()
				if nil != err {
					log.Fatal(err)
				}
				log.Printf("Inserting user activity [%d] %s: %d on %s\n", apiActivityAction.Activity.Id, apiActivityAction.Activity.Name, apiActivityAction.Id, performedAt)
				_, err = db.Exec("INSERT OR IGNORE INTO user_activities(id, user_id, fitocracy_group_id, activity_id, units, reps, weight, performed_at) VALUES($1, $2, $3, $4, $5, $6, $7, $8)",
					apiActivityAction.Id, user.Id, activityHistory.GroupId, apiActivityAction.Activity.Id, apiActivityAction.Units(), apiActivityAction.Effort1, apiActivityAction.Effort0, performedAt)
				if nil != err {
					log.Fatal(err)
				}
			}
		}
	}
}