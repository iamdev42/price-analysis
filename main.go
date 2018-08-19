package main

import (
	"log"
	"net/http"
	"strings"
	"encoding/json"
	"fmt"
	"os"
	"io/ioutil"
	"time"
	"strconv"
)

const returnTicker string = "https://api.idex.market/returnTicker"

type tickerList struct {
	Symbols []symbol `json:"symbols"`
}

type symbol struct {
	Last string `json:"last"`
	PercentChange string `json:"percentChange"`
	BaseVolume string `json:"baseVolume"`
	QuoteVolume string `json:"quoteVolume"`
}

func getExchangeInfo() map[string]symbol {
	response, err := http.Get(returnTicker)
	if err != nil {
		panic(err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		var newSymbolList map[string]symbol
		json.Unmarshal(data, &newSymbolList)
		return newSymbolList
	}
}

var previousSymbolList map[string]symbol
const dipThreashold float64 = 20.0

func main() {

	//sample := `{"ETH_DX":{"last":"80","high":"0.00000129","low":"0.00000112","lowestAsk":"0.000001219999982558","highestBid":"0.00000112","percentChange":"-12.5","baseVolume":"8.626480612675553493","quoteVolume":"7249641.978793682040393169"},
	//			"ETH_AUC":{"last":"0.00023","high":"0.00023","low":"0.000200740231454066","lowestAsk":"0.000221978867611803","highestBid":"0.000211002014382012","percentChange":"6.47655201","baseVolume":"7.783199997744259321","quoteVolume":"36661.98265640783736685"}}`
	//sample2 := `{"ETH_DX":{"last":"100","high":"0.00000129","low":"0.00000112","lowestAsk":"0.000001219999982558","highestBid":"0.00000112","percentChange":"-12.5","baseVolume":"8.626480612675553493","quoteVolume":"7249641.978793682040393169"},
	//			"ETH_AUC":{"last":"0.00023","high":"0.00023","low":"0.000200740231454066","lowestAsk":"0.000221978867611803","highestBid":"0.000211002014382012","percentChange":"6.47655201","baseVolume":"7.783199997744259321","quoteVolume":"36661.98265640783736685"}}`
	//
	//byt := []byte(sample)
	//byt2 := []byte(sample2)
	//var newSymbolList map[string]symbol
	//
	//json.Unmarshal(byt2, &previousSymbolList)
	//
	//parseErr := json.Unmarshal(byt, &newSymbolList)
	//if parseErr != nil {
	//	panic(parseErr)
	//}
	log.Printf("Symbol Dip Volume PricePrevious PriceLast\n")
	ticker := time.NewTicker(50 * time.Second)
	go func() {
		for _ = range ticker.C {
			checkDips()
		}
	}()

	// Set new to previous one


	http.HandleFunc("/", sayHello)
	//fmt.Println("Starting Server")
	err := http.ListenAndServe(":8087", nil)
	if err != nil {
		fmt.Printf("HTTP failed: %s\n", err.Error())
		os.Exit(1)
	}
}

func checkDips() {
	var newSymbolList = getExchangeInfo()
	//fmt.Printf("%s", newSymbolList)
	// If there exists something to compare with
	if previousSymbolList != nil {
		// Got through all symbols and compare with previous values to check for dips
		for symbol, _ := range newSymbolList {
			previousPrice, e1 := strconv.ParseFloat(previousSymbolList[symbol].Last, 64)
			if e1 != nil {
				continue
			}
			newPrice, e2 := strconv.ParseFloat(newSymbolList[symbol].Last, 64)
			if e2 != nil {
				continue
			}

			//fmt.Printf("symbol = %s | %.2f <=> %.2f\n", symbol, newPrice, previousPrice)

			if newPrice < previousPrice {

				var percents =  (previousPrice - newPrice) / previousPrice * 100
				// If there is a large dip, we put this out to log
				if percents > dipThreashold {
					f, _ := strconv.ParseFloat(newSymbolList[symbol].BaseVolume, 64)
					log.Printf("%s %.2f %.2f %.10f %.10f\n", symbol, percents,f , previousPrice, newPrice)
				}
				//else{
				//	log.Printf("N - %s = %.2f",symbol, percents)
				//}
			}
		}
	}
	//log.Println("Nothing found this time")
	previousSymbolList = newSymbolList
}


//func checkSocketMessage(msgStream []byte) {
//
//	var message socketMsg
//	err := json.Unmarshal(msgStream, &message)
//	if err != nil {
//		log.Print(err)
//	}
//	//log.Print(message)
//
//	// Map message with mapper
//	obj := message.Info.(map[string]interface{})
//
//	// Get kline open time
//	openTime := int(obj["t"].(float64))
//
//	// Get close price
//	//closePrice, err := strconv.ParseFloat(obj["c"].(string), 64)
//	//if err != nil {
//	//	panic(err)
//	//}
//
//	// Get low price
//	lowPrice, err := strconv.ParseFloat(obj["l"].(string), 64)
//	if err != nil {
//		panic(err)
//	}
//	symbol := obj["s"].(string)
//	mutex.Lock()
//	existingKline := klineMap[symbol]
//
//	// New time frame kline
//	if existingKline.OpenTime < openTime {
//		// Get open price, to calculate actual dip
//		openPrice, err := strconv.ParseFloat(obj["o"].(string), 64)
//		if err != nil {
//			panic(err)
//		}
//
//		// Calculate dip in percent
//		dipPercent := (openPrice - existingKline.LowPrice) / openPrice * 100
//		//log.Print("Dip percent: ", dipPercent)
//		if dipPercent >= dipThreashold {
//			log.Print("FOUND DIP: ", dipPercent)
//			// Write this record into firebase
//			log.Printf("OnMessage: %s\n", msgStream)
//		}
//
//		klineMap[symbol] = klineInfo{openTime, lowPrice}
//
//	} else {
//		// We are working on same time kline
//		klineMap[symbol] = klineInfo{openTime, lowPrice,}
//	}
//	mutex.Unlock()
//}

func sayHello(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/")
	message = "Hello " + message
	w.Write([]byte(message))
}

