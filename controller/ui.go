package main

import (
	"log"
	"net"
	"sync"

	"github.com/zkynetio/lynx/helpers"
)

type UIServer struct {
	IP         string
	PORT       string
	ClientList []*UI
	mutex      sync.Mutex
}

func (ui *UIServer) AddUI(TAG string, client *UI) {
	ui.mutex.Lock()
	ui.ClientList = append(ui.ClientList, client)
	ui.mutex.Unlock()
}

func (ui *UIServer) RemoveUI(TAG string) {
	ui.mutex.Lock()
	// c.UIS[TAG] = nil
	// TODO...
	ui.mutex.Unlock()
}
func (ui *UIServer) Start(watcherChannel chan int) {
	defer func(watcherChannel chan int) {
		if r := recover(); r != nil {
			log.Println("UI server panic!", r)
		}
		watcherChannel <- 1
	}(watcherChannel)
	helpers.DebugLog("listening on:", ui.IP+":"+ui.PORT)
	ln, err := net.Listen("tcp", ui.IP+":"+ui.PORT)
	helpers.PanicX(err)
	for {
		conn, err := ln.Accept()
		helpers.PanicX(err)
		go ui.AcceptConnection(conn)
	}
}
func (ui *UIServer) AcceptConnection(conn net.Conn) {

	ui.AddUI("meow", &UI{
		Conn:        conn,
		DataChannel: make(chan []byte),
	})

	// start handshake..
}

type UI struct {
	Conn        net.Conn
	Config      Config
	DataChannel chan []byte
}
type Group struct {
	DD []*DimentionalData
}

type DimentionalData struct {
	X          int64
	Y          int64
	Z          int64
	Blink      int64
	Luminocity int64
	Tag        string
	Size       int
}

type Config struct {
	Z          Z
	Y          Y
	X          X
	Blink      Blink
	Luminocity Luminocity
	Size       Size
	// Rate in milliseconds
	UpdateRate int
	Paused     bool
}

type Blink struct {
	Index     int
	Normalize bool
}
type Luminocity struct {
	Index     int
	Normalize bool
}

type Size struct {
	Index     int
	Normalize bool
}

type X struct {
	Index     int
	Normalize bool
}
type Y struct {
	Index     int
	Normalize bool
}
type Z struct {
	Index     int
	Normalize bool
}

func meow() {

}

// func connectUI(ui *UI, controller *Controller, tag string) {
// 	helpers.DebugLog("UI:")
// 	controller.AddUI(tag+"TODORANDOMINT?", ui)
// 	for {
// 		outgoing := <-ui.SendChannel
// 		log.Println(outgoing)
// 		_, err := ui.Conn.Write([]byte(outgoing))
// 		//TODO: handle better
// 		helpers.PanicX(err)
// 	}
// }

// func (c *Controller) sendToAllUis(msg string) {
// 	for _, ui := range c.UIs {
// 		ui.SendChannel <- msg
// 	}
// }
