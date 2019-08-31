package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/zkynetio/safelocker"
)

type UIServer struct {
	ClientList     map[string]*UI
	Mutex          sync.RWMutex
	IP             string
	Port           string
	DPChan         chan []byte
	HistoryChannel chan DPCollection
	History        map[string]DPCollection
}
type UI struct {
	Conn        *websocket.Conn
	Config      Config
	DataChannel chan []byte
	Buffer      []*OutgoingData
	safelocker.SafeLocker
}

var UIHTTPsrv *http.Server

func (uis *UIServer) SaveToUIBuffer() {
	uis.History = make(map[string]DPCollection)
	for {
		dpc := <-uis.HistoryChannel

		olddpc, ok := uis.History[dpc.Tag]
		if ok {
			hasChanged := false
			msg := dpc.Tag + "/"
			for _, v := range dpc.DPS {
				for _, iv := range olddpc.DPS {
					if iv.Index == v.Index {
						if iv.Value != v.Value {
							hasChanged = true
							msg = msg + strconv.Itoa(v.Index) + "/" + strconv.Itoa(v.Value) + "/"
						}
					}
				}
			}

			if hasChanged {
				// log.Println("sending on the dp chan!")
				uis.DPChan <- []byte(msg)
			} else {
				log.Println("HAS NOT CHANGED !!")
			}
		}

		uis.History[dpc.Tag] = dpc
	}
}

func (uis *UIServer) ShipToUIS() {
	for {
		time.Sleep(1000 * time.Millisecond)
		dpcLength := len(uis.DPChan)
		var data []byte
		// log.Println("PD chan length", dpcLength)
		for i := 0; i < dpcLength; i++ {
			msg := <-uis.DPChan
			data = append(data, byte(44))
			data = append(data, msg...)
		}

		if len(data) < 1 {
			continue
		}
		for _, v := range uis.ClientList {
			v.Conn.WriteMessage(1, data)
		}

	}
}

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

	router := mux.NewRouter()
	router.HandleFunc("/", uis.wsHandler).Methods("GET")

	srv := http.Server{
		Addr:    uis.IP + ":" + uis.Port,
		Handler: router,
	}

	UIHTTPsrv = &srv
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
		panic("meow")
	}

}

func (uis *UIServer) wsHandler(w http.ResponseWriter, r *http.Request) {

	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}
	go uis.AcceptConnection(conn)

}

func (uis *UIServer) AcceptConnection(conn *websocket.Conn) {
	UID := conn.RemoteAddr().String()
	uis.AddUI(UID, &UI{
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
	uis.ClientList[UID].Config = config
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

}

type Grid struct {
	Point map[int]map[int]map[int]*ParsedDataPoint
}
type ParsedDataPointValues struct {
	Values []*ParsedDataPoint
}
type ParsedCollection struct {
	Tag       string
	DPS       []*ParsedDataPoint
	BasePoint bool
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
	DataPoints []ParsedCollection
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
	Tag int
}

type Config struct {
	Indexes []int
	// Rate in milliseconds
	UpdateRate   int
	WantsUpdates bool
}
