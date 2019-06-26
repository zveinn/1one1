package main

import (
	"log"
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
	collector.Controllers = make(map[string]*processor.Controller)

	stats.InitStats()

	tag := os.Getenv("TAG")
	if os.Args[1] != "" {
		tag = os.Args[1]
	}
	processor.ConnectToControllers(
		os.Getenv("CONTROLLERS"),
		tag,
		collector,
	)

	watcherChannel := make(chan int)
	go collector.MaintainControllerCommunications(watcherChannel)
	go collector.CollectStats(watcherChannel)

	// todo
	// go collector.SendStats()

	if os.Getenv("DEBUG") == "true" {
		helpers.DebugLog(collector)
		helpers.DebugLog("Collector running...")
	}

	// capture stop signal in order to exit gracefully.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	for {
		select {
		case index := <-watcherChannel:
			if index == 1 {
				go collector.CollectStats(watcherChannel)
			} else if index == 2 {
				go collector.MaintainControllerCommunications(watcherChannel)
			}
			log.Println("goroutine number", index, "just restarted...")
			break
		case <-stop:
			// TODO: handle exit gracefully
			log.Println("handle exit gracefully")
			os.Exit(1)
		}
	}
}
