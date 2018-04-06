package main

import (
	"time"
	"github.com/ysaliens/market/handlers"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
)

func main() {
	address := "localhost:8080"

	// Initialize a subroutine to update ticker prices
	price := handlers.CreateTickers()
	go func(){
		for {
			handlers.UpdateTickers(price)
			//Change to change ticker refresh rate
			time.Sleep(15 * time.Second)
		}
	}()

	router := gin.Default()
	store  := sessions.NewCookieStore([]byte(handlers.RToken(64)))
	store.Options(sessions.Options{
		Path:   "/",
		MaxAge: 86400 * 7,
	})
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(sessions.Sessions("market", store))
	router.Static("/css", "./static/css")
	router.Static("/img", "./static/img")
	router.LoadHTMLGlob("templates/*")

	router.GET("/", handlers.LoginHandler)
	router.GET("/login", handlers.LoginHandler)
	router.GET("/auth", handlers.AuthHandler(address))

	authorized := router.Group("/user")
	authorized.Use(handlers.Authorize(address))
	{
		authorized.GET("/account", handlers.AccountHandler(price,address))
		authorized.POST("/account", handlers.TradeHandler(price,address))
	}

	router.Run(address)
}
