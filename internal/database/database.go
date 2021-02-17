package database

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
)

// Database wraps a *mongo.Database
type Database struct {
	Db *mongo.Database
}

// member contains a replica set member entry
type member struct {
	Name     string `bson:"name"`
	State    int    `bson:"state"`
	StateStr string `bson:"stateStr"`
	PingMs   uint   `bson:"pingMs"`
	Health   int    `bson:"health"`
}

// replicaSetResponse stores the response returned by replSetGetStatus
type replicaSetResponse struct {
	Members []member `bson:"members"`
}

// mongoUri connects to a local MongoDB server and extracts and assembles replica set information into a fully formed URI. If no replica set exists (running in development), it returns mongodb://localhost:27017
func mongoUri() (string, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return "", errors.New("client connect: " + err.Error()) // empty mongo URI
	}

	res := client.Database("admin").RunCommand(context.Background(), bson.D{{"replSetGetStatus", 1}})
	if res.Err() != nil {
		if strings.Contains(res.Err().Error(), "NoReplicationEnabled") {
			return "mongodb://localhost:27017", nil // nil error
		} else {
			return "", errors.New("replSetGetStatus query: " + res.Err().Error()) // empty mongo URI
		}
	}

	var result replicaSetResponse
	err = res.Decode(&result)
	if err != nil {
		return "", errors.New("decoding result: " + err.Error()) // empty mongo URI
	}

	var dbHosts []string
	for _, member := range result.Members {
		dbHosts = append(dbHosts, member.Name)
	}

	return "mongodb://" + strings.Join(dbHosts, ",") + "/?replSet=packetframe", nil // nil error
}

// New constructs a new database object
func New() *Database {
	dbUri, err := mongoUri()
	if err != nil {
		log.Fatalf("mongoUri: %v", err)
	}

	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(dbUri))
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Debugf("connected to database at %s", dbUri)

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
	// Run DB query
	if err := d.Db.Collection("nodes").FindOne(context.Background(), bson.M{"_id": nodeObjectId}).Decode(&node); err != nil {
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
