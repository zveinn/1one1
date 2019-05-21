package main

import (
	"os"
	"os/signal"

	"github.com/zkynetio/lynx/collector/processor"
	"github.com/zkynetio/lynx/collector/stats"
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
	collector.PointMap = make(map[int][]byte)
	collector.StaticMap = make(map[int]string)

	stats.InitStats()

	processor.ConnectToControllers(
		os.Getenv("CONTROLLERS"),
		os.Getenv("TAG"),
		collector,
	)

	go collector.EngageControllerCommunications()
	go collector.MaintainControllerCommunications()
	// Each stats category should be it's own goroutine?

	// missing header means no change
	// header with a length of 1 means we're back to base.
	go collector.CollectStats()

	// todo
	// go collector.SendStats()

	if os.Getenv("DEBUG") == "true" {
		helpers.DebugLog(collector)
		helpers.DebugLog("Collector running...")
	}

	// capture stop signal in order to exit gracefully.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}
