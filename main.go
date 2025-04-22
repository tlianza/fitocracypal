package main

import (
	"encoding/csv"
	"flag"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

type Config struct {
	Exercises []Exercise `toml:"exercises"`
}

func main() {
	//load up our application config
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Fatalf("Fatal error config file: %s \n", err)
	}

	//load up our mapping file
	var config Config
	_, err = toml.DecodeFile(viper.GetString("exercise_mappings"), &config)
	if err != nil {
		log.Fatalf("Fatal error reading mapping file %s: %s \n", viper.GetString("exercise_mappings"), err)
	}

	exerciseMapper := NewExerciseMapper(config.Exercises)

	db, err := getDB()
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	username := flag.String("user", "", "Fitocracy Username")
	password := flag.String("pass", "", "Fitocracy Password")
	flag.Parse()

	if "" == *username {
		flag.PrintDefaults()
		log.Fatal("Required arguments not provided")
	}

	if "" == *password {
		log.Println("No password provided. Won't communicate with Fitocracy and will just dump CSVs based on local db")
	}

	//Fill the sqlite db with data from the API
	if "" != *password {
		PopulateDB(db, *username, *password)
	}

	err = DumpCSV(db, *username, viper.GetString("fitocracy_csv"), exerciseMapper, FitocracyCSVDumper{})
	if err != nil {
		log.Fatal("error generating csv: ", err)
	}
	err = DumpCSV(db, *username, viper.GetString("virtuagym_csv"), exerciseMapper, VirtuaGymCSVDumper{})
	if err != nil {
		log.Fatal("error generating csv: ", err)
	}
}

type CSVDumper interface {
	Dump(*csv.Writer, UserActivityDetail, *ExerciseMapper)
}

// Dump the contents of an already populated db into a csv
func DumpCSV(db *sqlx.DB, username string, filename string, exerciseMapper *ExerciseMapper, dumper CSVDumper) (err error) {

	err, user := GetUserByUsername(db, username)
	if nil != err {
		log.Fatal(err)
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0777)
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
