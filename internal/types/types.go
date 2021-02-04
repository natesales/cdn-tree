package types

// Node stores a single edge node
type Node struct {
	ID         string  `json:"id,omitempty" bson:"_id,omitempty"`
	Provider   string  `json:"provider"`
	Latitude   float32 `json:"latitude"`
	Longitude  float32 `json:"longitude"`
	Authorized bool    `json:"authorized"`
}

// Record stores a DNS record
type Record struct {
	Label string `json:"label"`
	TTL   uint64 `json:"ttl"`
	Value string `json:"value"`
}

// Zone stores a DNS zone
type Zone struct {
	ID      string   `json:"id,omitempty" bson:"_id,omitempty"`
	Zone    string   `json:"zone"`
	Users   []string `json:"users"`
	Serial  uint64   `json:"serial"`
	Records []string `json:"records"`
}
