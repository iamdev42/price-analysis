package main

import (
	"log"
	"net/http"
	"strings"
	"fmt"
	"os"
	"github.com/rgamba/evtwebsocket"
)

// const address string = "wss://stream.binance.com:9443/ws/bnbbtc@trade"
const address string = "wss://stream.binance.com:9443/ws/wavesbtc@trade"

//type trade struct{
//	OpenTime int
//	CloseTime int
//	ClosePrice string
//	LowPrice string
//	Dip float64
//}

func main() {
	log.Println("Welcome")


	connection := evtwebsocket.Conn{

		// When connection is established
		OnConnected: func(w *evtwebsocket.Conn) {
			log.Println("Connected")
		},

		// When a message arrives
		OnMessage: func(msg []byte, w *evtwebsocket.Conn) {
			log.Printf("OnMessage: %s\n", msg)
			//checkSocketMessage(msg)
		},

		// When the client disconnects for any reason
		OnError: func(err error) {
			log.Printf("** ERROR **\n%s\n", err.Error())
		},

		// This is used to match the request and response messagesP>termina
		MatchMsg: func(req, resp []byte) bool {
			return string(req) == string(resp)
		},

		// Auto reconnect on error
		Reconnect: true,

		// Set the ping interval (optional)
		// PingIntervalSecs: 50,

		// Set the ping message (optional)
		PingMsg: []byte("PING"),
	}

	// Connect
	if err := connection.Dial(address, ""); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", sayHello)
	fmt.Println("Starting Server")
	err := http.ListenAndServe(":8087", nil)
	if err != nil {
		fmt.Printf("HTTP failed: %s\n", err.Error())
		os.Exit(1)
	}
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")
	message = "Hello " + message
	w.Write([]byte(message))
}

