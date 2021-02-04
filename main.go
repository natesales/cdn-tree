package main

import (
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	socketio "github.com/googollee/go-socket.io"
	"github.com/natesales/cdnv3/internal/database"
	"github.com/natesales/cdnv3/internal/transport"
	"github.com/natesales/cdnv3/internal/types"
	log "github.com/sirupsen/logrus"
	"time"
)

var (
	sio *socketio.Server // Socket.IO server
	db  *database.Database
)

// HTTP endpoint handlers

// handlePing emits a ping message to all ECAs
func handlePing(ctx *fiber.Ctx) error {
	log.Println("Sending global ping")
	sio.BroadcastToRoom("/", "global", "global_ping")
	return ctx.SendString("Sent global ping")
}

// handleConnections shows a list of nodes and their last message timestamp
func handleConnections(ctx *fiber.Ctx) error {
	log.Println("Getting connections")
	sio.ForEach("/", "global", func(s socketio.Conn) {
		lastMessage := time.Since(time.Unix(s.Context().(int64), 0)).Truncate(time.Millisecond) // assert context type to int64, parse as UNIX timestamp, compute time since then, and truncate to milliseconds
		log.Printf("%s last message %s\n", transport.GetAuthKey(s), lastMessage)
	})
	return ctx.SendString("Displayed connections")
}

// handleAddNode handles a HTTP POST request to add a new node
func handleAddNode(c *fiber.Ctx) error {
	newNode := new(types.Node)

	// Parse body into struct
	if err := c.BodyParser(newNode); err != nil {
		return c.Status(400).SendString(err.Error())
	}

	// Insert the new node
	insertionResult, err := db.Db.Collection("nodes").InsertOne(database.NewContext(10*time.Second), newNode)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	log.Printf("Inserted new node: %s\n", insertionResult.InsertedID)
	return c.Status(201).JSON(insertionResult)
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

	// Create a new socket.io server
	sio, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Setup Socket.IO handlers
	transport.SetupHandlers(sio, db)

	// start socket.io handler
	log.Println("Starting socket server goroutine")
	//goland:noinspection GoUnhandledErrorResult
	go sio.Serve()
	//goland:noinspection GoUnhandledErrorResult
	defer sio.Close()

	app := fiber.New()
	app.Get("/ping", handlePing)
	app.Get("/connections", handleConnections)
	app.All("/socket.io/", adaptor.HTTPHandler(sio))

	app.Post("/nodes/add", handleAddNode)
	app.Post("/zones/add", handleAddZone)

	log.Println("Starting gofiber API")
	log.Fatal(app.Listen(":3000"))
}
