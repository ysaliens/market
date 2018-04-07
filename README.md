# CryptoMarket
CryptoMarket is a virtual cryptocurrency exchange supporting BTC, LTC, DOGE, and XMR written in Golang.

## About 
CryptoMarket uses Google's OAuth2 API for sign-in. Users sign in using their Google credentials. CryptoMarket only gets access to basic user profile info from Google such as name and email and does not store these values. After a new user registration, the app will credit the user $10000.
Once logged in, the user can check their balance as well as buy and sell cryptocurrency.


## Setup
* Install [Go](https://golang.org/)
* Install [MongoDB](https://www.mongodb.com/download-center?jmp=homepage#community)
* Add Go to PATH
* Add $GOPATH ("C:\Users\\$USER\go") to PATH
* Add MongoDB to PATH
* `go get github.com/ysaliens/market` to get project files and all dependencies
* `mkdir $GOPATH/src/github.com/ysaliens/market/database/db` <--- This is where database files will be stored
* `cd $GOPATH/src/github.com/ysaliens/market/`

## Build
`go build` from root of project.

If missing dependencies (or used git clone), run `go get -v ./...` to get all dependencies

## Run 
* `mongod --dbpath "$GOPATH/src/github.com/ysaliens/market/database/db` to start database
* `./market.exe` to start server (separate window)
* Navigate to `http://localhost:8080`

## Architecture
CryptoMarket uses MongoDB for database. It is written in Golang utilizing the high-performance Gin-Gonic framework. Google's OAuth2 is used for authentication with full support for sessions. This avoids having to encrypt, salt, or store passwords and is a welcome change for users.

When the server starts, it breaks into two main routines - server routine and exchanges routine. By default, every 15 seconds the exchanges routine will spawn four (BTC, LTC, DOGE, XMR) routines that will query the exchanges for the price of each ticker. Two exchanges (Bittrex and Poloniex) will be volume weighted to create the price for alt coins (LTC, DOGE, XMR). For Bitcoin (BTC), only Coinbase is currently used. 

The server routine (and any subroutines it spawns for users) are passed the address of the struct holding current prices for all coins. This avoids each user session trying to query prices of cryptocurrencies and hence keeps load on the server low. A Read/Write mutex is used to synchronize access to the struct for performance.

## TO-DOs
* __Re-write exchanges update code.__ Current code requests updates per ticker from exchanges - spawning routines for every ticker. While this works great for a proof of concept such as this application (low number of tickers and a slow refresh rate (15 sec)), it is suboptimal. A more scalable approach would be to spawn routines per every exchange supported. Each of those routines would get all the tickers supported from the exchange and map them. An update function would then look up data from the maps to update the price struct. The exchanges.go file would become it's own package and be broken down into files based on exchange. I did not have time to do this.
* __Automated unit testing__ As I re-wrote things a lot (testing by hand), I did not have time to write automated unit tests. Golang has very good support for testing and given more time I would add that.
* __Better Logout__ The current logout button is a redirect to the login page allowing another login. To logout, one needs to log out of the Google account they used before logging as a different user in CryptoMarket.
* __Human Friendly Floats__ I would like to use the [Decimal Package](https://github.com/shopspring/decimal) for a better representation of floats on the UI (and more accurate compute)
* __UI Improvements__ The UI could use a lot of love. Since Golang cannot run client-side code, adding some Javascript and a new design would really help CryptoMarket.

