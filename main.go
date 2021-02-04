package main

import (
	"errors"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/natesales/cdnv3/internal/crypto"
	"github.com/natesales/cdnv3/internal/database"
	"github.com/natesales/cdnv3/internal/types"
)

var (
	db       *database.Database
	validate *validator.Validate
)

// Response helpers

// sendResponse helps return a JSON response message from a go error type or string
func sendResponse(ctx *fiber.Ctx, code int, reason interface{}) error {
	var success bool // Did the request succeed?
	var message string

	// Check if the reason type is an error
	switch reason.(type) {
	case error:
		success = false
		message = reason.(error).Error()
	default:
		success = true
		message = reason.(string)
	}

	return ctx.Status(code).JSON(map[string]interface{}{"success": success, "message": message})
}

// HTTP endpoint handlers

// handleAddNode handles a HTTP POST request to add a new node
func handleAddNode(ctx *fiber.Ctx) error {
	newNode := new(types.Node)

	// Parse body into struct
	if err := ctx.BodyParser(newNode); err != nil {
		return sendResponse(ctx, 400, err)
	}

	// Validate node struct
	err := validate.Struct(newNode)
	if err != nil {
		return sendResponse(ctx, 400, err)
	}

	// Insert the new node
	_, err = db.Db.Collection("nodes").InsertOne(database.NewContext(10*time.Second), newNode)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error collection") {
			return sendResponse(ctx, 400, err)
		}
		return sendResponse(ctx, 500, err)
	}

	return sendResponse(ctx, 201, "added new node")
}

// handleAddZone handles a HTTP POST request to add a new zone
func handleAddZone(ctx *fiber.Ctx) error {
	newZone := new(types.Zone)

	// Parse body into struct
	if err := ctx.BodyParser(newZone); err != nil {
		return sendResponse(ctx, 400, err)
	}

	// Validate zone struct
	err := validate.Struct(newZone)
	if err != nil {
		return sendResponse(ctx, 400, err)
	}

	// Remove trailing dot if present
	if strings.HasSuffix(newZone.Zone, ".") {
		newZone.Zone = newZone.Zone[:len(newZone.Zone)-1]
	}

	// Set default zone serial
	newZone.Serial = uint64(time.Now().UnixNano())

	// Create DNSSEC key
	newZone.DNSSEC = crypto.NewKey(newZone.Zone)

	// Create empty arrays
	newZone.Users = []string{}
	newZone.Records = []string{}

	// Insert the new zone
	_, err = db.Db.Collection("zones").InsertOne(database.NewContext(10*time.Second), newZone)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error collection") {
			return sendResponse(ctx, 400, err)
		}
		return sendResponse(ctx, 500, err)
	}

	return sendResponse(ctx, 201, "added new zone")
}

// handleAddRecord handles a HTTP POST request to create a new DNS record
func handleAddRecord(ctx *fiber.Ctx) error {
	// Get zone to add record to
	zoneID, err := primitive.ObjectIDFromHex(ctx.Params("zone"))
	if err != nil {
		return sendResponse(ctx, 400, errors.New("invalid zone ID"))
	}

	// New record struct
	newRecord := new(types.Record)

	// Parse body into struct
	if err := ctx.BodyParser(newRecord); err != nil {
		return sendResponse(ctx, 400, err)
	}

	// Validate struct
	err = validate.Struct(newRecord)
	if err != nil {
		return sendResponse(ctx, 400, err)
	}

	// Push the new record
	pushResult, err := db.Db.Collection("zones").UpdateOne(
		database.NewContext(10*time.Second),
		bson.M{"_id": zoneID},
		bson.M{"$push": bson.M{"records": newRecord}},
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error collection") {
			return sendResponse(ctx, 400, err)
		}
		return sendResponse(ctx, 500, err)
	}

	// If nothing was modified (and there wasn't a previously caught error) then the zone must not exist
	if pushResult.ModifiedCount < 1 {
		return sendResponse(ctx, 400, errors.New("zone with given ID doesn't exist"))
	}

	return sendResponse(ctx, 201, "record added")
}

// handleAddUser handles a HTTP POST request to create a new USER
func handleAddUser(ctx *fiber.Ctx) error {
	newUser := new(types.User)

	// Parse body into struct
	if err := ctx.BodyParser(newUser); err != nil {
		return sendResponse(ctx, 400, err)
	}

	// Validate struct
	err := validate.Struct(newUser)
	if err != nil {
		return sendResponse(ctx, 400, err)
	}

	// Set user defaults
	newUser.Enabled = false
	newUser.Admin = false

	// Insert the new node
	_, err = db.Db.Collection("users").InsertOne(database.NewContext(10*time.Second), newUser)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error collection") {
			return sendResponse(ctx, 400, err)
		}
		return sendResponse(ctx, 500, err)
	}

	return sendResponse(ctx, 201, "added new user")
}

func main() {
	log.SetLevel(log.DebugLevel)

	db = database.New("mongodb://localhost:27017")

	validate = validator.New()

	app := fiber.New()
	app.Post("/nodes/add", handleAddNode)
	app.Post("/zones/add", handleAddZone)
	app.Post("/zones/:zone/add", handleAddRecord)

	// Authentication
	app.Post("/auth/register", handleAddUser)

	log.Println("Starting API")
	log.Fatal(app.Listen(":3000"))
}
