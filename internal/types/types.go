package types

import (
	"natesales.net/packetframe/cdn-v3/internal/crypto"
)

// Node stores a single edge node
type Node struct {
	ID         string  `json:"-" bson:"_id,omitempty"`
	Endpoint   string  `json:"endpoint" validate:"required"`
	Provider   string  `json:"provider" validate:"required"`
	Latitude   float32 `json:"latitude" validate:"required"`
	Longitude  float32 `json:"longitude" validate:"required"`
	Authorized bool    `json:"-"`
}

// Record stores a DNS RR string
type Record struct {
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

// LoginRequest stores a username/password combination
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
