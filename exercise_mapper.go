package main

import "github.com/cloudflare/cfssl/log"

type Exercise struct {
	FitocracyName string `toml:"fitocracy_name"`
	FitocracyId   int    `toml:"fitocracy_id"`
	MFPName       string `toml:"mfp_name"`
	MFPId         int    `toml:"mfp_id"`
	VirtuaGymName string `toml:"virtuagym_name"`
	VirtuaGymId   int    `toml:"virtuagym_id"`
}

type ExerciseMapper struct {
	exercises []Exercise
	ByFitocracyId map[int]Exercise
	ByVirtuaGymId map[int]Exercise
}

/**
	Instantiate a new exercise mapper, with lookup tables pre-initialized
 */
func NewExerciseMapper(exercises []Exercise) *ExerciseMapper {
	mapper := new(ExerciseMapper)
	mapper.exercises = exercises
	mapper.ByFitocracyId =  make(map[int]Exercise)
	mapper.ByVirtuaGymId =  make(map[int]Exercise)

	for _, e := range exercises {
		mapper.ByFitocracyId[e.FitocracyId] = e
		if e.VirtuaGymId > 0 {
			mapper.ByVirtuaGymId[e.VirtuaGymId] = e
		} else {
			log.Warningf("No virtuagym mapping for %d - %s", e.FitocracyId, e)
		}
	}

	return mapper
}