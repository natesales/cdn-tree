package config

import (
	etcd "go.etcd.io/etcd/client/v3"
	"time"
)

// Client wraps a etcd.Client
type Client struct {
	client *etcd.Client
}

// New constructs a new Client
func New(endpoints []string) *Client {
	client, _ := etcd.New(etcd.Config{
		DialTimeout: time.Second * 10,
		Endpoints:   endpoints,
	})
	defer client.Close()

	return &Client{client}
}
