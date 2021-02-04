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
	"github.com/natesales/cdnv3/internal/util"
)

var (
	db       *database.Database
	validate *validator.Validate
)

// Helpers

// sendResponse helps return a JSON response message from a go error type or string
func sendResponse(ctx *fiber.Ctx, code int, reason interface{}, data interface{}) error {
	var success bool   // Did the request succeed?
	var message string // What did the request do?

	// Check if the reason type is an error
	switch reason.(type) {
	case error:
		success = false
		message = reason.(error).Error()
	default:
		success = true
		message = reason.(string)
	}

	return ctx.Status(code).JSON(map[string]interface{}{
		"success": success,
		"message": message,
		"data":    data,
	})
}

// requireGenericAuth checks if a user is authenticated and is present in the database
func requireGenericAuth(ctx *fiber.Ctx) (error, types.User) {
	// Get API key header
	apiKey := string(ctx.Request().Header.Peek("Authorization"))

	// Find user by API key in the database
	var user types.User
	result := db.Db.Collection("users").FindOne(database.NewContext(10*time.Second), &bson.M{"apikey": apiKey})
	// Decode database result into user struct
	err := result.Decode(&user)
	if err != nil {
		return err, types.User{}
	}

	return nil, user // no error; a user with this API key exists
}

// HTTP endpoint handlers

// handleAddNode handles a HTTP POST request to add a new node
func handleAddNode(ctx *fiber.Ctx) error {
	newNode := new(types.Node)

	// Parse body into struct
	if err := ctx.BodyParser(newNode); err != nil {
		return sendResponse(ctx, 400, err, nil)
	}

	// Validate node struct
	err := validate.Struct(newNode)
	if err != nil {
		return sendResponse(ctx, 400, err, nil)
	}

	// Insert the new node
	_, err = db.Db.Collection("nodes").InsertOne(database.NewContext(10*time.Second), newNode)
	if err != nil {
		return sendResponse(ctx, 500, err, nil)
	}

	// Return 201 Created OK response
	return sendResponse(ctx, 201, "added new node", nil)
}

// handleAddZone handles a HTTP POST request to add a new zone
func handleAddZone(ctx *fiber.Ctx) error {
	err, user := requireGenericAuth(ctx)
	if err != nil {
		return sendResponse(ctx, 403, errors.New("unauthorized"), nil)
	}

	newZone := new(types.Zone)

	// Parse body into struct
	if err := ctx.BodyParser(newZone); err != nil {
		return sendResponse(ctx, 400, err, nil)
	}

	// Validate zone struct
	err = validate.Struct(newZone)
	if err != nil {
		return sendResponse(ctx, 400, err, nil)
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
	newZone.Users = []string{user.ID}
	newZone.Records = []types.Record{}

	// Insert the new zone
	_, err = db.Db.Collection("zones").InsertOne(database.NewContext(10*time.Second), newZone)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error collection") {
			return sendResponse(ctx, 400, err, nil)
		}
		return sendResponse(ctx, 500, err, nil)
	}

	// Return 201 Created OK response
	return sendResponse(ctx, 201, "added new zone", nil)
}

// handleAddRecord handles a HTTP POST request to create a new DNS record
func handleAddRecord(ctx *fiber.Ctx) error {
	err, user := requireGenericAuth(ctx)
	if err != nil {
		return sendResponse(ctx, 403, errors.New("unauthorized"), nil)
	}

	// Get zone to add record to
	zoneID, err := primitive.ObjectIDFromHex(ctx.Params("zone"))
	if err != nil {
		return sendResponse(ctx, 400, errors.New("invalid zone ID"), nil)
	}

	// Find zone
	var zone types.Zone
	result := db.Db.Collection("zones").FindOne(database.NewContext(10*time.Second), &bson.M{"_id": zoneID})
	err = result.Decode(&zone)
	if err != nil || !util.Includes(zone.Users, user.ID) { // If error or the zone doesn't contain this user as authorized
		return sendResponse(ctx, 400, err, "decoding zone")
	}

	// New record struct
	newRecord := new(types.Record)

	// Parse body into struct
	if err := ctx.BodyParser(newRecord); err != nil {
		return sendResponse(ctx, 400, err, "parsing record body")
	}

	// Validate struct
	err = validate.Struct(newRecord)
	if err != nil {
		return sendResponse(ctx, 400, err, "validating record body")
	}

	// Push the new record
	pushResult, err := db.Db.Collection("zones").UpdateOne(
		database.NewContext(10*time.Second),
		bson.M{"_id": zoneID},
		bson.M{"$push": bson.M{"records": newRecord}},
	)
	if err != nil {
		return sendResponse(ctx, 400, err, "pushing new record")
	}

	// If nothing was modified (and there wasn't a previously caught error) then the zone must not exist
	if pushResult.ModifiedCount < 1 {
		return sendResponse(ctx, 400, errors.New("zone with given ID doesn't exist"), nil)
	}

	// Return 201 Created OK response
	return sendResponse(ctx, 201, "record added", nil)
}

// handleAddUser handles a HTTP POST request to create a new USER
func handleAddUser(ctx *fiber.Ctx) error {
	newUser := new(types.User)

	// Parse body into struct
	if err := ctx.BodyParser(newUser); err != nil {
		return sendResponse(ctx, 400, err, nil)
	}

	// Validate struct
	err := validate.Struct(newUser)
	if err != nil {
		return sendResponse(ctx, 400, err, nil)
	}

	// Set user defaults
	newUser.Enabled = false
	newUser.Admin = false

	// Generate a random API key
	newUser.APIKey = crypto.RandomString()

	// Compute the user's password hash
	newUser.Hash, err = crypto.PasswordHash(newUser.Password)
	if err != nil {
		return sendResponse(ctx, 500, err, nil)
	}

	// Zero out the plaintext password
	newUser.Password = ""

	// Insert the new node
	_, err = db.Db.Collection("users").InsertOne(database.NewContext(10*time.Second), newUser)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error collection") {
			return sendResponse(ctx, 400, err, nil)
		}
		return sendResponse(ctx, 500, err, nil)
	}

	// Return 201 Created OK response
	return sendResponse(ctx, 201, "added new user", nil)
}

// handleUserLogin handles a HTTP POST request to authenticate a user
func handleUserLogin(ctx *fiber.Ctx) error {
	loginReq := new(types.LoginRequest)

	// Parse body into struct
	if err := ctx.BodyParser(loginReq); err != nil {
		return sendResponse(ctx, 400, err, nil)
	}

	// Validate node struct
	err := validate.Struct(loginReq)
	if err != nil {
		return sendResponse(ctx, 400, err, nil)
	}

	// Find user by email
	var user types.User
	result := db.Db.Collection("users").FindOne(database.NewContext(10*time.Second), &bson.M{"email": loginReq.Email})
	err = result.Decode(&user)
	if err != nil {
		return sendResponse(ctx, 400, err, nil)
	}

	// Validate the provided hash with the stored one in database
	if crypto.ValidHash(user.Hash, loginReq.Password) {
		// If success, return the user's API key
		return sendResponse(ctx, 201, "user authenticated", map[string]string{"apikey": user.APIKey})
	} else {
		return sendResponse(ctx, 403, errors.New("unauthorized"), nil)
	}
}

func main() {
	log.SetLevel(log.DebugLevel)

	db = database.New("mongodb://localhost:27017")

	// Type/data validator
	validate = validator.New()

	// Fiber API server
	app := fiber.New()

	// API Routes

	// Node management
	// TODO: Authenticate these routes
	app.Post("/nodes/add", handleAddNode)

	// DNS management
	app.Post("/zones/add", handleAddZone)
	app.Post("/zones/:zone/add", handleAddRecord)

	// Authentication
	app.Post("/auth/register", handleAddUser)
	app.Post("/auth/login", handleUserLogin)

	log.Println("Starting API")
	log.Fatal(app.Listen(":3000"))
}
