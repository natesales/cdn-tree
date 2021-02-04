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
	Label string `json:"label"`
	TTL   uint64 `json:"ttl"`
	Value string `json:"value"`
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
