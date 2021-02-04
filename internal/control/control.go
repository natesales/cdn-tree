package control

import (
	"github.com/natesales/cdnv3/internal/database"
	"github.com/natesales/cdnv3/internal/types"
	"go.mongodb.org/mongo-driver/bson"
	"time"
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
