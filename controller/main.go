package main

import (
	"bufio"
	"encoding/binary"
	"errors"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/zkynetio/lynx/helpers"
)

func main() {
	rand.Seed(time.Now().Unix())

	Settings := &Settings{}
	Settings.LoadConfigFromFile("config.yml")

	log.Println("Settings:", Settings)
	UIServer := &UIServer{
		Settings:   Settings,
		ClientList: make(map[string]*UI),
		DPChan:     make(chan DataPoint, 1000000),
	}
	go UIServer.ShipToUIS()
	controller := Controller{
		Settings:             Settings,
		Collectors:           make(map[string]*Collector),
		Buffer:               make(chan *DataPoint, 1000000),
		BufferDirectoryPath:  "./buffers/",
		MinLinesInBufferFile: 10,
		UI:                   UIServer,
	}

	LiveBuffer = &Collection{
		Map: make(map[string]map[string]map[uint64][]byte),
	}
	defer controller.CleanupOnExit()
	go controller.EngageBufferPipe()
	watcherChannel := make(chan int)

	go UIServer.Start(watcherChannel)
	go controller.start(watcherChannel)
	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt)
	for {
		select {
		case index := <-watcherChannel:
			if index == 1 {
				go UIServer.Start(watcherChannel)
			} else if index == 2 {
				go controller.start(watcherChannel)
			}
			log.Println("goroutine number", index, "just closed...")
			break
		case <-stop:
			// TODO: handle exit gracefully
			log.Println("handle exit gracefully")
			os.Exit(1)
		}
	}

}

type Controller struct {
	Buffer chan *DataPoint
	// Recovery in sqlite?
	//RecoveryFile string
	Settings             *Settings
	BufferDirectoryPath  string
	PORT                 string
	IP                   string
	Collectors           map[string]*Collector
	UIs                  map[string]*UI
	mutex                sync.Mutex
	MinLinesInBufferFile int
	UI                   *UIServer
}
type Collection struct {
	// instace,namespace,[year.month.day.hour.minute.second],[]byte
	Map map[string]map[string]map[uint64][]byte
	Mux sync.Mutex
}

var LiveBuffer *Collection

// mat zoe? abive and beyond
func (c *Controller) CleanupOnExit() {
	helpers.DebugLog("Cleaning up on exit...")
	for _, collector := range c.Collectors {
		if collector.Conn != nil {
			_, _ = collector.Conn.Write([]byte("c\n"))
			helpers.DebugLog("Closing:", collector.TAG)
			_ = collector.Conn.Close()
		}
	}
}

type Collector struct {
	TAG         string
	Conn        net.Conn
	LastCheckin time.Time
	SendChannel chan string
}

func (c *Controller) AddCollector(TAG string, collector *Collector) error {
	c.mutex.Lock()
	if c.Collectors[TAG] != nil {
		log.Println("A controller already exists with this tag:", TAG)
		return errors.New("This TAG already exists")
	}
	c.Collectors[TAG] = collector
	c.mutex.Unlock()
	return nil
}

func (c *Controller) RemoveCollector(TAG string) {
	c.mutex.Lock()
	c.Collectors[TAG] = nil
	c.mutex.Unlock()
}

func (controller *Controller) start(watcherChannel chan int) {
	defer func(watcherChannel chan int) {
		if r := recover(); r != nil {
			log.Println("controller crahed !!", r)
		}
		watcherChannel <- 2
	}(watcherChannel)
	helpers.DebugLog("listening on:", controller.Settings.IP+":"+controller.Settings.PORT)
	ln, err := net.Listen("tcp", controller.Settings.IP+":"+controller.Settings.PORT)
	helpers.PanicX(err)
	for {
		conn, err := ln.Accept()
		helpers.PanicX(err)
		go receiveConnection(conn, controller)
	}
}

func receiveConnection(conn net.Conn, controller *Controller) {
	defer func() {
		r := recover()
		if r != nil {
			helpers.DebugLog("recovered in connection receiver ", r)
		}
	}()
	helpers.DebugLog("Connection from:", conn.RemoteAddr())

	message, _ := bufio.NewReader(conn).ReadString('\n')

	// STEP 2
	// Connect a collector
	connectCollector(&Collector{
		LastCheckin: time.Now(),
		Conn:        conn,
	}, controller, message)
}
func connectCollector(collector *Collector, controller *Controller, message string) {
	collector.TAG = strings.TrimSuffix(message, "\n")

	helpers.DebugLog("COLLECTOR:", collector.TAG)
	err := controller.AddCollector(collector.TAG, collector)
	if err != nil {
		_, _ = collector.Conn.Write([]byte("E: this TAG already exists\n"))
		_ = collector.Conn.Close()
		return
	}
	defer func() {
		helpers.DebugLog("Closing read pipe from", collector.TAG)
		controller.RemoveCollector(collector.TAG)
	}()

	msg, _ := bufio.NewReader(collector.Conn).ReadString('\n')
	if !strings.Contains(string(msg), "H||") {
		helpers.PanicX(errors.New("no host data found in handskae" + string(msg)))
	}
	helpers.DebugLog("HOST DATA:", string(msg))

	_, err = collector.Conn.Write([]byte("k\n"))
	if err != nil {
		helpers.DebugLog("could not establish coms with collector", err)
		controller.RemoveCollector(collector.TAG)
		return
	}

	log.Println("Starting general collection from:", collector.TAG)
	readFromConnectionOriginal(collector, controller)
}
func readFromConnectionOriginal(collector *Collector, controller *Controller) {
	reader := bufio.NewReader(collector.Conn)
	// TODO: move all parsing into go routine ?
	for {
		// log.Println("loop one!")
		controlBytes := make([]byte, 3)
		_, err := reader.Read(controlBytes)
		if err != nil {
			panic(err)
		}
		// log.Println(controlBytes)
		length := binary.LittleEndian.Uint16(controlBytes[0:3])
		// log.Println("Length:", length)
		// log.Println("control byte:", controlBytes[2])
		// timestamp := binary.LittleEndian.Uint64(controlBytes[3:])

		// log.Println("timestamp", timestamp, "MS:", (timestamp / 1000000), controlBytes[3:])
		// +8 for the timestamp at the start of the data
		data := make([]byte, length+8)
		_, err = reader.Read(data)
		if err != nil {
			panic(err)
		}
		// log.Println("data from read:", data)
		// ParseDataPoint(data)

		go controller.parseIncomingData(collector.TAG, data, int(controlBytes[2]))
		go controller.sendToUIS(collector.TAG, data)

	}

}

type DataPoint struct {
	Value       []byte
	Tag         string
	Timestamp   int64
	ControlByte int
}

func (c *Controller) parseIncomingData(tag string, data []byte, controlByte int) {

	//helpers.DebugLog("DATA:", msg)
	c.Buffer <- &DataPoint{
		Value:       data,
		Tag:         tag,
		ControlByte: controlByte,
	}
}
func (c *Controller) sendToUIS(tag string, data []byte) {

	//helpers.DebugLog("DATA:", msg)
	c.UI.DPChan <- DataPoint{
		Value: data,
		Tag:   tag,
	}
}
func (c *Controller) EngageBufferPipe() {

	// TODO: grow properly, we are writing strings.. but the buffer grows in bytes
	//	buffer.Grow(c.MinLinesInBufferFile * 500)
	// size := 0
	for {
		if len(c.Buffer) > 100 {
			//buffer.Grow(buffer.Len() * 2)
			helpers.DebugLog("chan length", len(c.Buffer))
		}

		// CollectionBuffer.mux.Lock()
		// c.CollectionBuffer[dp.Tag] = append(c.CollectionBuffer[dp.Tag], dp)
		// CollectionBuffer.mux.Unlock()
		c.ParseDataPointIntoMemoryMap(<-c.Buffer)
		// log.Println(c.CollectionBuffer)
		// size = size + len(dp.Value)
		// if size > 3000 {
		// 	// write data
		// 	// go c.WriteBufferToFile()
		// }

	}
}
func (c *Controller) ParseDataPointIntoMemoryMap(dp *DataPoint) {

	// timestamp := dp.Timestamp.Unix()
	// tag := dp.Tag
	// log.Println(dp.Value)
	timestamp := binary.LittleEndian.Uint64(dp.Value[:8])
	// log.Println("data starting sequence:", dp.Value[8:20])
	// log.Println("timestamp", timestamp, dp.Value[:8])

	// tm := time.Unix(0, int64(timestamp))
	// fmt.Println(tm.Format(time.RFC3339))

	log.Println("Data from:", dp.Tag)
	// log.Println(dp.Value[8:])
	// ParseDataPoint(dp.Value[8:])
	// for _, v := range dp.Value {
	if LiveBuffer.Map[dp.Tag] == nil {
		LiveBuffer.Map[dp.Tag] = make(map[string]map[uint64][]byte)
	}
	if LiveBuffer.Map[dp.Tag][dp.Tag] == nil {
		LiveBuffer.Map[dp.Tag][dp.Tag] = make(map[uint64][]byte)
	}

	LiveBuffer.Mux.Lock()
	LiveBuffer.Map[dp.Tag][dp.Tag][timestamp] = dp.Value
	LiveBuffer.Mux.Unlock()
	//
	// log.Println(LiveBuffer.Map)
	// for i, v := range LiveBuffer.Map {
	// 	for _, iv := range v {
	// 		for iii, _ := range iv {
	// 			log.Println("A Record:", i, iii)
	// 		}
	// 	}
	// }
}

// func (c *Controller) WriteBufferToFile() {

// 	now := time.Now().Format(time.RFC3339Nano)
// 	now = strings.Replace(now, "-", "/", -1)
// 	now = strings.Replace(now, "T", "/", -1)
// 	now = strings.Replace(now, ":", "/", 1)
// 	//	helpers.DebugLog(strings.Split(now, ":")[0])
// 	err := os.MkdirAll(c.BufferDirectoryPath+strings.Split(now, ":")[0], 0700)
// 	now = strings.Replace(now, ":", "/", 1)
// 	helpers.PanicX(err)
// 	helpers.DebugLog("writing to file:", c.BufferDirectoryPath+now)
// 	file, err := os.OpenFile(c.BufferDirectoryPath+now, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
// 	helpers.PanicX(err)
// 	for i, v := range c.CollectionBuffer {
// 		// file.Write(v. )
// 		for ii, iv := range v {
// 			log.Println("FileName:", iv.Tag)
// 			// 1. open file by tag/datapoint/timestamp
// 			// 2. write contents

// 		}
// 	}
// 	// buffer.WriteTo(file)
// }
