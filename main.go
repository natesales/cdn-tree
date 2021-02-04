package main

import (
	"errors"
	"github.com/natesales/cdnv3/internal/crypto"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"

	"github.com/natesales/cdnv3/internal/database"
	"github.com/natesales/cdnv3/internal/types"
)

var (
	db       *database.Database
	validate *validator.Validate
)

// Response helpers

// sendError helps return a JSON error message from a go error type
func sendError(ctx *fiber.Ctx, code int, err error) error {
	return ctx.Status(code).JSON(map[string]interface{}{"success": false, "message": err.Error()})
}

// HTTP endpoint handlers

// handleAddNode handles a HTTP POST request to add a new node
func handleAddNode(ctx *fiber.Ctx) error {
	newNode := new(types.Node)

	// Parse body into struct
	if err := ctx.BodyParser(newNode); err != nil {
		return sendError(ctx, 400, err)
	}

	// Validate node struct
	err := validate.Struct(newNode)
	if err != nil {
		return sendError(ctx, 400, err)
	}

	// Insert the new node
	insertionResult, err := db.Db.Collection("nodes").InsertOne(database.NewContext(10*time.Second), newNode)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error collection") {
			return ctx.Status(400).SendString(err.Error())
		}
		return ctx.Status(500).SendString(err.Error())
	}

	log.Printf("Inserted new node: %s\n", insertionResult.InsertedID)
	return ctx.Status(201).JSON(insertionResult)
}

// handleAddZone handles a HTTP POST request to add a new zone
func handleAddZone(ctx *fiber.Ctx) error {
	newZone := new(types.Zone)

	// Parse body into struct
	if err := ctx.BodyParser(newZone); err != nil {
		return sendError(ctx, 400, err)
	}

	// Validate zone struct
	err := validate.Struct(newZone)
	if err != nil {
		return sendError(ctx, 400, err)
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
	insertionResult, err := db.Db.Collection("zones").InsertOne(database.NewContext(10*time.Second), newZone)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error collection") {
			return ctx.Status(400).SendString(err.Error())
		}
		return ctx.Status(500).SendString(err.Error())
	}

	log.Printf("Inserted new zone: %s\n", insertionResult.InsertedID)
	return ctx.Status(201).JSON(insertionResult)
}

// handleAddRecord handles a HTTP POST request to create a new DNS record
func handleAddRecord(ctx *fiber.Ctx) error {
	// Get zone to add record to
	zoneID, err := primitive.ObjectIDFromHex(ctx.Params("zone"))
	if err != nil {
		return sendError(ctx, 400, errors.New("invalid zone ID"))
	}

	// New record struct
	newRecord := new(types.Record)

	// Parse body into struct
	if err := ctx.BodyParser(newRecord); err != nil {
		return sendError(ctx, 400, err)
	}

	// Validate struct
	err = validate.Struct(newRecord)
	if err != nil {
		return sendError(ctx, 400, err)
	}

	// Push the new record
	pushResult, err := db.Db.Collection("zones").UpdateOne(
		database.NewContext(10*time.Second),
		bson.M{"_id": zoneID},
		bson.M{"$push": bson.M{"records": newRecord}},
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error collection") {
			return sendError(ctx, 400, err)
		}
		return sendError(ctx, 500, err)
	}

	if pushResult.ModifiedCount < 1 {
		return sendError(ctx, 400, errors.New("zone with given ID doesn't exist"))
	}

	log.Printf("Added new record: %v\n", newRecord)
	return ctx.Status(201).SendString("added")
}

func main() {
	log.SetLevel(log.DebugLevel)

	db = database.New("mongodb://localhost:27017")

	validate = validator.New()

	app := fiber.New()
	app.Post("/nodes/add", handleAddNode)
	app.Post("/zones/add", handleAddZone)
	app.Post("/zones/:zone/add", handleAddRecord)

	log.Println("Starting API")
	log.Fatal(app.Listen(":3000"))
}
