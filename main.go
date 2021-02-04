package main

import (
	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"time"

	"github.com/natesales/cdnv3/internal/database"
	"github.com/natesales/cdnv3/internal/types"
)

var (
	db *database.Database
)

// HTTP endpoint handlers

// handleAddNode handles a HTTP POST request to add a new node
func handleAddNode(ctx *fiber.Ctx) error {
	newNode := new(types.Node)

	// Parse body into struct
	if err := ctx.BodyParser(newNode); err != nil {
		return ctx.Status(400).SendString(err.Error())
	}

	// Validate node struct
	if newNode.Latitude == 0 || newNode.Longitude == 0 {
		return ctx.Status(400).SendString("Invalid longitude and/or latitude")
	}
	if newNode.Provider == "" {
		return ctx.Status(400).SendString("Invalid provider string")
	}

	// Insert the new node
	insertionResult, err := db.Db.Collection("nodes").InsertOne(database.NewContext(10*time.Second), newNode)
	if err != nil {
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
		return ctx.Status(400).SendString(err.Error())
	}

	// Insert the new node
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

	app := fiber.New()
	app.Post("/nodes/add", handleAddNode)
	app.Post("/zones/add", handleAddZone)

	log.Println("Starting gofiber API")
	log.Fatal(app.Listen(":3000"))
}
