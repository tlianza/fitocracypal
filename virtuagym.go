package main

import (
	"encoding/csv"
	"strconv"
	"log"
)

type VirtuaGymCSVDumper struct { }

func (c VirtuaGymCSVDumper) Dump (csvWriter *csv.Writer, userActivityDetail UserActivityDetail, exerciseMapper *ExerciseMapper) {
	e := exerciseMapper.ByFitocracyId[userActivityDetail.Activity.Id]

	if e.VirtuaGymId <= 0 {
		//log.Printf("Unknown virtuagym exercise: %d, %s\n", userActivityDetail.Activity.Id, userActivityDetail.Name)
		return //can't add to csv
	}

	if err := csvWriter.Write([]string{userActivityDetail.PerformedAt.String(), strconv.Itoa(e.VirtuaGymId), userActivityDetail.Name, strconv.FormatFloat(userActivityDetail.Weight,'f', -1, 32), strconv.FormatFloat(userActivityDetail.Reps,'f', -1, 32)}); err != nil {
		log.Fatalln("error writing record to csv:", err)
	}
}