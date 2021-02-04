package control

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"time"

	"github.com/natesales/cdnv3/internal/database"
	"github.com/natesales/cdnv3/internal/types"
)

// Manifest gets a list of zone:serial pairs
func Manifest(db *database.Database) (error, []map[string]interface{}) {
	// Find all zones from database
	cursor, err := db.Db.Collection("zones").Find(database.NewContext(10*time.Second), bson.M{})
	if err != nil {
		return err, nil
	}

	// Declare local zones manifest
	var zones []map[string]interface{}

	// Iterate over each zone and add to local zones manifest
	for cursor.Next(database.NewContext(10 * time.Second)) {
		var zone types.Zone
		err := cursor.Decode(&zone)
		if err != nil {
			return err, nil
		}

		// Append to local zones manifest
		zones = append(zones, map[string]interface{}{"zone": zone.Zone, "serial": zone.Serial})
	}

	return nil, zones // nil error
}

// MassRequest sends an HTTP POST request to all edge nodes
func MassRequest(db *database.Database, endpoint string, body interface{}) (error, []*http.Response) {
	// Find all zones from database
	cursor, err := db.Db.Collection("nodes").Find(database.NewContext(10*time.Second), bson.M{})
	if err != nil {
		return err, nil
	}

	// Store list of node responses
	var responses []*http.Response

	// Ignore insecure TLS certificates (self signed)
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// Iterate over each zone and add to local zones manifest
	for cursor.Next(database.NewContext(10 * time.Second)) {
		var node types.Node
		err := cursor.Decode(&node)
		if err != nil {
			return err, nil // nil data
		}

		// Marshal the body data to JSON
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return err, nil
		}

		// Create a client with timeout and send the request
		httpClient := &http.Client{Timeout: time.Second * 10}
		response, err := httpClient.Post("https://"+node.Endpoint+"/"+endpoint, "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			log.Warnf("node %s failed: %v\n", node.ID, err) // TODO: report this error...maybe return err?
		}

		// Append the response to the array
		responses = append(responses, response)
	}

	return nil, responses // nil error
}
