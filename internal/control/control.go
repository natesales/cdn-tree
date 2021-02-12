package control

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"sync"
	"time"

	"github.com/natesales/cdn-tree/internal/database"
	"github.com/natesales/cdn-tree/internal/types"
)

// Manifest gets a list of zone:serial pairs
func Manifest(db *database.Database) ([]map[string]interface{}, error) {
	// Find all zones from database
	cursor, err := db.Db.Collection("zones").Find(database.NewContext(10*time.Second), bson.M{})
	if err != nil {
		return nil, err // nil data
	}

	// Declare local zones manifest
	var zones []map[string]interface{}

	// Iterate over each zone and add to local zones manifest
	for cursor.Next(database.NewContext(10 * time.Second)) {
		var zone types.Zone
		err := cursor.Decode(&zone)
		if err != nil {
			return nil, err // nil error
		}

		// Append to local zones manifest
		zones = append(zones, map[string]interface{}{"zone": zone.Zone, "serial": zone.Serial})
	}

	return zones, nil // nil error
}

// MassRequest sends an HTTP POST request to all edge nodes
func MassRequest(db *database.Database, endpoint string, body interface{}) ([]*http.Response, error) {
	// Find all zones from database
	cursor, err := db.Db.Collection("nodes").Find(database.NewContext(10*time.Second), bson.M{})
	if err != nil {
		return nil, err
	}

	// Store list of node responses
	var responses []*http.Response

	// Response lock
	var wg sync.WaitGroup

	// Ignore insecure TLS certificates (self signed)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Iterate over each zone and add to local zones manifest
	for cursor.Next(database.NewContext(10 * time.Second)) {
		var node types.Node
		err := cursor.Decode(&node)
		if err != nil {
			return nil, err // nil data
		}

		// Marshal the body data to JSON
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err // nil data
		}

		// Create a new client with timeout and send the request
		httpClient := &http.Client{Timeout: time.Second * 10}

		// Add positive delta to WaitGroup
		wg.Add(1)

		// Make the request in a new goroutine
		go func() {
			// Defer lock release
			defer wg.Done()

			// Send the HTTP request
			log.Debugln("Sending HTTP POST to https://" + node.Endpoint + "/" + endpoint)
			response, err := httpClient.Post("https://"+node.Endpoint+"/"+endpoint, "application/json", bytes.NewBuffer(jsonBody))
			if err != nil {
				log.Warnf("node %s failed: %v\n", node.ID, err) // TODO: report this error...maybe return err?
			}
			log.Debugln("Received response from " + node.Endpoint)

			// Append the response to the array
			responses = append(responses, response)
		}()
	}

	return responses, nil // nil error
}

// Update sends a update request to all nodes
func Update(db *database.Database) {
	manifest, err := Manifest(db)
	if err != nil {
		log.Debug(err)
	}

	responses, err := MassRequest(db, "/update", manifest)
	if err != nil {
		log.Debug(err)
	}

	log.Println(responses)
}
