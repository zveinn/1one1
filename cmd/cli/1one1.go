package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/zkynetio/lynx/alerting"
	"github.com/zkynetio/lynx/brain"
	"github.com/zkynetio/lynx/collector"
	"github.com/zkynetio/lynx/controller"
)

func main() {

	start := flag.String("start", "11", "The module you want to start")
	connect := flag.String("connectTo", "", "A network Address.. as IP:PORT or a domain")
	tag := flag.String("tag", "", "A tag")

	// debug := flag.String("debug", "", "Start this module specifically in debug mode")
	flag.Parse()

	switch *start {
	case "controller":
		if *connect == "" {
			log.Println("flag missing: [ -connectTo=X.X.X.X:XXXX ] this flag tells the controller where to connect.")
			os.Exit(1)
		}

		cfg := readControllerConfig()
		alerting := ReadAlertingConfig()
		controller.Start(*connect, &cfg, alerting)
	case "brain":
		newBrain := NewBrain()
		brain.Start(&newBrain)
	case "collector":

		if *connect == "" {
			log.Println("flag missing: [ -connectTo=X.X.X.X:XXXX ] this flag tells the controller where to connect.")
			os.Exit(1)
		}

		if *tag == "" {
			log.Println("tag missing: [ -tag=XXXXX ] this tag will tell the controller which collector is connecting")
			os.Exit(1)
		}

		collector.Start(*tag, *connect)
	default:
		if *start != "" {
			log.Println("Looks like the option [", *start, "] is not a valid one ..")
			log.Println("Here are the valid options: [ controller, collector, brain ]")
		} else {
			log.Println("You need to specify what you want to start: [ controller, collector, brain ]")
		}

	}

}

func ReadAlertingConfig() []alerting.Alerting {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	var alertingSlice []alerting.Alerting

	for _, f := range files {
		if strings.Contains(f.Name(), "alerts") {
			file, err := ioutil.ReadFile(f.Name())
			if err != nil {
				panic(err)
			}

			data := alerting.Alerting{}
			_ = json.Unmarshal([]byte(file), &data)
			alertingSlice = append(alertingSlice, data)
		}
	}
	return alertingSlice
}

func NewBrain() brain.Brain {
	file, err := ioutil.ReadFile("brain.json")
	if err != nil {
		panic(err)
	}
	data := brain.Config{}
	_ = json.Unmarshal([]byte(file), &data)
	// b.Config = data
	return brain.Brain{
		Config: data,
	}

}
func ReadCollectionConfig() {
	file, err := ioutil.ReadFile("collector.json")
	if err != nil {
		panic(err)
	}
	data := brain.Collecting{}
	_ = json.Unmarshal([]byte(file), &data)
	log.Println(data)
}
func readControllerConfig() controller.ControllerConfig {
	file, err := ioutil.ReadFile("controller.json")
	if err != nil {
		panic(err)
	}
	data := controller.ControllerConfig{}
	_ = json.Unmarshal([]byte(file), &data)
	return data
}
