package main

import (
	"bufio"
	"bytes"
	"errors"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/zkynetio/lynx/helpers"
)

func main() {
	rand.Seed(time.Now().Unix())
	loadEnvironmentVariables()

	controller := Controller{
		IP:                   os.Getenv("IP"),
		PORT:                 os.Getenv("PORT"),
		Collectors:           make(map[string]*Collector),
		Buffer:               make(chan string, 10000),
		BufferDirectoryPath:  "./buffers/",
		MinLinesInBufferFile: 10,
	}

	defer controller.CleanupOnExit()
	go controller.EngageBufferPipe()
	controller.start()

	// capture stop signal in order to exit gracefully.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}

type UI struct {
	Conn        net.Conn
	Filter      string
	SendChannel chan string
}
type Stat struct {
	Label       string
	Value       int
	StringValue string
}

type Controller struct {
	Buffer chan string
	// Recovery in sqlite?
	//RecoveryFile string
	BufferDirectoryPath  string
	PORT                 string
	IP                   string
	Collectors           map[string]*Collector
	UIs                  map[string]*UI
	mutex                sync.Mutex
	MinLinesInBufferFile int
}

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

func loadEnvironmentVariables() {
	err := godotenv.Load(".env")
	if err != nil {
		helpers.DebugLog(err)
		log.Fatal("Error loading .env file")
	}
}
func (c *Controller) AddCollector(TAG string, collector *Collector) {
	c.mutex.Lock()
	c.Collectors[TAG] = collector
	c.mutex.Unlock()
}

func (c *Controller) RemoveCollector(TAG string) {
	c.mutex.Lock()
	c.Collectors[TAG] = nil
	c.mutex.Unlock()
}
func (c *Controller) AddUI(TAG string, ui *UI) {
	c.mutex.Lock()
	c.UIs[TAG] = ui
	c.mutex.Unlock()
}

func (c *Controller) RemoveUI(TAG string) {
	c.mutex.Lock()
	c.UIs[TAG] = nil
	c.mutex.Unlock()
}

func (controller *Controller) start() {

	helpers.DebugLog("listening on:", controller.IP+":"+controller.PORT)
	ln, err := net.Listen("tcp", controller.IP+":"+controller.PORT)
	helpers.PanicX(err)
	for {
		conn, err := ln.Accept()
		helpers.PanicX(err)
		go receiveConnection(conn, controller)
	}
}

func connectUI(ui *UI, controller *Controller, tag string) {
	helpers.DebugLog("UI:")
	controller.AddUI(tag+"TODORANDOMINT?", ui)
	for {
		outgoing := <-ui.SendChannel
		log.Println(outgoing)
		_, err := ui.Conn.Write([]byte(outgoing))
		//TODO: handle better
		helpers.PanicX(err)
	}
}

func deliverNamespaces(conn net.Conn) {
	_, err := conn.Write([]byte("NS|general.entropy:5,memory.total:5\n"))
	helpers.PanicX(err)
}
func receiveConnection(conn net.Conn, controller *Controller) {
	helpers.DebugLog("Connection from:", conn.RemoteAddr())

	message, _ := bufio.NewReader(conn).ReadString('\n')
	if message == "ui\n" {
		connectUI(&UI{Conn: conn}, controller, message)
	} else {
		// STEP 2
		// Connect a collector
		connectCollector(&Collector{
			LastCheckin: time.Now(),
			Conn:        conn,
		}, controller, message)
	}
}
func connectCollector(collector *Collector, controller *Controller, message string) {
	helpers.DebugLog("COLLECTOR:", strings.TrimSuffix(message, "\n"))
	collector.TAG = strings.TrimSuffix(message, "\n")

	controller.AddCollector(collector.TAG, collector)
	defer func() {
		helpers.DebugLog("Closing read pipe from", collector.TAG)
		controller.RemoveCollector(collector.TAG)
	}()

	// STEP 3
	// Send namespaces to collector
	deliverNamespaces(collector.Conn)

	// STEP 6
	// Listen for an OK
	msg, _ := bufio.NewReader(collector.Conn).ReadString('\n')
	if msg == "k\n" {
		helpers.DebugLog("K from collector")
	} else {
		helpers.DebugLog(msg)
		helpers.PanicX(errors.New("NOT OK FROM COLLECTOR"))
	}

	// STEP 7
	// Listen for host data
	msg, _ = bufio.NewReader(collector.Conn).ReadString('\n')
	if !strings.Contains(string(msg), "H||") {
		helpers.PanicX(errors.New("no host data found in handskae" + string(msg)))
	}
	helpers.DebugLog("HOST DATA:", string(msg))

	// STEP 9
	// Listen for final OK
	helpers.DebugLog("waiting for final ok ...")
	msg, _ = bufio.NewReader(collector.Conn).ReadString('\n')
	if msg != "k\n" {
		helpers.DebugLog(msg)
		helpers.PanicX(errors.New("NOT OK FROM COLLECTOR"))
	}

	log.Println("Starting general collection from:", collector.TAG)
	readFromConnectionOriginal(collector, controller)
}
func readFromConnectionOriginal(collector *Collector, controller *Controller) {
	for {
		message, err := bufio.NewReader(collector.Conn).ReadString('\n')
		if err != nil {
			// TODO: handle better
			helpers.DebugLog("ERROR IN READ LOOP:", err)
			err = collector.Conn.Close()
			if err != nil {
				// TODO: handle better
				helpers.DebugLog("ERROR CLOSING INSIDE READ LOOP:", err)
				break
			}
		}
		helpers.DebugLog("Length of message:", strconv.Itoa(len(message)))
		go controller.parseIncomingData(collector.TAG + ":::" + message)
		//TODO: IMPLEMENT SQLITE
		//helpers.DebugLog("IN:", string(message))

	}
}

func (c *Controller) sendToAllUis(msg string) {
	for _, ui := range c.UIs {
		ui.SendChannel <- msg
	}
}
func (c *Controller) parseIncomingData(msg string) {
	msg = strings.TrimSuffix(msg, "\n")

	//helpers.DebugLog("DATA:", msg)
	c.Buffer <- msg
}
func (c *Controller) EngageBufferPipe() {

	var buffer bytes.Buffer
	// TODO: grow properly, we are writing strings.. but the buffer grows in bytes
	//	buffer.Grow(c.MinLinesInBufferFile * 500)
	count := 0
	for {
		if len(c.Buffer) > 100 {
			//buffer.Grow(buffer.Len() * 2)
			helpers.DebugLog("chan length", len(c.Buffer))
		}

		message := <-c.Buffer
		//helpers.DebugLog("CONTROLLER BUFFER RECEIVED:", message)

		go c.sendToAllUis(message)
		_, err := buffer.WriteString(message)

		//helpers.DebugLog("wrote", data)
		if err != nil {
			helpers.DebugLog(err)
			break
		}

		count++
		//helpers.DebugLog(count)
		if count > c.MinLinesInBufferFile && count >= len(c.Collectors)-1 {
			go c.WriteBufferToFile(buffer)
			buffer.Reset()
			count = 0
		}
	}
}

func (c *Controller) WriteBufferToFile(buffer bytes.Buffer) {

	now := time.Now().Format(time.RFC3339Nano)
	now = strings.Replace(now, "-", "/", -1)
	now = strings.Replace(now, "T", "/", -1)
	now = strings.Replace(now, ":", "/", 1)
	//	helpers.DebugLog(strings.Split(now, ":")[0])
	err := os.MkdirAll(c.BufferDirectoryPath+strings.Split(now, ":")[0], 0700)
	now = strings.Replace(now, ":", "/", 1)
	helpers.PanicX(err)
	helpers.DebugLog("writing to file:", c.BufferDirectoryPath+now)
	file, err := os.OpenFile(c.BufferDirectoryPath+now, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	helpers.PanicX(err)
	buffer.WriteTo(file)
}
