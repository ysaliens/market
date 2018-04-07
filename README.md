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
