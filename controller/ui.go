package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type UIServer struct {
	ClientList map[string]*UI
	Mutex      sync.RWMutex
	Settings   *Settings
	DPChan     chan ParsedCollection
	DPBuffer   []ParsedCollection
}
type UI struct {
	Conn        *websocket.Conn
	Config      Config
	DataChannel chan []byte
	Buffer      []*OutgoingData
	Mux         sync.RWMutex
}

func (uis *UIServer) ShipToUIS() {

	for {
		// time.Sleep(1 * time.Second)
		go uis.AddToBuffer(<-uis.DPChan)
		// for _, v := range uis.ClientList {

		// 	dp.Timestamp = int64(binary.LittleEndian.Uint64(dp.Value[:8]))
		// 	v.Conn.WriteJSON(dp)
		// }
	}
}

func (uis *UIServer) AddToBuffer(pc ParsedCollection) {
	uis.Mutex.Lock()
	uis.DPBuffer = append(uis.DPBuffer, pc)
	uis.Mutex.Unlock()
}

// func AddDataPointToUIS(dp *DataPoint){
// 	for i, v := range
// }

func (uis *UIServer) AddUI(TAG string, client *UI) {
	uis.Mutex.Lock()
	uis.ClientList[TAG] = client
	uis.Mutex.Unlock()
}

func (uis *UIServer) RemoveUI(TAG string) {
	uis.Mutex.Lock()
	// ui.ClientList[TAG] = nil
	delete(uis.ClientList, TAG)
	uis.Mutex.Unlock()
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

	_, configString, err := conn.ReadMessage()
	if err != nil {
		log.Println(err)
		return
	}

	var config Config
	err = json.Unmarshal(configString, &config)
	if err != nil {
		log.Println(err)
	}
	log.Println(config)
	for {
		// log.Println("reading ....")
		_, msg, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			return
		}
		// conn.WriteMessage(1, []byte(`{"meow":"meow"}`))
		// msh = strings.Trim(configData, "\n")

		log.Println("message from the UI ")
		log.Println(msg)
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

func (uis *UIServer) ParseDataPoints() {
	for {
		log.Println("sleeping 1 second befor parsing")
		time.Sleep(5 * time.Second)
		length := len(uis.DPBuffer)
		if length == 0 {
			continue
		}
		var newBuffer []ParsedCollection
		uis.Mutex.Lock()

		newBuffer = uis.DPBuffer[0:length]
		// log.Println("uis buffer before:", len(uis.DPBuffer))
		uis.DPBuffer = uis.DPBuffer[length:]
		// log.Println("uis buffer after:", len(uis.DPBuffer))
		// log.Println("new buffer:", len(newBuffer))
		// log.Println()
		uis.Mutex.Unlock()
		outgoing := &OutgoingData{}
		for _, v := range newBuffer {

			outgoing.DataPoints = append(outgoing.DataPoints, &v)
			// ParseDataPoint(v.Value[8:])
			// log.Println("ONE DATA POINT:")
			// for _, v := range parsedValues.Values {
			// 	// log.Print("SECTION::" + v.Tag)
			// 	log.Print(v.Tag+":", v.Index, ":", v.SubIndex, ":", v.Value)
			// }
		}

		// You can do XYZ and Alarms at the same time with channels.
		// Parse Alarms

		// User specific
		// Normalize
		// parse groups

	}

}
func (uis *UIServer) ParseForIndividualUsers(og *OutgoingData) {

	for _, client := range uis.ClientList {
		// // inside the client.
		// grid := &Grid{
		// 	Point: make(map[int]map[int]map[int]*ParsedDataPoint),
		// }
		for _, outgoing := range og.DataPoints {
			for _, dp := range outgoing.DPS {
				// z := 0
				// y := 0
				// x := 0
				index := strconv.Itoa(dp.Index) + "." + strconv.Itoa(dp.SubIndex)
				if index == client.Config.X.Index {
					// x = math.
				}
			}
		}
	}
}

type Grid struct {
	Point map[int]map[int]map[int]*ParsedDataPoint
}
type ParsedDataPointValues struct {
	Values []*ParsedDataPoint
}
type ParsedCollection struct {
	Tag string
	DPS []*ParsedDataPoint
}
type ParsedDataPoint struct {
	SubIndex int
	Index    int
	Tag      string
	Value    int64
	Group    string
}
type OutgoingData struct {
	Grid       *Grid
	Groups     []*Group
	DataPoints []*ParsedCollection
	Alarms     []*ParsedCollection
}
type Alarm struct {
	Collection string
	Name       string
	Triggered  time.Time
}

type Group struct {
	X int
	Y int
	Z int
	// Size int
	ID int
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
	WantsUpdates bool
	Indexes      []int
}

type Blink struct {
	Index string
	Tag   string
	// Normalize bool
}
type Luminocity struct {
	Index string
	Tag   string
	// Normalize bool
}

type Size struct {
	Index string
	Tag   string
	// Normalize bool
}

type X struct {
	Index string
	Tag   string
	// Normalize bool
}
type Y struct {
	Index string
	Tag   string
	// Normalize bool
}
type Z struct {
	Index string
	Tag   string
	// Normalize bool
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
