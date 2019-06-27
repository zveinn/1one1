package main

import (
	"encoding/binary"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type UIServer struct {
	ClientList map[string]*UI
	mutex      sync.Mutex
	Settings   *Settings
	DPChan     chan DataPoint
}
type UI struct {
	Conn        *websocket.Conn
	Config      Config
	DataChannel chan []byte
	Buffer      []DataPoint
}

func (uis *UIServer) ShipToUIS() {
	for {
		dp := <-uis.DPChan
		for _, v := range uis.ClientList {
			dp.Timestamp = int64(binary.LittleEndian.Uint64(dp.Value[:8]))
			v.Conn.WriteJSON(dp)
		}
	}
}

// func AddDataPointToUIS(dp *DataPoint){
// 	for i, v := range
// }

func (uis *UIServer) AddUI(TAG string, client *UI) {
	uis.mutex.Lock()
	uis.ClientList[TAG] = client
	uis.mutex.Unlock()
}

func (uis *UIServer) RemoveUI(TAG string) {
	uis.mutex.Lock()
	// ui.ClientList[TAG] = nil
	delete(uis.ClientList, TAG)
	uis.mutex.Unlock()
}
func (uis *UIServer) Start(watcherChannel chan int) {
	defer func(watcherChannel chan int) {
		if r := recover(); r != nil {
			log.Println("UI server panic!", r)
		}
		watcherChannel <- 1
	}(watcherChannel)

	http.HandleFunc("/", uis.wsHandler)
	panic(http.ListenAndServe(uis.Settings.UIIP+":"+uis.Settings.UIPORT, nil))

}

func (uis *UIServer) wsHandler(w http.ResponseWriter, r *http.Request) {

	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}
	go uis.AcceptConnection(conn)

}

func (uis *UIServer) AcceptConnection(conn *websocket.Conn) {

	uis.AddUI("meow", &UI{
		Conn: conn,
		// we need this in case the client disconnects.
		DataChannel: make(chan []byte),
	})

	for {
		// log.Println("reading ....")
		_, msg, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			return
		}
		// msh = strings.Trim(configData, "\n")

		log.Println("from ui", msg)
	}

	// config := &Config{
	// 	Blink:      Blink{},
	// 	X:          X{},
	// 	Y:          Y{},
	// 	Z:          Z{},
	// 	Luminocity: Luminocity{},
	// 	Size:       Size{},
	// }
	// err := json.Unmarshal([]byte(configData), config)
	// helpers.DebugLog(err)
	// start handshake..
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
	UpdateRate   int
	wantsUpdates bool
	Indexes      []int
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
