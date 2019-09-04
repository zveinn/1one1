package ui

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/zkynetio/safelocker"
)

var UIHTTPsrv *http.Server
var Server *UIServer

type UIServer struct {
	ClientList map[string]*UI
	safelocker.SafeLocker
	IP         string
	Port       string
	HasClients bool
}
type UI struct {
	Conn   *websocket.Conn
	Config Config
	safelocker.SafeLocker
}

type Config struct {
	Indexes []int
	// Rate in milliseconds
	UpdateRate   int
	WantsUpdates bool
}

func NewUIServer() *UIServer {
	Server = &UIServer{}
	return Server
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
	log.Println("Listening for UI clients at: ", uis.IP+":"+uis.Port)
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

func (uis *UIServer) AddUI(TAG string, client *UI) {
	uis.Lock()
	uis.ClientList[TAG] = client
	uis.HasClients = true
	uis.Unlock()
}

func (uis *UIServer) RemoveUI(TAG string) {
	uis.Lock()
	delete(uis.ClientList, TAG)
	if len(uis.ClientList) == 0 {
		uis.HasClients = false
	}
	uis.Unlock()
}
