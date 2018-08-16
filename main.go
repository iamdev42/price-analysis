package main

import (
	"log"
	"net/http"
	"strings"
	"encoding/json"
	"strconv"
	"fmt"
	"io/ioutil"
	"os"
	"math"
	"github.com/rgamba/evtwebsocket"
	"sync"
)

// const address string = "wss://stream.binance.com:9443/ws/bnbbtc@trade"
const addressPlaceholder string = "wss://stream.binance.com:9443/ws/%s@kline_1m"
const symbolList string = "https://api.binance.com/api/v1/exchangeInfo"

type socketMsg struct {
	Info interface{} `json:"k"`
}

type exchangeInfo struct {
	Symbols []symbol `json:"symbols"`
}

type symbol struct {
	Symbol string `json:"symbol"`
	QuoteAsset string `json:"quoteAsset"`
}

func getExchangeInfo() exchangeInfo {
	response, err := http.Get(symbolList)
	if err != nil {
		panic(err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
		exchInfo := exchangeInfo{}
		err := json.Unmarshal(data, &exchInfo)
		if err != nil {
			panic(err)
		}
		log.Print(exchInfo)
		return exchInfo
	}
}

var klineMap = map[string]klineInfo{}
var mutex sync.Mutex

func main() {
	log.Println("Welcome")

	// Get all trading symbols
	exchangeInfoResponse := getExchangeInfo()
	for _, element := range exchangeInfoResponse.Symbols {
		if element.QuoteAsset == "BTC" {
			socketAddress := fmt.Sprintf(addressPlaceholder, strings.ToLower(element.Symbol))
			log.Print(socketAddress)

			mutex.Lock()
			klineMap[element.Symbol] = klineInfo{math.MaxInt64,0}
			mutex.Unlock()

			go func(socketAddress string) {
				connection := evtwebsocket.Conn{

					// When connection is established
					OnConnected: func(w *evtwebsocket.Conn) {
						log.Println("Connected")
					},

					// When a message arrives
					OnMessage: func(msg []byte, w *evtwebsocket.Conn) {
						//log.Printf("OnMessage: %s\n", msg)
						checkSocketMessage(msg)
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
				if err := connection.Dial(socketAddress, ""); err != nil {
					log.Fatal(err)
				}
			}(socketAddress)
			//if index > 100 {
			//	break
			//}
		}
	}

	//log.Print(klineMap)
	//log.Print(klineMap["BNBBTC"])

	// For testing purposes
	//checkSocketMessage([]byte(jsonStream))



	http.HandleFunc("/", sayHello)
	fmt.Println("Starting Server")
	err := http.ListenAndServe(":8087", nil)
	if err != nil {
		fmt.Printf("HTTP failed: %s\n", err.Error())
		os.Exit(1)
	}
}

type klineInfo struct{
	OpenTime int
	LowPrice float64
}

const dipThreashold float64 = 5.0
//var existingKline klineInfo = klineInfo{math.MaxInt64,0}

func checkSocketMessage(msgStream []byte) {

	var message socketMsg
	err := json.Unmarshal(msgStream, &message)
	if err != nil {
		log.Print(err)
	}
	//log.Print(message)

	// Map message with mapper
	obj := message.Info.(map[string]interface{})

	// Get kline open time
	openTime := int(obj["t"].(float64))

	// Get close price
	//closePrice, err := strconv.ParseFloat(obj["c"].(string), 64)
	//if err != nil {
	//	panic(err)
	//}

	// Get low price
	lowPrice, err := strconv.ParseFloat(obj["l"].(string), 64)
	if err != nil {
		panic(err)
	}
	symbol := obj["s"].(string)
	mutex.Lock()
	existingKline := klineMap[symbol]

	// New time frame kline
	if existingKline.OpenTime < openTime {
		// Get open price, to calculate actual dip
		openPrice, err := strconv.ParseFloat(obj["o"].(string), 64)
		if err != nil {
			panic(err)
		}

		// Calculate dip in percent
		dipPercent := (openPrice - existingKline.LowPrice) / openPrice * 100
		//log.Print("Dip percent: ", dipPercent)
		if dipPercent >= dipThreashold {
			log.Print("FOUND DIP: ", dipPercent)
			// Write this record into firebase
			log.Printf("OnMessage: %s\n", msgStream)
		}

		klineMap[symbol] = klineInfo{openTime, lowPrice}

	} else {
		// We are working on same time kline
		klineMap[symbol] = klineInfo{openTime, lowPrice,}
	}
	mutex.Unlock()
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")
	message = "Hello " + message
	w.Write([]byte(message))
}

