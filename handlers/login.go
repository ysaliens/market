package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/ysaliens/market/database"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var cred Credentials
var conf *oauth2.Config

// Credentials which stores google ids.
type Credentials struct {
	Cid     string `json:"cid"`
	Csecret string `json:"csecret"`
}

// Get session login URL
func getLoginURL(state string) string {
	return conf.AuthCodeURL(state)
}

// Create a random token with @l length
func RToken(l int) string {
	btoken := make([]byte, l)
	rand.Read(btoken)
	return base64.StdEncoding.EncodeToString(btoken)
}

// Read Google id and secret from credentials.json
// Executed automatically on package import
// Developer needs to set OAuth2 credentials for their app
//
func init() {
	file, err := ioutil.ReadFile("./credentials.json")
	if err != nil {
		log.Printf("File error: %v\n", err)
		os.Exit(1)
	}
	json.Unmarshal(file, &cred)

	conf = &oauth2.Config{
		ClientID:     cred.Cid,
		ClientSecret: cred.Csecret,
		RedirectURL:  "http://localhost:8080/auth",
		Scopes: []string{
			// Select scope from https://developers.google.com/identity/protocols/googlescopes#google_sign-in
			"https://www.googleapis.com/auth/userinfo.email", 
		},
		Endpoint: google.Endpoint,
	}
}

// Authenticate user with Google OAuth2
// Creates user account if needed
func AuthHandler(address string) gin.HandlerFunc {
	return func(c *gin.Context){
		// Redirect link for errors
		link := "http://" + address

		// Verify session codes
		session := sessions.Default(c)
		sessionState := session.Get("state")
		urlState := c.Request.URL.Query().Get("state")
		if sessionState != urlState {
			log.Printf("Invalid session state: retrieved: %s; Param: %s", sessionState, urlState)
			c.HTML(http.StatusUnauthorized, "error.tmpl", gin.H{"message": "Invalid session state.", "link":link})
			return
		}

		// Get token
		code := c.Request.URL.Query().Get("code")
		tok, err := conf.Exchange(oauth2.NoContext, code)
		if err != nil {
			log.Println(err)
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Login failed. Please try again.", "link":link})
			return
		}

		// Get and decode user info
		client := conf.Client(oauth2.NoContext, tok)
		userinfo, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		defer userinfo.Body.Close()
		data, _ := ioutil.ReadAll(userinfo.Body)
		u := database.User{}
		if err = json.Unmarshal(data, &u); err != nil {
			log.Println(err)
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Error decoding input. Please try again.", "link":link})
			return
		}
		session.Set("user-id", u.Email)
		err = session.Save()
		if err != nil {
			log.Println(err)
			c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Error saving session. Please try again.", "link":link})
			return
		}

		// Check database for user
		seen := false
		db := database.MongoDBConnection{}
		uFound, e := db.LoadUser(u.Email)
		// User found
		if e == nil {
			seen = true
		// Create a new user with initial balance
		} else {
			u.USD = 10000
			u.LTC = 0
			u.DOGE = 0
			u.XMR = 0
			err = db.SaveUser(&u)
			if err != nil {
				log.Println(err)
				c.HTML(http.StatusBadRequest, "error.tmpl", gin.H{"message": "Error saving new user. Please try again.", "link":link})
				return
			}
			// Display balance correctly on first login
			uFound = u
		}

		account  := link + "/user/account"
		c.HTML(http.StatusOK, "portal.tmpl", gin.H{"email": uFound.Email,"seen": seen, "account": account})
	}
}

// Load login page, start session
// Actual login is done in AuthHandler
func LoginHandler(c *gin.Context) {
	state := RToken(32)
	session := sessions.Default(c)
	session.Set("state", state)
	session.Save()
	link := getLoginURL(state)
	c.HTML(http.StatusOK, "auth.tmpl", gin.H{"link": link})
}
