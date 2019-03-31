package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/shirou/gopsutil/mem"
)

// NOTES
// check for duplicate port/ip on collectors

func debugLog(v ...interface{}) {
	if os.Getenv("DEBUG") == "true" {
		log.Println(v...)
	}
}
func panicX(err error) {
	if err != nil {
		panic(err)
	}
}

type Collector struct {
	Buffer             chan string
	RecoveryFile       string
	TAG                string
	Controllers        []*Controller
	mutex              sync.Mutex
	MaintainerInterval int
	ListenerInterval   int
}

func (c *Collector) GetIntervalsFromEnvironmentVariables() {
	// set listener interval
	listenerInterval, err := strconv.Atoi(os.Getenv("LISTENER_INTERVAL"))
	if err != nil {
		debugLog("Error setting listener interval, default valur selected", err)
		listenerInterval = 5
	}
	c.ListenerInterval = listenerInterval

	// set maintainer interval
	maintainerInterval, err := strconv.Atoi(os.Getenv("MAINTAINER_INTERVAL"))
	if err != nil {
		debugLog("Error setting maintainer interval, default value selected", err)
		maintainerInterval = 5
	}
	c.MaintainerInterval = maintainerInterval
}
func (c *Collector) AddController(cont *Controller) {
	c.mutex.Lock()
	c.Controllers = append(c.Controllers, cont)
	c.mutex.Unlock()
}
func (c *Collector) CleanupOnExit() {
	for _, controller := range c.Controllers {
		log.Println("Closing:", controller.Address)
		if controller.Conn != nil {
			_ = controller.Conn.Close()
		}
	}
}

type Controller struct {
	Address     string
	Active      bool
	HasListener bool
	Conn        net.Conn
	Retry       int
	mutex       sync.Mutex
	//InactiveSince time.Time
}

func (c *Controller) ChangeActiveStatus(status bool) {
	c.mutex.Lock()
	c.Active = status
	c.mutex.Unlock()
}
func (c *Controller) ChangeListenerStatus(status bool) {
	c.mutex.Lock()
	c.HasListener = status
	c.mutex.Unlock()
}
func (c *Controller) Setconnection(conn net.Conn) {
	c.mutex.Lock()
	c.Conn = conn
	c.mutex.Unlock()
}

func loadEnvironmentVariables() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(err)
		log.Fatal("Error loading .env file")
	}
}

func listenToController(controller *Controller) {
	defer func() {
		debugLog("defering read pipe from", controller.Address)
		controller.ChangeActiveStatus(false)
		controller.ChangeListenerStatus(false)
		controller.Setconnection(nil)
	}()

	for {
		msg, err := bufio.NewReader(controller.Conn).ReadString('\n')
		// TODO: handle better
		if err != nil || msg == "c\n" {
			_ = controller.Conn.Close()
			debugLog("Error and/or message in read pipe from" + controller.Address + " // " + msg + " //" + err.Error())
			break
		}

		go handleMessageFromController(msg)
	}

}

func handleMessageFromController(msg string) {
	debugLog("IN:", msg)
}

func engageControllerListeners(collector *Collector) {
	for {
		time.Sleep(time.Duration(collector.ListenerInterval) * time.Second)
		for _, controller := range collector.Controllers {
			if !controller.Active || controller.HasListener {
				continue
			}
			debugLog("Engaging controller listener to", controller.Address)
			go listenToController(controller)
			controller.ChangeListenerStatus(true)
		}
	}

}
func maintainControllerConnections(collector *Collector) {
	for {
		// TODO: implement rand int sleeper
		time.Sleep(time.Duration(collector.MaintainerInterval) * time.Second)
		debugLog("5 second controller maintnance starting ...")
		for _, controller := range collector.Controllers {
			if controller.Active {
				continue
			}
			if err := dialAndHandshake(controller, collector.TAG); err != nil {
				debugLog("CONTROLLER COM. ERROR:", controller.Address)
				continue
			}
			debugLog("Recovered connection to:", controller.Address)
			controller.ChangeActiveStatus(true)
		}
	}

}

func dialController(controller *Controller) (err error) {
	conn, err := net.Dial("tcp", controller.Address)
	if err != nil {
		return
	}
	controller.Setconnection(conn)
	return
}
func handShakeWithController(controller *Controller, tag string) (err error) {
	controller.Conn.Write([]byte(tag + "\n"))
	message, err := bufio.NewReader(controller.Conn).ReadString('\n')
	if err != nil || message != "k\n" {
		_ = controller.Conn.Close()
		controller.Setconnection(nil)
		// TODO: handle better
		err = errors.New("messsage from controller was:" + message + " // pipe read error was" + err.Error())
	}

	return
}
func dialAndHandshake(controller *Controller, tag string) (err error) {
	if err = dialController(controller); err != nil {
		return
	}
	if err = handShakeWithController(controller, tag); err != nil {
		return
	}
	return
}
func connectToControllers(controllers string, tag string, collector *Collector) {
	for _, v := range strings.Split(controllers, ",") {
		controller := &Controller{Address: v, Active: false, HasListener: false}
		collector.AddController(controller)

		debugLog("Connecting to:", v)
		if err := dialAndHandshake(controller, tag); err != nil {
			debugLog("CONTROLLER COM. ERROR:", controller.Address)
			continue
		}
		debugLog("Connected to:", controller.Address)
		controller.ChangeActiveStatus(true)
	}
}

func main() {

	loadEnvironmentVariables()

	// Initialize a new controller
	collector := &Collector{
		TAG:          os.Getenv("TAG"),
		RecoveryFile: os.Getenv("RCOVERYFILE"),
	}
	collector.GetIntervalsFromEnvironmentVariables()
	defer collector.CleanupOnExit()

	connectToControllers(
		os.Getenv("CONTROLLERS"),
		os.Getenv("TAG"),
		collector,
	)

	go engageControllerListeners(collector)
	go maintainControllerConnections(collector)
	go collectStats(collector)

	if os.Getenv("DEBUG") == "true" {
		debugLog(collector)
		debugLog("Collector running...")
	}

	// capture stop signal in order to exit gracefully.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}

func collectStats(collector *Collector) {
	defer func() {
		if r := recover(); r != nil {
			debugLog("recovered in stats collecting rutine")
			return
		}
	}()
	for {
		time.Sleep(1 * time.Millisecond)
		vmStat, err := mem.VirtualMemory()
		panicX(err)

		for _, controller := range collector.Controllers {
			if controller.Active {
				_, err := controller.Conn.Write([]byte(vmStat.String() + "\n"))
				//debugLog("wrote", data)
				if err != nil {
					debugLog("ERROR WHEN WRITING STATS:", err)
					return
				}
			}
		}
	}

}
