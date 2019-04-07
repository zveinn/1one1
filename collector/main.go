package main

import (
	"os"
	"os/signal"

	"github.com/zkynetio/lynx/collector/processor"
	"github.com/zkynetio/lynx/helpers"
)

func main() {

	helpers.LoadEnvironmentVariables()

	// Initialize a new controller
	collector := &processor.Collector{
		TAG:          os.Getenv("TAG"),
		RecoveryFile: os.Getenv("RCOVERYFILE"),
	}
	collector.GetIntervalsFromEnvironmentVariables()
	defer collector.CleanupOnExit()

	processor.ConnectToControllers(
		os.Getenv("CONTROLLERS"),
		os.Getenv("TAG"),
		collector,
	)

	go collector.EngageControllerListeners()
	go collector.MaintainControllerConnections()
	go collector.CollectStats()

	if os.Getenv("DEBUG") == "true" {
		helpers.DebugLog(collector)
		helpers.DebugLog("Collector running...")
	}

	// capture stop signal in order to exit gracefully.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}
