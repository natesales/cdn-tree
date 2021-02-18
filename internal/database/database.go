// Package database provides functions and types for interacting with MongoDB, as well as a message queue built on mongo
package database

import (
	"context"
	"errors"
	"github.com/natesales/cdn-tree/internal/crypto"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strings"
	"time"
)

// Error constants
var (
	ErrNoMessagesInQueue = errors.New("no messages in queue")
)

// Database wraps a *mongo.Database
type Database struct {
	Db *mongo.Database
}

// Node stores a single edge node
type Node struct {
	ID         string  `json:"-" bson:"_id,omitempty"`
	Endpoint   string  `json:"endpoint" validate:"required"`
	Provider   string  `json:"provider" validate:"required"`
	Latitude   float32 `json:"latitude" validate:"required"`
	Longitude  float32 `json:"longitude" validate:"required"`
	Region     string  `json:"region" validate:"region"`
	Authorized bool    `json:"-"`
}

// DNSRecord stores a DNS RR string
type DNSRecord struct {
	RRString string `json:"rr" validate:"required"`
}

// Zone stores a DNS zone
type Zone struct {
	ID      string           `json:"-" bson:"_id,omitempty"`
	Zone    string           `json:"zone" validate:"required,fqdn"`
	Users   []string         `json:"-"`
	Serial  uint64           `json:"-"`
	Records []string         `json:"-"`
	DNSSEC  crypto.DNSSECKey `json:"-"`
}

// User stores a CDN user
type User struct {
	ID       string `json:"-" bson:"_id,omitempty"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	APIKey   string `json:"-"`
	Enabled  bool   `json:"-"`
	Admin    bool   `json:"-"`
	Hash     []byte `json:"-"`
}

// QueueMessage stores a single queue entry
type QueueMessage struct {
	ID       primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	Payload  map[string]string  `json:"payload"`
	Created  int64              `json:"-"`
	Locked   bool               `json:"-"`
	LockedAt int64              `json:"-"`
}

// Metadata label "enum"
type MetaLabel int

const (
	LabelAcmeAccount MetaLabel = iota
	LabelNetworkConfig
)

// String gets the string representation of MetaLabel
func (l MetaLabel) String() string {
	return [...]string{"LabelNetworkConfig", "LabelNetworkConfig"}[l]
}

// MetadataElement stores a document in the metadata collection in mongo
type MetadataElement struct {
	ID      primitive.ObjectID `bson:"-" bson:"_id,omitempty"`
	Label   string             `bson:"label"`
	Payload map[string]string  `bson:"payload"`
}

// member contains a replica set node entry
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

// Message Queue

// AddQueueMessage appends a message to the queue
func (d Database) AddQueueMessage(message QueueMessage) error {
	// Set created timestamp
	message.Created = time.Now().UnixNano()

	// Disable work lock
	message.Locked = false

	// Insert the new message
	_, err := d.Db.Collection("queue").InsertOne(context.Background(), message)
	if err != nil {
		return err
	}

	return nil // nil error
}

// NextQueueMessage retrieves a single queue message
func (d Database) NextQueueMessage() (QueueMessage, error) {
	// Find an available and unlocked queue message
	var message QueueMessage
	if err := d.Db.Collection("queue").FindOne(context.Background(), bson.M{"locked": false}).Decode(&message); err != nil {
		if err.Error() == "mongo: no documents in result" {
			return QueueMessage{}, ErrNoMessagesInQueue
		} else {
			return QueueMessage{}, err // Other error
		}
	}

	// Lock the message
	updateResult, err := d.Db.Collection("queue").UpdateOne(
		context.Background(),
		bson.M{"_id": message.ID},
		bson.M{"$set": bson.M{"locked": true, "lockedat": time.Now().UnixNano()}},
	)
	if err != nil {
		return QueueMessage{}, err // Other error
	}

	if updateResult.ModifiedCount < 1 {
		return QueueMessage{}, errors.New("unable to lock queue message") // Other error
	}

	return message, nil // nil error
}

// QueueConfirm marks a queue message as complete
func (d Database) QueueConfirm(message QueueMessage) error {
	_, err := d.Db.Collection("queue").DeleteOne(context.Background(), bson.M{"_id": message.ID})
	if err != nil {
		return err
	}

	return nil // nil error
}

// ListQueue returns all messages in queue
func (d Database) ListQueue() ([]QueueMessage, error) {
	cursor, err := d.Db.Collection("queue").Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err // nil message array
	}

	var messages []QueueMessage
	for cursor.Next(context.Background()) {
		var message QueueMessage
		if err := cursor.Decode(&message); err != nil {
			return nil, err // nil message array
		}
		// Append the message to the list
		messages = append(messages, message)
	}

	return messages, nil // nil error
}

// AddMetadata adds a MetadataElement to the database
func (d Database) AddMetadata(m MetadataElement) error {
	// Insert the new message
	_, err := d.Db.Collection("metadata").InsertOne(context.Background(), m)
	if err != nil {
		return err
	}

	return nil // nil error
}
