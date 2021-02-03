package types

// Node stores a single edge node
type Node struct {
	ID         string  `json:"id,omitempty" bson:"_id,omitempty"`
	Provider   string  `json:"provider"`
	Latitude   float32 `json:"latitude"`
	Longitude  float32 `json:"longitude"`
	Authorized bool    `json:"authorized"`
}
