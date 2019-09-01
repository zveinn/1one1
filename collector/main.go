package collector

import (
	"log"
	"os"
	"os/signal"

	"github.com/zkynetio/lynx/collector/processor"
	"github.com/zkynetio/lynx/collector/stats"
	"github.com/zkynetio/lynx/namespaces"
)

func Start(tag string, address string) {

	// Initialize a new controller
	collector := &processor.Collector{
		TAG: tag,
	}
	defer collector.CleanupOnExit()
	// the point map is only for debugging.
	collector.Controllers = make(map[string]*processor.Controller)
	namespaces.Init()

	processor.ConnectToControllers(
		address,
		tag,
		collector,
	)

	stats.InitStats()

	watcherChannel := make(chan int)
	go collector.MaintainControllerCommunications(watcherChannel)
	go collector.CollectStats(watcherChannel)

	if os.Getenv("DEBUG") == "true" {
		log.Println(collector)
		log.Println("Collector running...")
	}

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
			log.Println("handle exit gracefully")
			collector.CleanupOnExit()

			os.Exit(1)
		}
	}
}
