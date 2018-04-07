package handlers

import (
	"log"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
)

// Coinbase JSON objects
type Coinbase struct {
	Data CoinbaseData
}
type CoinbaseData struct {
	Amount float64 `json:",string"`
	Currency string
}

// Bittrex JSON
type Bittrex struct {
	Success bool
	Message string
	Data []BittrexData `json:"result"`
}
type BittrexData struct {
	MarketName string  `json:"MarketName"`
	High	   float64 `json:"High"`
	Low		   float64 `json:"Low"`
	Volume     float64 `json:"Volume"`
	Last       float64 `json:"Last"`
	BaseVolume float64 `json:"BaseVolume"`
	TimeStamp  string  `json:"TimeStamp"`
	Bid        float64 `json:"Bid"`
	Ask        float64 `json:"Ask"`
	OpenBuyOrders int  `json:"OpenBuyOrders"`
	OpenSellOrders int `json:"OpenSellOrders"`
	PrevDay    float64 `json:"PrevDay"`
	Created    string  `json:"PrevDay"`
}

// Poloniex JSON
type Poloniex struct {
	Id            int     `json:"id"`
	Last          float64 `json:"last,string"`
	LowestAsk     float64 `json:"lowestAsk,string"`
	HighestBid    float64 `json:"highestBid,string"`
	PercentChange float64 `json:"percentChange,string"`
	BaseVolume    float64 `json:"baseVolume,string"`
	QuoteVolume   float64 `json:"quoteVolume,string"`
	IsFrozen      int     `json:"isFrozen,string"`
	High24Hr      float64 `json:"high24hr,string"`
	Low24Hr       float64 `json:"low24hr,string"`
}

// This holds all prices
// It uses a Read/Write mutex
// This allows multiple readers
// but only a single writer
type Cryptos struct {
	BTC     float64
	BTCLTC  float64
	BTCDOGE float64
	BTCXMR  float64
	Lock    *sync.RWMutex
}

// Makes an API request
func getContent(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// Function to update the price for a single ticker
// Ran concurrently for each ticker
// For alt-coins, price is calculated as a volume-weighted average price
// -----------------------------------------------------------------------------------------------------
// TO-DO: If more coins are tracked, re-write this based on exchanges and NOT tickers.
//        Currently, it is faster and simpler being ticker-based due to low number of tickers
//        and the slow poll rate. With a lot more coins added, it becomes optimal to get all market
//        data from every exchange at once, map it, and then consult the maps for prices of every ticker.
//        At that point, split into own package (exchanges) + break into files for every exchange.
// -----------------------------------------------------------------------------------------------------
func getTicker(price *Cryptos, ticker string){
	var c Coinbase
	var b Bittrex
	var p map[string]Poloniex
	var url,url2 string		
	var vPrice float64

	// Get API URL
	if ticker == "BTC" {
		url = "https://api.coinbase.com/v2/prices/spot?currency=USD"
	} else {
		url = "https://bittrex.com/api/v1.1/public/getmarketsummary?market="+"BTC-"+ticker
		url2 = "https://poloniex.com/public?command=returnTicker"
	}

	// Get JSON and decode
	if ticker == "BTC" {
		response, err := getContent(url)
		if err != nil {
			log.Printf("Error: %v", err)
			return
		}
		if err = json.Unmarshal(response, &c); err !=nil {
			log.Printf("Failed updating price: %v", err)
			return
		}
	} else {
		responseB, errB := getContent(url)
		responseP, errP := getContent(url2)
		if errB != nil || errP != nil {
			log.Printf("Error: %v\n%v\n", errB,errP)
			return
		}
		errB = json.Unmarshal(responseB, &b)
		errP = json.Unmarshal(responseP, &p)
		if errB !=nil || errP != nil {
			log.Printf("Failed updating price: %v\n%v\n", errB,errP)
			return
		}
		tickerP, _ := p["BTC_"+ticker]
		//log.Printf("Bittrex  %v Price: %v Volume: %v\n",ticker,b.Data[0].Last, b.Data[0].Volume )
		//log.Printf("Poloniex %v Price: %v Volume: %v\n",ticker,tickerP.Last, tickerP.QuoteVolume )

		// Volume Weighted Average Price
		vPrice = ((b.Data[0].Volume*b.Data[0].Last) + (tickerP.QuoteVolume*tickerP.Last)) / (b.Data[0].Volume + tickerP.QuoteVolume)
	}

	// Update price struct, lock rw mutex for writing
	// This will wait until our readers are done before updating
	price.Lock.Lock()
	switch ticker {
		case "BTC":
				price.BTC     = c.Data.Amount
		case "LTC":
				price.BTCLTC  = vPrice
		case "DOGE":
				price.BTCDOGE = vPrice
		case "XMR":
				price.BTCXMR  = vPrice
	}
	price.Lock.Unlock()
}

// Update ticker prices in parallel
// Updates are synchronized with a RW mutex
// Currently written to easily add more tickers
// If we add more exchanges, re-write this based on exchanges
func UpdateTickers(price *Cryptos){
	currencies := [4]string{"BTC","LTC","DOGE","XMR"}
	for _ , currency := range currencies {
		go getTicker(price, currency)
	}
	// Uncomment to see coin prices at every update tick
	//log.Printf("BTC: %v | LTC: %v | DOGE: %v | XMR: %v\n",price.BTC,price.BTCLTC,price.BTCDOGE,price.BTCXMR)
}

// Initialize prices struct
func CreateTickers() *Cryptos{
	price := &Cryptos{
		Lock:    &sync.RWMutex{},
		BTC:     0,
		BTCLTC:  0,
		BTCDOGE: 0,
		BTCXMR:  0,
	}
	return price
}
