package main

import (
	"encoding/json"
	"github.com/mpetavy/common"
	"io/ioutil"
	"os/user"
)

type Ranking struct {
	Scores map[string]int
}

func LoadlRanking() *Ranking {
	ranking := &Ranking{}
	ranking.Scores = make(map[string]int)

	if common.FileExists(filename) {
		ba, err := ioutil.ReadFile(filename)
		if !common.Error(err) {
			err = json.Unmarshal(ba, ranking)
			if !common.Error(err) {
				return ranking
			}
		}
	}

	return ranking
}

func (r *Ranking) Score(points int) (error, bool) {
	usr, err := user.Current()
	if common.Error(err) {
		return err, false
	}

	curPoints, ok := r.Scores[usr.Name]
	if ok {
		if curPoints > points {
			points = 0
		}
	}

	if points == 0 {
		return nil, false
	}

	highscore := true

	for _, v := range r.Scores {
		if v > points {
			highscore = false

			break
		}
	}

	r.Scores[usr.Name] = points

	err = r.Save()
	if common.Error(err) {
		return err, false
	}

	return nil, highscore
}

func (r *Ranking) Save() error {
	ba, err := json.Marshal(r)
	if common.Error(err) {
		return err
	}

	err = ioutil.WriteFile(filename, ba, common.DefaultFileMode)
	if common.Error(err) {
		return err
	}

	return nil
}
