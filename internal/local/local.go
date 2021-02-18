// Package local provides functions for local controlplane operations
package local

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// LoadCaddyConfig sets a new running Caddy config JSON
func LoadCaddyConfig(config map[string]interface{}) error {
	// Marshal config as JSON string
	jsonBody, err := json.Marshal(config)
	if err != nil {
		return err
	}

	// Create the load request
	req, err := http.NewRequest("POST", "http://localhost:2019/load", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	return nil // nil error
}
