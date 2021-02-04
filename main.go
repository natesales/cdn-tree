package main

import (
	"errors"
	"github.com/natesales/cdnv3/internal/crypto"
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
	if newNode.Latitude == 0 || newNode.Longitude == 0 {
		return sendError(ctx, 400, errors.New("invalid longitude and/or latitude"))
	}
	if newNode.Provider == "" {
		return sendError(ctx, 400, errors.New("invalid provider string"))
	}

	// Insert the new node
	insertionResult, err := db.Db.Collection("nodes").InsertOne(database.NewContext(10*time.Second), newNode)
	if err != nil {
		return sendError(ctx, 500, err)
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

	// Insert the new zone
	insertionResult, err := db.Db.Collection("zones").InsertOne(database.NewContext(10*time.Second), newZone)
	if err != nil {
		return ctx.Status(500).SendString(err.Error())
	}

	log.Printf("Inserted new zone: %s\n", insertionResult.InsertedID)
	return ctx.Status(201).JSON(insertionResult)
}

func main() {
	log.SetLevel(log.DebugLevel)

	db = database.New("mongodb://localhost:27017")

	validate = validator.New()

	app := fiber.New()
	app.Post("/nodes/add", handleAddNode)
	app.Post("/zones/add", handleAddZone)

	log.Println("Starting API")
	log.Fatal(app.Listen(":3000"))
}
