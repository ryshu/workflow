package mongo

import (
	"context"
	"sync"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Session struct {
	addr        string
	database    string
	flowColl    string
	historyColl string
	client      *mongo.Client
	initErr     error
	once        sync.Once
}

// New creates a new MongoDB session handler, and automatically
// attempts to connect to the server.
func New(database string, addr string, flowColl string, historyColl string) *Session {
	session := Session{
		addr:        addr,
		database:    database,
		flowColl:    flowColl,
		historyColl: historyColl,
	}
	go session.GetMongoClient()
	return &session
}

// GetMongoClient Return mongodb connection to work with
func (session *Session) GetMongoClient() (*mongo.Client, error) {
	//Perform connection creation operation only once.
	session.once.Do(func() {
		// Set client options
		clientOptions := options.Client().ApplyURI(session.addr)
		clientOptions.SetRegistry(mongoRegistry)

		// Connect to MongoDB
		client, err := mongo.Connect(context.TODO(), clientOptions)
		if err != nil {
			session.initErr = err
		}
		// Check the connection
		err = client.Ping(context.TODO(), nil)
		if err != nil {
			session.initErr = err
		}
		session.client = client
	})
	return session.client, session.initErr
}

func (session *Session) GetFlowColl() (*mongo.Collection, error) {
	client, err := session.GetMongoClient()
	if err != nil {
		return nil, err
	}
	collection := client.Database(session.database).Collection(session.flowColl)

	return collection, err
}

func (session *Session) GetHistoryColl() (*mongo.Collection, error) {
	client, err := session.GetMongoClient()
	if err != nil {
		return nil, err
	}
	collection := client.Database(session.database).Collection(session.historyColl)

	return collection, err
}
