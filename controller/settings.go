package controller

import (
	"io/ioutil"
	"log"
	"strconv"

	"gopkg.in/yaml.v2"
)

type Settings struct {
	IP      string
	PORT    string
	UIIP    string
	UIPORT  string
	Indexes []int
	Debug   bool
}

func (s *Settings) LoadConfigFromFile(path string) {

	file, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(file), &s)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}

func (s *Settings) FormatIndexesForNetworkWriting() (indexes string) {
	for i, v := range s.Indexes {
		if i == (len(s.Indexes) - 1) {
			indexes = indexes + strconv.Itoa(v)
		} else {
			indexes = indexes + strconv.Itoa(v) + ","
		}
	}
	return
}
