package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
)

var release = "dev" // Set by build process

type Config struct {
	ID string `json:"id"`
}

var (
	config     Config
	listenAddr = flag.String("l", ":8001", "Listen address:port to bind to")
	configFile = flag.String("c", "/opt/packetframe-eca.json", "JSON config file")
)

// loadConfig reads the configuration file and returns a Config struct
func loadConfig() Config {
	dat, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	config := Config{}
	err = json.Unmarshal(dat, &config)
	if err != nil {
		log.Fatal(err)
	}

	return config
}

// handleMeta handles a HTTP GET request for node metadata
func handleMeta(w http.ResponseWriter, r *http.Request) {
	jsonData, err := json.Marshal(config)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func main() {
	log.Printf("Initializing Packetframe ECA (%s)\n", release)

	// Parse command line flags
	flag.Parse()

	// Load config from JSON file
	config = loadConfig()

	log.Printf("Using node ID %s\n", config.ID)

	// HTTP handlers
	http.HandleFunc("/meta", handleMeta)

	log.Println("Starting HTTP server")
	// Start the HTTP server
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
