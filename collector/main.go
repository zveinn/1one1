package collector

import (
	"log"
	"os"
	"os/signal"

	"github.com/zkynetio/lynx/collector/processor"
	"github.com/zkynetio/lynx/collector/stats"
	"github.com/zkynetio/lynx/helpers"
	"github.com/zkynetio/lynx/namespaces"
)

func Start(tag string, address string) {

	// Initialize a new controller
	collector := &processor.Collector{
		TAG: tag,
	}
	defer collector.CleanupOnExit()
	// the point map is only for debugging.
	collector.PointMap = make(map[int][]byte)
	collector.Controllers = make(map[string]*processor.Controller)

	processor.ConnectToControllers(
		address,
		tag,
		collector,
	)

	stats.InitStats()

	watcherChannel := make(chan int)
	go collector.MaintainControllerCommunications(watcherChannel)
	go collector.CollectStats(watcherChannel)
	namespaces.Init()
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
