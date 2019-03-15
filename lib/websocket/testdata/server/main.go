// Testing websocket server using autobahn.io
package main

import (
	"log"
	"os"

	"github.com/shuLhan/share/lib/debug"
	"github.com/shuLhan/share/lib/websocket"
)

//
// handleBin from websocket by echo-ing back the payload.
//
func handleBin(conn int, payload []byte) {
	packet := websocket.NewFrameBin(false, payload)

	err := websocket.Send(conn, packet)
	if err != nil {
		log.Println("handleBin: " + err.Error())
	}
}

//
// handleText from websocket by echo-ing back the payload.
//
func handleText(conn int, payload []byte) {
	packet := websocket.NewFrameText(false, payload)

	if debug.Value >= 3 {
		log.Printf("testdata/server: handleText: {payload.len:%d}\n", len(payload))
	}

	err := websocket.Send(conn, packet)
	if err != nil {
		log.Println("handleText: " + err.Error())
	}
}

func main() {
	srv, err := websocket.NewServer(9001)
	if err != nil {
		log.Println("internal/server: " + err.Error())
		os.Exit(2)
	}

	srv.HandleBin = handleBin
	srv.HandleText = handleText

	srv.Start()
}