package main

import (
	"github.com/jmoiron/sqlx"
	"time"
)

type User struct {
	Id                int       `db:"id"`
	FitocracyId       int       `db:"fitocracy_id"`
	FitocracyUsername string    `db:"fitocracy_username"`
	CreatedAt         time.Time `db:"created_at"`
}

type Activity struct {
	Id          int       `db:"id"`
	FitocracyId int       `db:"fitocracy_id"`
	Name        string    `db:"name"`
	CreatedAt   time.Time `db:"created_at"`
}

type UserActivityCount struct {
	UserId     int       `db:"user_id"`
	ActivityId int       `db:"activity_id"`
	Count      int       `db:"count"`
	CreatedAt  time.Time `db:"created_at"`
}

type UserActivity struct {
	Id               int       `db:"id"`
	UserId           int       `db:"user_id"`
	ActivityId       int       `db:"activity_id"`
	FitocracyGroupId int       `db:"fitocracy_group_id"`
	Units            string    `db:"units"`
	Reps             float64   `db:"reps"`
	Weight           float64   `db:"weight"`
	PerformedAt      time.Time `db:"performed_at"`
	CreatedAt        time.Time `db:"created_at"`
}

// Log API events we perform that actually mutate state, so
// we have some facility for tracking/undoing them
type ApiActivityLog struct {
	Id        int       `db:"id"`
	Operation string    `db:"operation"`
	ResultId  string    `db:"result_id"`
	Status    int       `db:"status"`
	Note      string    `db:"note"`
	CreatedAt time.Time `db:"created_at"`
}

// Used for generating CSVs when you need to join these two tables together
type UserActivityDetail struct {
	*UserActivity
	*Activity
}

var schema = `
CREATE TABLE IF NOT EXISTS users (
    id                 INTEGER PRIMARY KEY,
    fitocracy_id       INTEGER UNIQUE NOT NULL,	
    fitocracy_username TEXT UNIQUE NOT NULL,
	created_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS activities (
    id       	INTEGER PRIMARY KEY,
	name        TEXT NOT NULL,
	created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_activity_counts (
    user_id			   INT REFERENCES users(id) ON DELETE CASCADE,
    activity_id        INT REFERENCES activities(id) ON DELETE CASCADE,
    count       	   INTEGER NOT NULL DEFAULT 0,
	created_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY(user_id, activity_id)
);

CREATE TABLE IF NOT EXISTS user_activities (
    id       		   	   INTEGER PRIMARY KEY,
 	fitocracy_group_id     INTEGER,
    user_id			       INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_id  		   INT NOT NULL REFERENCES activities(id) ON DELETE CASCADE,
    units        	       TEXT,
    reps        	       DECIMAL(6, 1),
	weight       	       DECIMAL(6, 1),
	performed_at  	       TIMESTAMP NOT NULL,
	created_at             TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS api_activity_log (
    id                 INTEGER PRIMARY KEY,
    operation          TEXT NOT NULL,	
    status             INTEGER,
    result_id 		   TEXT NULL,
	note 			   TEXT NULL,
	created_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);`

func getDB() (db *sqlx.DB, err error) {
	return sqlx.Connect("sqlite3", "./fitocracy.db")
}

func ensureSchema(db *sqlx.DB) {
	db.MustExec(schema)
}

func GetUserByFitocracyId(db *sqlx.DB, fitocracyUserId int) (err error, user User) {
	err = db.Get(&user, "SELECT * FROM users WHERE fitocracy_id=$1", fitocracyUserId)
	return
}

func GetUserByUsername(db *sqlx.DB, fitocracyUsername string) (err error, user User) {
	err = db.Get(&user, "SELECT * FROM users WHERE fitocracy_username=$1", fitocracyUsername)
	return
}
