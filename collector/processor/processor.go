package processor

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/binary"
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

type Collector struct {
	Buffer             chan []byte
	RecoveryFile       string
	TAG                string
	Controllers        map[string]*Controller
	mutex              sync.Mutex
	PointMap           map[int][]byte
	StaticMap          map[int]string
	LastBasePointIndex int
	CurrentPointIndex  int
	CurrentStaticIndex int
	MaintainerInterval int
	ListenerInterval   int
	CollectionInterval int
	mux                sync.Mutex
}

// 60 * 60 = 3600 data points = 1 hour
// x 48 = 172.800 = 2 days ( 48 hours )
// asuming each data point is 200 bytes
// we need 34.560.000 or 34,56MB of memory to store
// 48 hours worth of stats.
func (c *Collector) AddDataPoint(point []byte) (count int) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if len(c.PointMap) > 200000 {
		c.PointMap = nil
		c.PointMap = make(map[int][]byte)
		c.CurrentPointIndex = 0
	}
	c.PointMap[c.CurrentPointIndex] = point
	c.CurrentPointIndex++
	return c.CurrentPointIndex
}
func (c *Collector) AddStaticPoint(point string) (count int) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if len(c.StaticMap) > 200000 {
		c.StaticMap = nil
		c.StaticMap = make(map[int]string)
		c.CurrentStaticIndex = 0
	}
	c.StaticMap[c.CurrentStaticIndex] = point
	c.CurrentStaticIndex++
	return c.CurrentStaticIndex
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

	CollectionInterval, err := strconv.Atoi(os.Getenv("COLLECTION_INTERVAL"))
	if err != nil {
		helpers.DebugLog("Error setting collection interval, default value selected", err)
		maintainerInterval = 5
	}
	c.CollectionInterval = CollectionInterval
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
		helpers.DebugLog("Closing:", controller.Address)
		if controller.Conn != nil {
			_ = controller.Conn.Close()
		}
	}
}

func sendFirstBasePoint(collector *Collector, controller *Controller) {
	var data []byte
	if len(collector.PointMap) > 0 {
		data = collector.PointMap[collector.LastBasePointIndex]
	} else {
		data = stats.CollectBasePoint()
		_ = collector.AddDataPoint(data)
	}
	newbuffer := new(bytes.Buffer)
	err := binary.Write(newbuffer, binary.LittleEndian, int16(len(data)))
	if err != nil {
		panic(err)
	}
	newbuffer.Write([]byte{101})
	timestamp := time.Now().UnixNano()
	err = binary.Write(newbuffer, binary.LittleEndian, int64(timestamp))
	if err != nil {
		panic(err)
	}
	// helpers.DebugLog("buffer len after timestamp", newbuffer.Len())
	// helpers.DebugLog("last base index:", collector.LastBasePointIndex)
	// helpers.DebugLog(collector.PointMap[collector.LastBasePointIndex])
	newbuffer.Write(data)
	// newbuffer = append(newbuffer, data)
	// helpers.DebugLog("bytes from first base:", newbuffer.Bytes())
	_, err = newbuffer.WriteTo(controller.Conn)
	if err != nil {
		panic(err)
	}
	newbuffer.Reset()
}

func (collector *Collector) MaintainControllerCommunications(watcherChannel chan int) {
	defer func(watcherChannel chan int) {
		if r := recover(); r != nil {
			log.Println("collector panic...", r)
		}
		watcherChannel <- 2
	}(watcherChannel)
	for {
		// TODO: implement rand int sleeper
		time.Sleep(time.Duration(collector.MaintainerInterval) * time.Second)
		//helpers.DebugLog("5 second controller maintnance starting ...")
		for _, controller := range collector.Controllers {
			log.Println("maintaining controller:", controller)
			if controller.Active {
				continue
			}
			if err := collector.dialAndHandshake(controller, collector.TAG); err != nil {
				//helpers.DebugLog("CONTROLLER COM. ERROR:", controller.Address)
				continue
			}
			helpers.DebugLog("Recovered connection to:", controller.Address)
			controller.ChangeActiveStatus(true)

			helpers.DebugLog("Engaging controller listener to", controller.Address)
			go controller.OpenSendChannel()
			sendFirstBasePoint(collector, controller)
		}
	}

}

func (collector *Collector) CollectStats(watcherChannel chan int) {
	defer func(watcherChannel chan int) {
		if r := recover(); r != nil {
			log.Println("collector panic...", r)
		}
		watcherChannel <- 1
	}(watcherChannel)

	count := collector.CurrentPointIndex
	for {
		var data []byte
		time.Sleep(time.Duration(collector.CollectionInterval) * time.Millisecond)
		if count%60 == 0 {
			staticData := stats.CheckStaticDataForChanges()
			if staticData != "" {
				log.Println("do something with static data !")
			}
			data = stats.CollectBasePoint()
			collector.LastBasePointIndex = count
			count = collector.AddDataPoint(data)
		} else {
			data = stats.CollectDynamicData()
			count = collector.AddDataPoint(data)
		}

		accumilatedBytes := 0
		pointCount := 0
		var allBytes []byte
		for _, v := range collector.PointMap {
			accumilatedBytes = accumilatedBytes + len(v)
			allBytes = append(allBytes, v...)
			pointCount++
		}
		var b bytes.Buffer
		w := zlib.NewWriter(&b)
		w.Write(allBytes)
		w.Close()
		helpers.DebugLog("Current DP size:", len(data), "Accumilated DP size:", accumilatedBytes, " Compressed:", b.Len(), "DP count:", pointCount)
		// continue
		// log.Println(data)
		for _, controller := range collector.Controllers {
			// log.Println("controller", controller.Address, " is active:", controller.Active)
			if controller.Active {
				controller.Send <- data
			}
		}
	}
}

type Controller struct {
	Address string
	Active  bool
	Conn    net.Conn
	Retry   int
	mutex   sync.Mutex
	Send    chan []byte
	//InactiveSince time.Time
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
		helpers.DebugLog("Closing send loop to controller", c.Address)
		c.ChangeActiveStatus(false)
		close(c.Send)
	}()
	c.Send = make(chan []byte, 10000)
	newbuffer := new(bytes.Buffer)
	for {
		data, errx := <-c.Send
		if !errx {
			break
		}

		err := binary.Write(newbuffer, binary.LittleEndian, int16(len(data)))
		if err != nil {
			panic(err)
		}
		newbuffer.Write([]byte{102})
		timestamp := time.Now().UnixNano()
		err = binary.Write(newbuffer, binary.LittleEndian, int64(timestamp))
		if err != nil {
			panic(err)
		}

		newbuffer.Write(data)

		// newbuffer = append(newbuffer, data)
		// helpers.DebugLog("Bytes from pipes:", newbuffer.Bytes())
		n, err := newbuffer.WriteTo(c.Conn)
		newbuffer.Reset()
		// _, err := c.Conn.Write(newbuffer)
		if err != nil {
			helpers.DebugLog("ERROR WHEN WRITING STATS (Count", n, ") err:", err)
			close(c.Send)
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

	data := stats.GetStaticBasePoint()
	_ = c.AddStaticPoint(data)
	_, err = controller.Conn.Write([]byte(data + "\n"))
	helpers.PanicX(err)

	msg, err := bufio.NewReader(controller.Conn).ReadString('\n')
	if msg != "k\n" || err != nil {
		log.Println("Coult not handshake with controller", msg)
		if err != nil {
			return err
		}

		return errors.New("Could not handshake with controller")
	}
	controller.ChangeActiveStatus(true)
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
	return
}
func ConnectToControllers(controllers string, tag string, collector *Collector) {
	for _, v := range strings.Split(controllers, ",") {
		controller := &Controller{Address: v, Active: false}
		collector.AddController(controller)

		helpers.DebugLog("Connecting to:", v)
		if err := collector.dialAndHandshake(controller, tag); err != nil {
			helpers.DebugLog("CONTROLLER COM. ERROR:", controller.Address)
			// TODO: make sure this won't hurt functionality later on...
			// collector.RemoveController(controller)
			continue
		}
		helpers.DebugLog("Connected to:", controller.Address)
		go controller.OpenSendChannel()
		sendFirstBasePoint(collector, controller)
	}
}
