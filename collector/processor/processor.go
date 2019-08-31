package processor

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"log"
	"net"
	"os"
	"runtime/debug"
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

func sendBasePoint(collector *Collector, controller *Controller) {
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
		log.Println("Stopped maintining controllers...")
	}(watcherChannel)
	for {
		// TODO: implement rand int sleeper
		time.Sleep(time.Duration(collector.MaintainerInterval) * time.Second)
		//helpers.DebugLog("5 second controller maintnance starting ...")
		log.Println("Number of controllers to maintain:", len(collector.Controllers))
		for _, controller := range collector.Controllers {
			log.Println("maintaining controller:", controller)
			if controller.Active {
				log.Println("Controller is already active...")
				continue
			}
			if err := collector.dialAndHandshake(controller, collector.TAG); err != nil {
				helpers.DebugLog("CONTROLLER COM. ERROR:", controller.Address)
				continue
			}
			helpers.DebugLog("Recovered connection to:", controller.Address)
			controller.ChangeActiveStatus(true)

			helpers.DebugLog("Engaging controller listener to", controller.Address)
			go controller.OpenSendChannel()
			// sendBasePoint(collector, controller)
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

	// count := collector.CurrentPointIndex
	startTime := time.Now()
	for {
		var data []byte
		if !time.Now().After(startTime.Add(1000 * time.Millisecond)) {
			time.Sleep(50 * time.Millisecond)
			continue
		}
		startTime = time.Now()
		go func() {
			// time.Sleep(time.Duration(collector.CollectionInterval) * time.Millisecond)
			var ControlByte byte
			// if count%60 == 0 {
			// 	staticData := stats.CheckStaticDataForChanges()
			// 	if staticData != "" {
			// 		log.Println("do something with static data !")
			// 	}
			// 	ControlByte = 1
			// 	data = stats.CollectBasePoint()
			// 	collector.LastBasePointIndex = count
			// 	count = collector.AddDataPoint(data)
			// } else {
			data = stats.GetMinimumStats()
			// count = collector.AddDataPoint(data)
			// control byte 4 means this is a minimal data point
			ControlByte = 4
			// }

			// accumilatedBytes := 0
			// pointCount := 0
			// var allBytes []byte
			// for _, v := range collector.PointMap {
			// 	accumilatedBytes = accumilatedBytes + len(v)
			// 	allBytes = append(allBytes, v...)
			// 	pointCount++
			// }
			// var b bytes.Buffer
			// w := zlib.NewWriter(&b)
			// w.Write(allBytes)
			// w.Close()
			// helpers.DebugLog("Current DP size:", len(data), "Accumilated DP size:", accumilatedBytes, " Compressed:", b.Len(), "DP count:", pointCount)
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
			helpers.DebugLog("ERROR WHEN WRITING STATS (Count", n, ") err:", err)
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
	log.Println("HANDSHAKING!")
	// data := stats.GetStaticBasePoint()
	// _, err = controller.Conn.Write([]byte(data + "\n"))
	// helpers.PanicX(err)

	msg, err := bufio.NewReader(controller.Conn).ReadString('\n')
	if msg != "k\n" || err != nil {
		log.Println("Coult not handshake with controller", msg)
		if err != nil {
			return err
		}

		return errors.New("Could not handshake with controller")
	}
	log.Println("DONE HANDSHAKING!")
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
		// sendBasePoint(collector, controller)
	}
}
