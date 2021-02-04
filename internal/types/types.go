package types

import "github.com/natesales/cdnv3/internal/crypto"

// Node stores a single edge node
type Node struct {
	ID         string  `json:"-" bson:"_id,omitempty"`
	Provider   string  `json:"provider" validate:"required"`
	Latitude   float32 `json:"latitude" validate:"required"`
	Longitude  float32 `json:"longitude" validate:"required"`
	Authorized bool    `json:"-"`
}

// Record stores a DNS record
type Record struct {
	Type  string `json:"type" validate:"required"`
	Label string `json:"label" validate:"required"`
	TTL   uint64 `json:"ttl" validate:"required"`
	Value string `json:"value" validate:"required"`
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
