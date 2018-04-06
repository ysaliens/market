package database

import (
	"fmt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// User struct
type User struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Profile       string `json:"profile"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	USD          float64 `json:"USD"`
	BTC	         float64 `json:"BTC"`
    LTC          float64 `json:"LTC"`
    DOGE         float64 `json:"DOGE"`
    XMR          float64 `json:"XMR"`
	EmailVerified bool   `json:"email_verified"`
	Gender        string `json:"gender"`
}

// MongoDBConnection Encapsulates a connection to a database.
type MongoDBConnection struct {
	session *mgo.Session
}

// SaveUser register a user so we know that we saw that user already.
func (mdb MongoDBConnection) SaveUser(u *User) error {
	mdb.session = mdb.GetSession()
	defer mdb.session.Close()
	if _, err := mdb.LoadUser(u.Email); err == nil {
		return fmt.Errorf("User already exists!")
	}
	
	c := mdb.session.DB("market").C("users")
	err := c.Insert(u)
	return err
}

// Update user account
func (mdb MongoDBConnection) UpdateUser(u *User) error {
	mdb.session = mdb.GetSession()
	defer mdb.session.Close()

	c := mdb.session.DB("market").C("users")
	err := c.Update(bson.M{"email": u.Email}, u)
	return err
}

// LoadUser get data from a user.
func (mdb MongoDBConnection) LoadUser(Email string) (result User, err error) {
	mdb.session = mdb.GetSession()
	defer mdb.session.Close()
	c := mdb.session.DB("market").C("users")
	err = c.Find(bson.M{"email": Email}).One(&result)
	return result, err
}

// GetSession return a new session if there is no previous one.
// Remove hardcoded localhost if database is ever not local
func (mdb *MongoDBConnection) GetSession() *mgo.Session {
	if mdb.session != nil {
		return mdb.session.Copy()
	}
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)
	return session
}
