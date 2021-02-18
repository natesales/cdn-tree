// Package local provides functions for local controlplane operations
package local

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
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

// DisableBgp stops the BIRD BGP daemon to withdraw controlplane routes
func DisableBgp() error {
	birdSocket := "/run/bird/bird.ctl"
	conn, err := net.Dial("unix", birdSocket)
	if err != nil {
		return errors.New("bird socket connect: " + err.Error())
	}
	defer conn.Close()

	// Send the down command
	_, err = conn.Write([]byte("down"))
	if err != nil {
		return errors.New("bird write: " + err.Error())
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf[:])
	if err != nil {
		return errors.New("bird read: " + err.Error())
	}

	fmt.Println("bird response " + string(buf[:n]))

	return nil // nil error
}
