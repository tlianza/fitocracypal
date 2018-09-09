package main

import (
	"encoding/csv"
	"flag"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"strconv"
	"github.com/BurntSushi/toml"
)

type Config struct {
	Exercises []Exercise `toml:"exercises"`
}

func main() {
	//load up our mapping file
	var config Config
	_, err := toml.DecodeFile("exercise_mappings.toml", &config)
	exerciseMapper := NewExerciseMapper(config.Exercises)

	db, err := getDB()
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	username := flag.String("user", "", "Fitocracy Username")
	password := flag.String("pass", "", "Fitocracy Password")
	file := flag.String("file", "fitocracy.csv", "Path to exported csv")
	flag.Parse()

	if "" == *username || "" == *password {
		flag.PrintDefaults()
		log.Fatal("Required arguments not provided")
	}

	//TODO: split these into different operations that can be triggered with different flags

	//Fill the sqlite db with data from the API
	//PopulateDB(db, *username, *password)

	err = DumpCSV(db, *username, *file, exerciseMapper, FitocracyCSVDumper{})
	if err != nil {
		log.Fatal("error generating csv: ", err)
	}
	err = DumpCSV(db, *username, "virtuagym.csv", exerciseMapper, VirtuaGymCSVDumper{})
	if err != nil {
		log.Fatal("error generating csv: ", err)
	}
}

type CSVDumper interface {
	Dump(*csv.Writer, UserActivityDetail, *ExerciseMapper)
}

type FitocracyCSVDumper struct{}

// Dump the contents of an already populated db into a csv
func DumpCSV(db *sqlx.DB, username string, filename string, exerciseMapper *ExerciseMapper, dumper CSVDumper) (err error) {

	err, user := GetUserByUsername(db, username)
	if nil != err {
		log.Fatal(err)
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0777)
	defer file.Close()
	if err != nil {
		os.Exit(1)
	}

	rows, err := db.Queryx("SELECT * FROM user_activities JOIN activities ON user_activities.activity_id=activities.id WHERE user_id=$1", user.Id)
	if nil != err {
		log.Fatal(err)
	}

	csvWriter := csv.NewWriter(file)

	for rows.Next() {
		userActivityDetail := UserActivityDetail{}
		err := rows.StructScan(&userActivityDetail)
		if err != nil {
			return err
		}
		dumper.Dump(csvWriter, userActivityDetail, exerciseMapper)
	}

	// Write any buffered data to the underlying writer
	csvWriter.Flush()
	log.Printf("Wrote csv output to %s\n", filename)

	return
}

func (c FitocracyCSVDumper) Dump(csvWriter *csv.Writer, userActivityDetail UserActivityDetail, exerciseMapper *ExerciseMapper) {
	if err := csvWriter.Write([]string{userActivityDetail.PerformedAt.String(), strconv.Itoa(userActivityDetail.Activity.Id), userActivityDetail.Name, strconv.FormatFloat(userActivityDetail.Weight, 'f', -1, 32), strconv.FormatFloat(userActivityDetail.Reps, 'f', -1, 32)}); err != nil {
		log.Fatalln("error writing record to csv:", err)
	}
}
