package main

import (
	"flag"
	"log"
	"os"

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

		controller.Start(*connect)
	case "brain":
		brain.Start()
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
