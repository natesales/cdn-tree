package main

import (
	"context"
	"github.com/googollee/go-socket.io"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
	"time"
)

// Enable verbose logging
const debug = false

var (
	nodesCollection *mongo.Collection // MongoDB node collection
	sio             *socketio.Server  // Socket.IO server
)

// newContext returns a context with given duration
func newContext(duration time.Duration) context.Context {
	ctx, _ := context.WithTimeout(context.Background(), duration)
	return ctx
}

// getAuthKey returns a ECA's provided authentication header value
func getAuthKey(s socketio.Conn) string {
	return s.RemoteHeader().Get("X-Packetframe-Eca-Auth")
}

// dbConnect opens a connection to the MongoDB database
func dbConnect() {
	// Connect to DB
	ctx := newContext(10 * time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	ctx = newContext(10 * time.Second)
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to database")

	// Nodes DB collection
	nodesCollection = client.Database("cdnv3db").Collection("ecas")
}

// getNode looks up a node by string ID
func getNode(id string) bson.M {
	nodeObjectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Debug(err)
		return nil // Invalid ObjectId
	}

	var node bson.M
	ctx := newContext(10 * time.Second)
	// Run DB query
	if err := nodesCollection.FindOne(ctx, bson.M{"_id": nodeObjectId}).Decode(&node); err != nil {
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

// HTTP handlers

// handlePing emits a ping message to all ECAs
func handlePing(w http.ResponseWriter, r *http.Request) {
	log.Println("Sending global ping")
	sio.BroadcastToRoom("/", "global", "global_ping")
}

// handlePing emits a ping message to all ECAs
func handleConnections(w http.ResponseWriter, r *http.Request) {
	log.Println("Getting connections")
	sio.ForEach("/", "global", func(s socketio.Conn) {
		lastMessage := time.Since(time.Unix(s.Context().(int64), 0))
		log.Printf("%s last message %s\n", getAuthKey(s), lastMessage)
	})
}

func main() {
	// Create a new socket.io server
	sio = socketio.NewServer(nil)

	// Listen for socket.io client connections from ECAs
	sio.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext(time.Now().Unix()) // Set last message receive time
		//s.SetContext("") // TODO: This should store temporary ECA data for the duration of the current connection
		log.Println("ECA connected:", s.ID(), s.RemoteAddr(), s.RemoteHeader())

		node := getNode(getAuthKey(s))
		if node == nil {
			log.Warnf("Node not found or not allowed, terminating connection")
			s.Emit("terminate", nil)
			return nil // exit gracefully
		}

		log.Printf("ECA %s connected, authorizing now\n", getAuthKey(s))
		s.Join("global")
		log.Println(node)

		return nil
	})

	sio.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Printf("ECA %s disconnected: %s\n", getAuthKey(s), reason)
	})

	sio.OnError("/", func(s socketio.Conn, e error) {
		log.Println("socket.io error:", e)
	})

	sio.OnEvent("/", "global_pong", func(s socketio.Conn, e error) {
		log.Println("Received pong from", getAuthKey(s))
		s.SetContext(time.Now().Unix())
	})

	// Connect to database
	dbConnect()

	// start socket.io handler
	log.Println("Starting socket server goroutine")
	go sio.Serve()
	defer sio.Close()

	// Setup routes
	http.Handle("/socket.io/", sio)
	http.HandleFunc("/ping", handlePing)
	http.HandleFunc("/connections", handleConnections)

	// Start HTTP server
	log.Println("Serving at localhost:8000...")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
