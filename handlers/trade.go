package handlers

import (
	"log"
	"net/http"
	"github.com/ysaliens/market/database"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Trade Form
type TradeForm struct {
	Trade    string `form:"trade" binding:"required"`
	Currency string `form:"currency" binding:"required"`
	Amount   float64 `form:"amount" binding:"required"`
}

// Trade Handler
func TradeHandler(price *Cryptos, address string) gin.HandlerFunc {
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
			// This shouldn't happen but if we can't load the user after a login...
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "User not found. Please login to create a user.", "link": link})
			return
		}

		// Get current ticker prices for trade and release object for updates (Read lock)
		price.Lock.RLock()
		log.Printf("BTC: %v | LTC: %v | DOGE: %v | XMR: %v\n",price.BTC,price.BTCLTC,price.BTCDOGE,price.BTCXMR)
		BTC          := price.BTC
		BTCLTC       := price.BTCLTC
		BTCDOGE      := price.BTCDOGE
		BTCXMR       := price.BTCXMR
		price.Lock.RUnlock()

		// Calculate total account balance
		total := uFound.USD + BTC*(uFound.BTC + uFound.LTC*BTCLTC + uFound.DOGE*BTCDOGE + uFound.XMR*BTCXMR)

		// Get user input for trade & ensure it is complete
		var form TradeForm
		link = link + "/user/account"
		if c.ShouldBind(&form) != nil || form.Amount < 0 {
			log.Printf("Bad User Input Form binding")
			c.HTML(http.StatusOK, "account.tmpl", gin.H{"email": uFound.Email, 
			"USD": uFound.USD, "BTC": uFound.BTC, "LTC" : uFound.LTC,
			"DOGE": uFound.DOGE, "XMR": uFound.XMR,
			"marketBTC": BTC, "marketLTC": BTCLTC,
			"marketDOGE": BTCDOGE, "marketXMR": BTCXMR, 
			"err": "Please enter a valid amount.", "total": total, "link": link})
			return
		}

		// Perform trade
		// Assuming perfect, instant execution which doesn't happen.
		// Proof of concept. Real market orders have % buffers
		userErr := "" // Message to send back to user if amounts error
		log.Printf("ORDER: |%v| |%v| |%v|",form.Trade,form.Currency,form.Amount)
		switch form.Trade {
		case "Market Buy":
			switch form.Currency {
			case "BTC":
				if (uFound.USD >= BTC * form.Amount){
					uFound.BTC += form.Amount
					uFound.USD -= (BTC * form.Amount)
				} else {
					userErr = "You do not have enough USD."
				}
			case "LTC":
				if (uFound.BTC >= BTCLTC * form.Amount){
					uFound.LTC += form.Amount
					uFound.BTC -= (BTCLTC * form.Amount)
				} else {
					userErr = "You do not have enough BTC to make this transaction."
				}
			case "DOGE":
				if (uFound.BTC >= BTCDOGE * form.Amount){
					uFound.DOGE += form.Amount
					uFound.BTC -= (BTCDOGE * form.Amount)
				} else {
					userErr = "You do not have enough BTC to make this transaction."
				}
			case "XMR":
				if (uFound.BTC >= BTCXMR * form.Amount){
					uFound.XMR += form.Amount
					uFound.BTC -= (BTCXMR * form.Amount)
				} else {
					userErr = "You do not have enough BTC to make this transaction."
				}
			}
		case "Market Sell":
			log.Printf("ORDER: |%v| |%v| |%v|",form.Trade,form.Currency,form.Amount)
			switch form.Currency {
			case "BTC":
				if (uFound.BTC >= form.Amount){
					uFound.BTC -= form.Amount
					uFound.USD += (BTC * form.Amount)
				} else {
					userErr = "You do not have enough BTC to make this transaction."
				}
			case "LTC":
				if (uFound.LTC >= form.Amount){
					uFound.LTC -= form.Amount
					uFound.BTC += (BTCLTC * form.Amount)
				} else {
					userErr = "You do not have enough LTC to make this transaction."
				}
			case "DOGE":
				if (uFound.DOGE >= form.Amount){
					uFound.DOGE -= form.Amount
					uFound.BTC += (BTCDOGE * form.Amount)
				} else {
					userErr = "You do not have enough DOGE to make this transaction."
				}
			case "XMR":
				if (uFound.XMR >= form.Amount){
					uFound.XMR -= form.Amount
					uFound.BTC += (BTCXMR * form.Amount)
				} else {
					userErr = "You do not have enough XMR to make this transaction."
				}
			}
		}

		// Calculate new total account balance
		total = uFound.USD + BTC*(uFound.BTC + uFound.LTC*BTCLTC + uFound.DOGE*BTCDOGE + uFound.XMR*BTCXMR)

		// Update user in database
		err = db.UpdateUser(&uFound)
		if err != nil {
			log.Println(err)
			return
		}

		// Send response
		c.HTML(http.StatusOK, "account.tmpl", gin.H{"email": uFound.Email, 
			"USD": uFound.USD, "BTC": uFound.BTC, "LTC" : uFound.LTC,
			"DOGE": uFound.DOGE, "XMR": uFound.XMR,
			"marketBTC": BTC, "marketLTC": BTCLTC,
			"marketDOGE": BTCDOGE, "marketXMR": BTCXMR, 
			"err": userErr, "total": total, "link": link})
	}
}
