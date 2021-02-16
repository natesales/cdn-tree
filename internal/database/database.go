package database

import (
	"context"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
)

// Database wraps a mongo.Database
type Database struct {
	Db *mongo.Database
}

// NewContext returns a context with given duration
func NewContext(duration time.Duration) context.Context {
	// TODO: Remove this function and defer the cancel function
	ctx, _ := context.WithTimeout(context.Background(), duration)
	return ctx
}

// New constructs a new database object
func New() *Database {
	ctx := NewContext(10 * time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	ctx = NewContext(10 * time.Second)
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugln("Connected to database")

	cursor, err := client.Database("admin").RunCommandCursor(NewContext(10*time.Second), bson.M{"replSetGetStatus": 1})
	if err != nil {
		if strings.Contains(err.Error(), "NoReplicationEnabled") {
			log.Info("switching to single member database")
		} else {
			log.Fatalf("admin replSetGetStatus: %v", err)
		}
	} else {
		for cursor.Next(ctx) {
			var result bson.M
			err := cursor.Decode(&result)
			if err != nil {
				log.Fatal(err)
			}
			log.Println(result)
		}
	}

	// Create unique zone indices
	for collection, key := range map[string]string{
		"zones": "zone",
		"users": "user",
	} {
		_, err = client.Database("cdnv3db").Collection(collection).Indexes().CreateOne(
			context.Background(),
			mongo.IndexModel{
				Keys:    bson.D{{Key: key, Value: 1}},
				Options: options.Index().SetUnique(true),
			},
		)
		if err != nil {
			log.Fatal(err)
		}
	}

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
	ctx := NewContext(10 * time.Second)
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
