package controller

import (
	"log"
	"strconv"
	"time"

	"github.com/zkynetio/lynx/ui"
)

func SaveToUIBuffer(parse chan DPCollection, send chan []byte) {
	History := make(map[string]DPCollection)
	for {
		dpc := <-parse

		olddpc, ok := History[dpc.Tag]
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
				send <- []byte(msg)
			} else {
				log.Println("HAS NOT CHANGED !!")
			}
		}

		History[dpc.Tag] = dpc
	}
}

func ShipToUIS(send chan []byte) {
	for {
		time.Sleep(500 * time.Millisecond)
		dpcLength := len(send)
		if dpcLength == 0 {
			continue
		}
		var data []byte
		for i := 0; i < dpcLength; i++ {
			msg := <-send
			data = append(data, byte(44))
			data = append(data, msg...)
		}

		if len(data) < 1 {
			continue
		}
		for _, v := range ui.Server.ClientList {
			v.Conn.WriteMessage(1, data)
		}

	}
}
