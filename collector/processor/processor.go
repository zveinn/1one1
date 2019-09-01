package processor

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"log"
	"math/rand"
	"net"
	"runtime/debug"
	"sync"
	"time"

	stats "github.com/zkynetio/lynx/collector/stats"
	helpers "github.com/zkynetio/lynx/helpers"
	"github.com/zkynetio/lynx/namespaces"
)

type Collector struct {
	TAG         string
	Controllers map[string]*Controller
	mutex       sync.Mutex
	Namespaces  map[int]string
	mux         sync.Mutex
}
type Controller struct {
	Address     string
	Active      bool
	Conn        net.Conn
	Retry       int
	mutex       sync.Mutex
	SendChannel chan DataPoint
	//InactiveSince time.Time
}
type DataPoint struct {
	Data        []byte
	ControlByte byte
	Timestamp   int64
	Length      int
}

func (c *Collector) AddController(cont *Controller) {
	c.mutex.Lock()
	c.Controllers[cont.Address] = cont
	// c.Controllers = append(c.Controllers, cont)
	c.mutex.Unlock()
}
func (c *Collector) RemoveController(cont *Controller) {
	c.mutex.Lock()
	if cont.Conn != nil {
		_ = cont.Conn.Close()
	}
	delete(c.Controllers, cont.Address)
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

func (collector *Collector) MaintainControllerCommunications(watcherChannel chan int) {
	defer func(watcherChannel chan int) {
		if r := recover(); r != nil {
			log.Println("collector panic...", r)
		}
		watcherChannel <- 2
		log.Println("Stopped maintining controllers...")
	}(watcherChannel)
	for {
		rand.Seed(time.Now().UnixNano())
		n := rand.Intn(3000)
		time.Sleep(time.Duration(n+1000) * time.Millisecond)
		for _, controller := range collector.Controllers {
			if controller.Active {
				continue
			}
			log.Println("maintaining controller:", controller)
			if err := collector.dialAndHandshake(controller, collector.TAG); err != nil {
				log.Println("CONTROLLER COM. ERROR:", controller.Address)
				continue
			}

			log.Println("Engaging controller listener to", controller.Address)
			go controller.OpenSendChannel()
			controller.ChangeActiveStatus(true)
		}
	}

}

func (c *Controller) StartListening() {
	for {
		controlBytes := make([]byte, 3)
		_, err := c.Conn.Read(controlBytes)
		if err != nil {
			return
		}
		data := make([]byte, binary.LittleEndian.Uint16(controlBytes[:2]))
		_, err = c.Conn.Read(data)
		if err != nil {
			return
		}

		if controlBytes[2] == 0 {
			log.Println("Controller send an error", err)
			continue
		}
		if controlBytes[2] == 1 {
			// address change..

		}

	}
}
func (collector *Collector) CollectStats(watcherChannel chan int) {
	defer func(watcherChannel chan int) {
		if r := recover(); r != nil {
			log.Println(string(debug.Stack()))
			log.Println("collector panic...", r)
		}
		watcherChannel <- 1
	}(watcherChannel)
	startTime := time.Now()
	for {
		var data []byte
		if !time.Now().After(startTime.Add(100 * time.Millisecond)) {
			time.Sleep(20 * time.Millisecond)
			continue
		}
		startTime = time.Now()
		go func() {
			defer func() {
				r := recover()
				if r != nil {
					log.Println("panic while sending stats:", r, string(debug.Stack()))
				}
			}()
			data = stats.GetMinimumStats(collector.Namespaces)
			ControlByte := byte(4)
			for _, controller := range collector.Controllers {
				if controller.Active {
					controller.SendChannel <- DataPoint{
						Data:        data,
						ControlByte: ControlByte,
						Timestamp:   time.Now().UnixNano(),
						Length:      len(data),
					}
				}
			}
		}()

	}
}

func (c *Controller) ChangeActiveStatus(status bool) {
	c.mutex.Lock()
	c.Active = status
	c.mutex.Unlock()
}

func (c *Controller) Setconnection(conn net.Conn) {
	c.mutex.Lock()
	c.Conn = conn
	c.mutex.Unlock()
}

func (c *Controller) OpenSendChannel() {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered inside OpenSendChannel()")
		}
		log.Println("Closing send loop to controller", c.Address)
		c.ChangeActiveStatus(false)
		close(c.SendChannel)
	}()
	// 180 datapoints is 3 minutes of downtime
	c.SendChannel = make(chan DataPoint, 180)
	newbuffer := new(bytes.Buffer)
	for {
		// incase we are recovering from being offline for 3 minutes
		// we want to sleep 10 milliseconds before trying to send the next one.
		time.Sleep(10 * time.Millisecond)
		datapoint, errx := <-c.SendChannel
		if !errx {
			break
		}
		err := binary.Write(newbuffer, binary.LittleEndian, int16(datapoint.Length))
		if err != nil {
			panic(err)
		}
		newbuffer.Write([]byte{datapoint.ControlByte})
		err = binary.Write(newbuffer, binary.LittleEndian, int64(datapoint.Timestamp))
		if err != nil {
			panic(err)
		}
		newbuffer.Write(datapoint.Data)
		n, err := newbuffer.WriteTo(c.Conn)
		newbuffer.Reset()
		if err != nil {
			log.Println("ERROR WHEN WRITING STATS (Count", n, ") err:", err)
			break
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
func (c *Collector) handShakeWithController(controller *Controller, tag string) (err error) {
	_, err = controller.Conn.Write([]byte(tag + "\n"))
	helpers.PanicX(err)
	var ns []string

	msg, err := bufio.NewReader(controller.Conn).ReadString('\n')
	log.Println(string(msg))
	if err != nil {
		log.Println("Coult not handshake with controller", msg)
		controller.Conn.Close()
	}
	err = json.Unmarshal([]byte(msg), &ns)
	if err != nil {
		log.Println("Coult not handshake with controller", msg)
		controller.Conn.Close()
	}
	log.Println("GOT THESE NAMESPACES FROM THE CONTROLLER:", ns)
	c.Namespaces = namespaces.MakeMapFromNamespaces(ns)
	log.Println("NAMESPACE MAP:", c.Namespaces)

	return
}
func (c *Collector) dialAndHandshake(controller *Controller, tag string) (err error) {
	err = dialController(controller)
	if err != nil {
		return
	}

	err = c.handShakeWithController(controller, tag)
	if err != nil {
		return
	}
	controller.ChangeActiveStatus(true)
	return
}
func ConnectToControllers(address string, tag string, collector *Collector) {
	controller := &Controller{Address: address, Active: false}
	collector.AddController(controller)
	log.Println("Connecting to:", controller)
	if err := collector.dialAndHandshake(controller, tag); err != nil {
		log.Println("CONTROLLER COM. ERROR:", controller.Address)
		return
	}
	log.Println("Connected to:", controller.Address)
	go controller.OpenSendChannel()
}
