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
	Controllers        []*Controller
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
	Indexes            []string
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
	c.Controllers = append(c.Controllers, cont)
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

func (collector *Collector) EngageDataFlow() {
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
			controller.ChangeReceivingStatus(true)
			// send the first base point
			sendFirstBasePoint(collector, controller)
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

func (collector *Collector) MaintainControllerCommunications() {
	for {
		// TODO: implement rand int sleeper
		time.Sleep(time.Duration(collector.MaintainerInterval) * time.Second)
		//helpers.DebugLog("5 second controller maintnance starting ...")
		for _, controller := range collector.Controllers {
			if controller.Active {
				continue
			}
			if err := collector.dialAndHandshake(controller, collector.TAG); err != nil {
				//helpers.DebugLog("CONTROLLER COM. ERROR:", controller.Address)
				continue
			}
			helpers.DebugLog("Recovered connection to:", controller.Address)
			controller.ChangeActiveStatus(true)
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
			if controller.ReadyToReceive {
				controller.Send <- data
			}
		}
	}
}

type Controller struct {
	Address          string
	Active           bool
	HasListener      bool
	IndexesDelivered bool
	ReadyToReceive   bool
	Conn             net.Conn
	Retry            int
	mutex            sync.Mutex
	Send             chan []byte
	//InactiveSince time.Time
}

func (c *Controller) ChangeReceivingStatus(status bool) {
	c.mutex.Lock()
	c.ReadyToReceive = status
	c.mutex.Unlock()
}
func (c *Controller) ChangeActiveStatus(status bool) {
	c.mutex.Lock()
	c.Active = status
	c.mutex.Unlock()
}
func (c *Controller) HaveIndexesBeenDelivered(delivered bool) {
	c.mutex.Lock()
	c.IndexesDelivered = delivered
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

func (c *Controller) OpenSendChannel() {
	defer func() {
		helpers.DebugLog("Closing send loop to controller", c.Address)
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

func (c *Controller) Listen() {
	defer func() {

		helpers.DebugLog("defering read pipe from", c.Address)
		close(c.Send)
		c.ChangeActiveStatus(false)
		c.ChangeListenerStatus(false)
		c.ChangeReceivingStatus(false)
		c.Setconnection(nil)
		c.HaveIndexesBeenDelivered(false)
	}()

	for {
		msg, err := bufio.NewReader(c.Conn).ReadString('\n')
		// TODO: handle better
		if err != nil || msg == "c\n" {
			_ = c.Conn.Close()
			helpers.DebugLog("Error and/or message in read pipe from" + c.Address + " // " + msg + " //" + err.Error())
			break
		}

		helpers.DebugLog("IN:", msg)
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
	// STEP 1
	// send TAG to controller
	_, err = controller.Conn.Write([]byte(tag + "\n"))
	helpers.PanicX(err)

	// STEP 4
	// Listening for indexes from controller
	message, err := bufio.NewReader(controller.Conn).ReadString('\n')
	if !strings.Contains(message, "I:") {
		_ = controller.Conn.Close()
		controller.Setconnection(nil)
		// TODO: handle better
		err = errors.New("indexses not delivered" + message + " // pipe read error was" + err.Error())
		return
	}

	// see readme
	indexes := strings.Split(strings.Split(strings.TrimSuffix(message, "\n"), ":")[1], ",")
	helpers.DebugLog("Indexes:", indexes)
	c.Indexes = indexes
	controller.HaveIndexesBeenDelivered(true)

	// STEP 8
	// sending host data
	data := stats.GetStaticBasePoint()
	_ = c.AddStaticPoint(data)
	_, err = controller.Conn.Write([]byte(data + "\n"))
	helpers.PanicX(err)
	controller.ChangeActiveStatus(true)
	return
}
func (c *Collector) dialAndHandshake(controller *Controller, tag string) (err error) {
	if err = dialController(controller); err != nil {
		return
	}
	if err = c.handShakeWithController(controller, tag); err != nil {
		return
	}
	return
}
func ConnectToControllers(controllers string, tag string, collector *Collector) {
	for _, v := range strings.Split(controllers, ",") {
		controller := &Controller{Address: v, Active: false, HasListener: false, IndexesDelivered: false}
		collector.AddController(controller)

		helpers.DebugLog("Connecting to:", v)
		if err := collector.dialAndHandshake(controller, tag); err != nil {
			helpers.DebugLog("CONTROLLER COM. ERROR:", controller.Address)
			continue
		}
		helpers.DebugLog("Connected to:", controller.Address)

	}
}
