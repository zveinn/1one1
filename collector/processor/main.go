package processor

import (
	"bufio"
	"errors"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	stats "github.com/zkynetio/lynx/collector/stats"
	helpers "github.com/zkynetio/lynx/helpers"
)

var NS []string

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
		helpers.DebugLog("Error setting listener interval, default valur selected", err)
		listenerInterval = 5
	}
	c.ListenerInterval = listenerInterval

	// set maintainer interval
	maintainerInterval, err := strconv.Atoi(os.Getenv("MAINTAINER_INTERVAL"))
	if err != nil {
		helpers.DebugLog("Error setting maintainer interval, default value selected", err)
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

func (collector *Collector) EngageControllerListeners() {
	for {
		time.Sleep(time.Duration(collector.ListenerInterval) * time.Second)
		for _, controller := range collector.Controllers {
			if !controller.Active || controller.HasListener {
				continue
			}
			helpers.DebugLog("Engaging controller listener to", controller.Address)
			go controller.Listen()
			go controller.OpenSendChannel()
			controller.ChangeListenerStatus(true)
			_, err := controller.Conn.Write([]byte("k\n"))
			helpers.PanicX(err)
		}
	}

}
func (collector *Collector) MaintainControllerConnections() {
	for {
		// TODO: implement rand int sleeper
		time.Sleep(time.Duration(collector.MaintainerInterval) * time.Second)
		helpers.DebugLog("5 second controller maintnance starting ...")
		for _, controller := range collector.Controllers {
			if controller.Active {
				continue
			}
			if err := dialAndHandshake(controller, collector.TAG); err != nil {
				helpers.DebugLog("CONTROLLER COM. ERROR:", controller.Address)
				continue
			}
			helpers.DebugLog("Recovered connection to:", controller.Address)
			controller.ChangeActiveStatus(true)
		}
	}

}

func (collector *Collector) CollectStats() {
	defer func() {
		if r := recover(); r != nil {
			helpers.DebugLog("recovered in stats collecting rutine", r)
			return
		}
	}()
	for {

		//fetchMemory()
		//fetchDISK()
		//fetchCPU()
		//fetchCPUPercentage()
		//getHost()
		//fetchNetworkIFS()
		//log.Println(getUploadDownload("wlp2s0"))
		//getEntropy()
		//getUsers()
		//getProcesses()
		//getLoad()
		//getActiveSessions()

		time.Sleep(10000 * time.Millisecond)
		// TODO: make randomizer !
		for _, controller := range collector.Controllers {
			if controller.Active && controller.NSDelivered {
				controller.Send <- "boop\n"
			}
		}
	}

}

func (c *Controller) OpenSendChannel() {
	defer func() {
		log.Println("Closing send loop to controller", c.Address)
	}()
	c.Send = make(chan string, 10000)
	for {
		data, errx := <-c.Send
		if !errx {
			break
		}
		helpers.DebugLog("sending to controller:", data)
		_, err := c.Conn.Write([]byte(data))
		if err != nil {
			helpers.DebugLog("ERROR WHEN WRITING STATS:", err)
			close(c.Send)
			break
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
	NSDelivered bool
	Send        chan string
	//InactiveSince time.Time
}

func (c *Controller) ChangeActiveStatus(status bool) {
	c.mutex.Lock()
	c.Active = status
	c.mutex.Unlock()
}
func (c *Controller) HaveNamespacesBeenDelivered(delivered bool) {
	c.mutex.Lock()
	c.NSDelivered = delivered
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

func (c *Controller) Listen() {
	defer func() {

		helpers.DebugLog("defering read pipe from", c.Address)
		close(c.Send)
		c.ChangeActiveStatus(false)
		c.ChangeListenerStatus(false)
		c.Setconnection(nil)
		c.HaveNamespacesBeenDelivered(false)
	}()

	for {
		msg, err := bufio.NewReader(c.Conn).ReadString('\n')
		// TODO: handle better
		if err != nil || msg == "c\n" {
			_ = c.Conn.Close()
			helpers.DebugLog("Error and/or message in read pipe from" + c.Address + " // " + msg + " //" + err.Error())
			break
		}

		if strings.Contains(msg, "ns:") {
			namespaces := strings.Split(strings.Split(strings.TrimSuffix(msg, "\n"), ":")[1], ",")
			log.Println("NAMESPACES:", namespaces)
			NS = namespaces
			// TODO: deliver base stats
			_, err := c.Conn.Write([]byte(stats.GetHost() + "\n"))
			helpers.PanicX(err)
			c.HaveNamespacesBeenDelivered(true)
		}
		helpers.DebugLog("IN:", msg)
		helpers.DebugLog("NS:", NS)
		//go handleMessageFromController(msg)
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
	_, err = controller.Conn.Write([]byte(tag + "\n"))
	helpers.PanicX(err)
	message, err := bufio.NewReader(controller.Conn).ReadString('\n')
	log.Println(string(message))
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
func ConnectToControllers(controllers string, tag string, collector *Collector) {
	for _, v := range strings.Split(controllers, ",") {
		controller := &Controller{Address: v, Active: false, HasListener: false, NSDelivered: false}
		collector.AddController(controller)

		helpers.DebugLog("Connecting to:", v)
		if err := dialAndHandshake(controller, tag); err != nil {
			helpers.DebugLog("CONTROLLER COM. ERROR:", controller.Address)
			continue
		}
		helpers.DebugLog("Connected to:", controller.Address)
		controller.ChangeActiveStatus(true)
	}
}
