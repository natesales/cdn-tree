package transport

import (
	socketio "github.com/googollee/go-socket.io"
	"github.com/natesales/cdnv3/internal/database"
	log "github.com/sirupsen/logrus"
	"time"
)

// GetAuthKey returns a ECA's provided authentication header value
func GetAuthKey(s socketio.Conn) string {
	return s.RemoteHeader().Get("X-Packetframe-Eca-Auth")
}

func SetupHandlers(sio *socketio.Server, db *database.Database) {
	// Listen for socket.io client connections from ECAs
	sio.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext(time.Now().Unix()) // Set last message receive time
		//s.SetContext("") // TODO: This should store temporary ECA data for the duration of the current connection
		log.Println("ECA connected:", s.ID(), s.RemoteAddr(), s.RemoteHeader())

		node := db.GetNode(GetAuthKey(s))
		if node == nil {
			log.Warnf("Node not found or not allowed, terminating connection")
			s.Emit("terminate", "Node not found or not allowed")
			return nil // exit gracefully
		}

		log.Printf("ECA %s connected, authorizing now\n", GetAuthKey(s))
		s.Join("global")
		log.Println(node)

		return nil
	})

	sio.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Printf("ECA %s disconnected: %s\n", GetAuthKey(s), reason)
	})

	sio.OnError("/", func(s socketio.Conn, e error) {
		log.Println("socket.io error:", e)
	})

	sio.OnEvent("/", "global_pong", func(s socketio.Conn, e error) {
		log.Println("Received pong from", GetAuthKey(s))
		s.SetContext(time.Now().Unix())
	})
}
