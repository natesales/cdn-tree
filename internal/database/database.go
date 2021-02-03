package database

import (
	"context"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// Types

// Database wraps a cdnv3 database
type Database struct {
	Db *mongo.Database
}

// Node stores a single edge node
type Node struct {
	Provider   string  `json:"provider"`
	Latitude   float32 `json:"latitude"`
	Longitude  float32 `json:"longitude"`
	Authorized bool    `json:"authorized"`
}

// Functions

// newContext returns a context with given duration
func newContext(duration time.Duration) context.Context {
	ctx, _ := context.WithTimeout(context.Background(), duration)
	return ctx
}

// New constructs a new database object
func New(url string) *Database {
	ctx := newContext(10 * time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	ctx = newContext(10 * time.Second)
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugln("Connected to database")

	// Return database pointer
	return &Database{client.Database("cdnv3db")}
}

// GetNode looks up a node by string ID
func (d Database) GetNode(id string) bson.M {
	nodeObjectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Debug(err)
		return nil // Invalid ObjectId
	}

	var node bson.M
	ctx := newContext(10 * time.Second)
	// Run DB query
	if err := d.Db.Collection("nodes").FindOne(ctx, bson.M{"_id": nodeObjectId}).Decode(&node); err != nil {
		if err.Error() == "mongo: no documents in result" {
			log.Debug(err)
			return nil // Node with given ID doesn't exist, exit
		} else { // Leaving this useless else case for future error handling
			log.Debug(err)
			return nil // Other error
		}
	}

	// Check if node is authorized
	if node["authorized"] != true { // != true is required here as we're comparing an empty interface
		log.Debugf("Node %s is not authorized\n", id)
		return nil
	}

	return node
}

// AddNode inserts a new node into the database
func (d Database) AddNode(provider string, latitude float32, longitude float32) {
	node := Node{provider, latitude, longitude, false}
	insertResult, err := d.Db.Collection("nodes").InsertOne(newContext(10*time.Second), node)
	if err != nil {
		log.Warn(err)
	}

	log.Println("Inserted new object", insertResult)
}
