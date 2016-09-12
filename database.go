package main

import (
	"errors"
	"fmt"

	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

var (
	MirrorLockedAlready = errors.New("MirrorLockedAlready")
)

type Database interface {
	Connect() error
	IsConnected() (bool, error)

	LockMirror(mirror string) error
	UnlockMirror(mirror string) error

	GetServerMirrorState(server, mirror string) (string, error)
	SetServerMirrorState(mirror, state string, server ...string) error

	SetMirrorRevision(mirror, revision string)

	GetSuccessServers(mirror string) ([]string, error)
}

type DatabaseMongo struct {
	session  *mgo.Session
	database *mgo.Database
	dsn      string
}

func NewDatabaseMongo(dsn string) Database {
	return &DatabaseMongo{
		dsn: dsn,
	}
}

func (mongo *DatabaseMongo) Connect() error {
	var err error

	mongo.session, err = mgo.Dial(mongo.dsn)
	if err != nil {
		return err
	}

	mongo.database = mongo.session.DB("")
	mongo.nodes = mongo.database.C("nodes")

	return nil
}

func (mongo *DatabaseMongo) IsConnected() (bool, error) {
	err := mongo.session.Ping()
	if err != nil {
		return false, err
	}

	return true, nil
}

func (mongo *DatabaseMongo) LockMirror(mirror string) error {
	data := map[string]interface{}{}

	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"count": 1}},
		Upsert:    true,
		ReturnNew: false,
	}

	_, err := deployerd.DB.Locks.Find(
		bson.M{"domain": domain},
	).Apply(change, &data)
	if err != nil {
		return fmt.Errorf("can't upsert lock document: %s", err)
	}

	if _, ok := data["count"]; ok {
		return fmt.Errorf("environment already locked")
	}

	return nil
}
