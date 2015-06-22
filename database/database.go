package database

import (
	"os"

	"gopkg.in/mgo.v2"
)

var (
	GMyDb *MyDb
)

type MyDb struct {
	session  *mgo.Session
	database *mgo.Database
}

func NewMyDb() *MyDb {
	myDb := new(MyDb)

	var err error
	myDb.session, err = mgo.Dial(os.Getenv("MONGODB_ADDRESS"))
	if err != nil {
		panic(err)
	}

	// Optional. Switch the session to a monotonic behavior.
	myDb.session.SetMode(mgo.Monotonic, true)
	myDb.database = myDb.session.DB(os.Getenv("MONGODB_DATABASE"))

	return myDb
}

func (myDb *MyDb) Destroy() {
	myDb.session.Close()
}

func (myDb *MyDb) GetCollection(collectionName string) *mgo.Collection {
	return myDb.database.C(collectionName)
}

func Init() {
	GMyDb = NewMyDb()
}
