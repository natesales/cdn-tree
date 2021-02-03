package main

import (
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	socketio "github.com/googollee/go-socket.io"
	"github.com/natesales/cdnv3/internal/database"
	"github.com/natesales/cdnv3/internal/transport"
	log "github.com/sirupsen/logrus"
	"time"
)

var (
	sio *socketio.Server // Socket.IO server
)

// HTTP endpoint handlers

// handlePing emits a ping message to all ECAs
func handlePing(c *fiber.Ctx) error {
	log.Println("Sending global ping")
	sio.BroadcastToRoom("/", "global", "global_ping")
	return c.SendString("Sent global ping")
}

// handleConnections shows a list of nodes and their last message timestamp
func handleConnections(c *fiber.Ctx) error {
	log.Println("Getting connections")
	sio.ForEach("/", "global", func(s socketio.Conn) {
		lastMessage := time.Since(time.Unix(s.Context().(int64), 0)).Truncate(time.Millisecond) // assert context type to int64, parse as UNIX timestamp, compute time since then, and truncate to milliseconds
		log.Printf("%s last message %s\n", transport.GetAuthKey(s), lastMessage)
	})
	return c.SendString("Displayed connections")
}

func main() {
	log.SetLevel(log.DebugLevel)

	db := database.New("mongodb://localhost:27017")

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
	log.Fatal(app.Listen(":3000"))
}
