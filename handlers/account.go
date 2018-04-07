package handlers

import (
	"log"
	"net/http"
	"github.com/ysaliens/market/database"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Account Balance Handler
func AccountHandler(price *Cryptos, address string) gin.HandlerFunc {
	return func(c *gin.Context){
		// Redirect link for errors
		link := "http://" + address

		session := sessions.Default(c)
		userID := session.Get("user-id")

		// Load user from db
		db := database.MongoDBConnection{}
		uFound, err := db.LoadUser(userID.(string))
		if err != nil {
			log.Println(err)
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "User not found. Please login to create a user.", "link": link})
			return
		}

		// Get current ticker prices for trade and release object for updates (Read lock)
		price.Lock.RLock()
		BTC     := price.BTC
		BTCLTC  := price.BTCLTC
		BTCDOGE := price.BTCDOGE
		BTCXMR  := price.BTCXMR
		price.Lock.RUnlock()

		// Calculate total account balance
		total := uFound.USD + BTC*(uFound.BTC + uFound.LTC*BTCLTC + uFound.DOGE*BTCDOGE + uFound.XMR*BTCXMR)

		link = link + "/user/account"

		// Send response
		c.HTML(http.StatusOK, "account.tmpl", gin.H{"email": uFound.Email, 
			"USD": uFound.USD, "BTC": uFound.BTC, "LTC" : uFound.LTC,
			"DOGE": uFound.DOGE, "XMR": uFound.XMR, "total": total,
			"marketBTC": BTC, "marketLTC": BTCLTC,
			"marketDOGE": BTCDOGE, "marketXMR": BTCXMR, "link": link})
	}
}
