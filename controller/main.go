package controller

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/zkynetio/lynx/alerting"
	"github.com/zkynetio/lynx/helpers"
	"github.com/zkynetio/lynx/namespaces"
	"github.com/zkynetio/lynx/ui"
	"github.com/zkynetio/safelocker"
)

var GlobalController *Controller
var GlobalBrain *Brain

type Brain struct {
	safelocker.SafeLocker
	Socket      net.Conn
	Address     string
	SendChannel chan []byte
	Alerting    []alerting.Alerting `json:"alerting"`
	Collecting  Collecting          `json:"collecting"`
}
type Collecting struct {
	Default []struct {
		Tag        string   `json:"tag"`
		Namespaces []string `json:"namespaces"`
	} `json:"default"`
	Custom []struct {
		Tag        string   `json:"tag"`
		Namespaces []string `json:"namespaces"`
	} `json:"custom"`
}
type ControllerConfig struct {
	Restart  bool
	IP       string
	Debug    bool
	Shutdown bool
	UI       struct {
		IP   string
		Port int
	}
	Collector struct {
		IP   string
		Port int
	}
}

func Start(address string) {
	rand.Seed(time.Now().Unix())
	namespaces.Init()
	Brain := Brain{
		Address: address,
	}
	GlobalBrain = &Brain
	socket, err := net.Dial("tcp", Brain.Address)
	if err != nil {
		panic(err)
	}

	lengthBytes := make([]byte, 2)
	_, err = socket.Read(lengthBytes)
	if err != nil {
		log.Println("Error reading length bytes from brain", err)
		os.Exit(1)
	}
	data := make([]byte, binary.LittleEndian.Uint16(lengthBytes))
	_, err = socket.Read(data)
	if err != nil {
		log.Println("Error reading data fom brain ...", err)
		os.Exit(1)
	}

	// log.Println(string(data))
	var config ControllerConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Println("Something went wrong when receiving a config from the brain", err)
		os.Exit(1)
	}
	log.Println("GOT A CONFIG:", config)
	// os.Exit(1)
	Brain.Socket = socket
	Brain.SendChannel = make(chan []byte, 10000)

	UIServer := ui.NewUIServer()
	UIServer.ClientList = make(map[string]*ui.UI)
	UIServer.IP = config.UI.IP
	UIServer.Port = strconv.Itoa(config.UI.Port)

	controller := Controller{
		Config:     config,
		Collectors: make(map[string]*Collector),
		Buffer:     make(chan *DataPoint, 1000000),
		UIServer:   UIServer,
	}
	// os.Exit(1)
	watcherChannel := make(chan int, 10)
	closeChannel := make(chan bool, 10)
	go Brain.MaintainLinkToBrain(watcherChannel, closeChannel)
	controller.UISendChannel = make(chan []byte, 1000000)
	controller.UIParseChannel = make(chan DPCollection, 100000)
	go ShipToUIS(controller.UISendChannel)
	go SaveToUIBuffer(controller.UIParseChannel, controller.UISendChannel)

	GlobalController = &controller

	LiveBuffer = &Collection{
		Map:               make(map[string]map[string]map[uint64][]byte),
		CurrentBase:       make(map[string]map[string]int64),
		CollectorStatsMap: make(map[string]map[string]int64),
	}
	defer controller.CleanupOnExit()
	go controller.EngageBufferPipe()

	go UIServer.Start(watcherChannel)
	go controller.start(watcherChannel, closeChannel)
	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt)
	for {
		select {
		case index := <-watcherChannel:
			time.Sleep(1 * time.Second)
			if index == 1 {
				go UIServer.Start(watcherChannel)
			} else if index == 2 {
				go controller.start(watcherChannel, closeChannel)
			} else if index == 3 {
				go Brain.MaintainLinkToBrain(watcherChannel, closeChannel)
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

func (b *Brain) MaintainLinkToBrain(watcherChannel chan int, closeChannel chan bool) {
	defer func(watcherChannel chan int) {
		if r := recover(); r != nil {
			log.Println("Brain link crahed !!", r, string(debug.Stack()))
		}
		watcherChannel <- 3
	}(watcherChannel)

	// buf := make([]byte, 20000)
	for {
		log.Println("listening to brain ...")
		// make a decoder
		lengthBytes := make([]byte, 2)
		_, err := b.Socket.Read(lengthBytes)
		if err != nil {
			log.Println("Error reading length bytes from brain", err)
			return
		}
		data := make([]byte, binary.LittleEndian.Uint16(lengthBytes))
		_, err = b.Socket.Read(data)
		if err != nil {
			log.Println("Error reading data fom brain ...", err)
		}

		err = b.DecodeConfig(closeChannel, data)
		if err == nil {
			continue
		}
		err = b.DecodeBrain(data)
		if err == nil {
			continue
		}

	}
}
func (b *Brain) DecodeBrain(data []byte) error {
	var brain Brain
	// log.Println(string(data))
	err := json.Unmarshal(data, &brain)
	if err != nil {
		log.Println("could not decode brain", err)
		return err
	}
	b.Lock()
	if len(brain.Alerting) > 0 {
		b.Alerting = brain.Alerting
	}
	if len(brain.Collecting.Custom) > 0 || len(brain.Collecting.Default) > 0 {
		b.Collecting = brain.Collecting
	}

	b.Unlock()
	log.Println(b.Collecting)
	log.Println(b.Alerting)
	return nil
}
func (b *Brain) DecodeConfig(closeChannel chan bool, data []byte) error {
	defer func() {
		if r := recover(); r != nil {
			log.Println("crashed while restarting..", r, string(debug.Stack()))
			GlobalController.SafeUnlock()
		}

	}()

	// Decode the config
	var config ControllerConfig
	err := json.Unmarshal(data, &config)
	if err != nil {
		log.Println(err)
		return err
	} else if config.IP == "" {
		log.Println("IP missing, Assuming an errror")
		return errors.New("IP missing, Assuming an errror")
	} else if config.Shutdown {
		log.Println("Brain told me to exit...")
		os.Exit(1)
	} else if config.Restart {
		rand.Seed(time.Now().UnixNano())
		n := rand.Intn(5)
		time.Sleep(time.Duration(n) * time.Second)
		ui.UIHTTPsrv.Close()
		GlobalController.Lock()
		for _, v := range GlobalController.UIServer.ClientList {
			if v != nil {
				_ = v.Conn.Close()
			}
		}
		for _, v := range GlobalController.Collectors {
			if v != nil {
				_ = v.Conn.Close()
			}
		}

		GlobalController.Unlock()
		closeChannel <- true
	}

	GlobalController.Config = config
	return nil

}

type Controller struct {
	safelocker.SafeLocker
	Buffer chan *DataPoint
	// Recovery in sqlite?
	//RecoveryFile string
	Config         ControllerConfig
	PORT           string
	IP             string
	Collectors     map[string]*Collector
	UIs            map[string]*ui.UI
	UIServer       *ui.UIServer
	UIParseChannel chan DPCollection
	UISendChannel  chan []byte
}
type Collection struct {
	// instace,namespace,[year.month.day.hour.minute.second],[]byte
	Map               map[string]map[string]map[uint64][]byte
	CurrentBase       map[string]map[string]int64
	CollectorStatsMap map[string]map[string]int64
	Mux               sync.Mutex
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
	Namespaces  map[int]string
}

func (c *Controller) AddCollector(TAG string, collector *Collector) error {
	c.Lock()
	defer c.Unlock()
	if c.Collectors[TAG] != nil {
		log.Println("A controller already exists with this tag:", TAG)
		return errors.New("This TAG already exists")
	}
	c.Collectors[TAG] = collector
	return nil
}

func (c *Controller) RemoveCollector(TAG string) {
	c.Lock()
	c.Collectors[TAG] = nil
	c.Unlock()
}

func (controller *Controller) start(watcherChannel chan int, closeChannel chan bool) {
	defer func(watcherChannel chan int) {
		if r := recover(); r != nil {
			log.Println("controller crahed !!", r, string(debug.Stack()))
		}
		watcherChannel <- 2
	}(watcherChannel)
	helpers.DebugLog("listening on:", controller.Config.Collector.IP+":"+strconv.Itoa(controller.Config.Collector.Port))
	ln, err := net.Listen("tcp", controller.Config.Collector.IP+":"+strconv.Itoa(controller.Config.Collector.Port))
	helpers.PanicX(err)
	for {
		conn, err := ln.Accept()
		log.Println("Checking close channel ...")
		select {
		case <-closeChannel:
			log.Println("closing collector port..")
			_ = ln.Close()
			_ = conn.Close()
			return
		default:
			log.Println("nothing to close ... ")
		}
		helpers.PanicX(err)
		go receiveConnection(conn, controller)

	}
}

func receiveConnection(conn net.Conn, controller *Controller) {
	defer func() {
		r := recover()
		if r != nil {
			helpers.DebugLog("recovered in connection receiver ", r, string(debug.Stack()))
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
func findCollectorNamespaces(collector *Collector) (namespaces []string) {
	var defaultNamespaces []string
	for _, v := range GlobalBrain.Collecting.Default {
		if strings.Contains(collector.TAG, v.Tag) || v.Tag == "*" {
			defaultNamespaces = append(defaultNamespaces, v.Namespaces...)
		}
	}
	var customNamespaces []string
	for _, v := range GlobalBrain.Collecting.Custom {
		if strings.Contains(collector.TAG, v.Tag) {
			customNamespaces = append(customNamespaces, v.Namespaces...)
		}
	}

	if len(customNamespaces) > 0 {
		log.Println("USING CUSTOM INDEXES:", customNamespaces)
		namespaces = customNamespaces
	} else {
		log.Println("USING DEFAULT INDEXES:", defaultNamespaces)
		namespaces = defaultNamespaces
	}
	return
}
func connectCollector(collector *Collector, controller *Controller, message string) {
	collector.TAG = strings.TrimSuffix(message, "\n")
	ns := findCollectorNamespaces(collector)
	collector.Namespaces = namespaces.MakeMapFromNamespaces(ns)

	defer func() {
		helpers.DebugLog("Closing read pipe from", collector.TAG)
		controller.RemoveCollector(collector.TAG)
	}()

	helpers.DebugLog("COLLECTOR:", collector.TAG)
	err := controller.AddCollector(collector.TAG, collector)
	if err != nil {
		_, _ = collector.Conn.Write([]byte("E: this TAG already exists\n"))
		_ = collector.Conn.Close()
		return
	}

	jsonIndexes, err := json.Marshal(ns)
	if err != nil {
		panic(err)
	}
	log.Println("SENDING INDEXES:", string(jsonIndexes))
	_, err = collector.Conn.Write([]byte(string(jsonIndexes) + "\n"))
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
		controlBytes := make([]byte, 3)
		_, err := reader.Read(controlBytes)
		if err != nil {
			log.Println("Closing Collector read pipe 1:", err)
			collector.Conn.Close()
			return
		}
		length := binary.LittleEndian.Uint16(controlBytes[0:2])
		// log.Println("DATA LENGTH BYTES:", controlBytes[0:2], "length:", length)
		data := make([]byte, length+8)
		_, err = reader.Read(data)
		if err != nil {
			log.Println("Closing Collector read pipe 2:", err)
			collector.Conn.Close()
			return
		}

		go controller.HandleDataPoint(collector, data, int(controlBytes[2]))

	}

}

type DataPoint struct {
	Value       []byte
	Tag         string
	Timestamp   int64
	ControlByte int
}

func (c *Controller) HandleDataPoint(collector *Collector, data []byte, controlByte int) {
	defer func() {
		// recover if we get a broken data point.
		r := recover()
		if r != nil {
			log.Println("panic!", r, string(debug.Stack()))
		}
	}()
	var DPC DPCollection
	if controlByte == 4 {
		// log.Println("TIME:", timestamp)
		// log.Println("FULL DATA:", data)
		// log.Println("TAG:", tag, " CONTROL BYTE:", controlByte)
		DPC.Timestamp = binary.LittleEndian.Uint64(data[:8])
		DPC.Tag = collector.TAG
		DPC.ControlByte = controlByte
		DPC = ParseMinimumDataPoint(data[8:], collector.Namespaces)

		log.Println()
		for _, v := range DPC.DPS {
			fmt.Print(v.Index, "/", v.Value, "  - ")
		}

		// log.Println(DPC)
	}

	c.UIParseChannel <- DPC
}

func (c *Controller) saveData(tag string, data []byte, controlByte int) {
	c.Buffer <- &DataPoint{
		Value:       data,
		Tag:         tag,
		ControlByte: controlByte,
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
	//
	// log.Println("Data from:", dp.Tag)
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
