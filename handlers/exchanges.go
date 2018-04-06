package handlers

import (
	"fmt"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
)

// Coinbase JSON object is compound
type Coinbase struct {
	Data CoinbaseData
}

// Amount is spot price
type CoinbaseData struct {
	Amount float64 `json:",string"`
	Currency string
}

// Bittrex JSON is also compound
type Bittrex struct {
	Success bool
	Message string
	Data BittrexData `json:"result"`
}

// Use last price as spot price
type BittrexData struct {
	Bid  float64
	Ask  float64
	Amount float64 `json:"last"`
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

// Function to update the price of a single ticker
// Ran concurrently for each ticker
// Last price is kept if an error occurs updating it
func getTicker(price *Cryptos, ticker string){
	var url string
	var b Coinbase
	var alt Bittrex

	// Get API URL
	if ticker == "BTC" {
		url = "https://api.coinbase.com/v2/prices/spot?currency=USD"
	} else {
		url = "https://bittrex.com/api/v1.1/public/getticker?market="+ticker
	}

	// Get price as a JSON object
	response, err := getContent(url)
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}

	// Decode
	if ticker == "BTC" {
		if err = json.Unmarshal(response, &b); err !=nil {
			fmt.Printf("Failed updating price: %v", err)
			return
		}
		fmt.Printf("%v: |%v|\n",ticker, b.Data.Amount) 
	} else {
		if err = json.Unmarshal(response, &alt); err !=nil {
			fmt.Printf("Failed updating price: %v", err)
			return
		}
		fmt.Printf("%v: |%v|\n",ticker, alt.Data.Amount) 
	}

	// Update price struct, lock rw mutex for writing
	// This will wait until our readers are done before updating
	price.Lock.Lock()
	switch ticker {
		case "BTC":
				price.BTC     = b.Data.Amount
		case "BTC-LTC":
				price.BTCLTC  = alt.Data.Amount
		case "BTC-DOGE":
				price.BTCDOGE = alt.Data.Amount
		case "BTC-XMR":
				price.BTCXMR  = alt.Data.Amount
	}
	price.Lock.Unlock()
}

// Update ticker prices in parallel
// Updates are synchronized with a RW mutex
// Currently written to easily add more tickers
// If we add more exchanges, re-write this based on exchanges
func UpdateTickers(price *Cryptos){
	currencies := [4]string{"BTC","BTC-LTC","BTC-DOGE","BTC-XMR"}
	for _ , currency := range currencies {
		go getTicker(price, currency)
	}
	fmt.Printf("BTC: %v | LTC: %v | DOGE: %v | XMR: %v\n\n",price.BTC,price.BTCLTC,price.BTCDOGE,price.BTCXMR)
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

