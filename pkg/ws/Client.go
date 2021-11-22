package ws

import (
	"bytes"
	"encoding/json"
	"fmt"
	ws "go-websocket/pkg/ws/messages"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	// Response codes
	SUCCESS          byte = 0
	EMAIL_IN_USE     byte = 1
	EMAIL_INVALID    byte = 2
	PASSWORD_INVALID byte = 3
	USERAME_IN_USE   byte = 4
	USERAME_INVALID  byte = 5
	UNKNOWN          byte = 6
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	Hub *Hub

	// The websocket connection.
	Conn *websocket.Conn

	// Buffered channel of outbound messages.
	Send chan []byte

	// Http connection
	Web http.ResponseWriter

	// DB Proxy
	DB *WebDataProxy
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		msgType, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))

		switch string(message[:]) {

		case "login":
			fmt.Println("User attempted login")

			_, jsonData, _ := c.Conn.ReadMessage()

			var loginDetails ws.Login
			json.Unmarshal(jsonData, &loginDetails)

			// TODO: Check that format and JSON data is correct in terms of standards
			// Password have has to sent way so no point checking db if not valid
			// Same with emamil

			//TODO: Go to server and check login information
			username := "dbConfirmedUsername2"

			expirationTime := time.Now().Add(24 * time.Hour)
			token, _ := CreateToken(username, expirationTime)

			// Add the HTTP cookie to the clients cookie list
			http.SetCookie(c.Web, &http.Cookie{
				Name:     "token",
				Value:    token,
				Expires:  expirationTime,
				Path:     "/",
				HttpOnly: true,
			})

			// Message to tell the client that is was a success
			result := ws.LoginResult{
				BaseMessage: ws.BaseMessage{
					Command: "loginResult",
				},
				Result:   true,
				Username: username,
			}

			// Convert object to json
			resultJson, err := json.Marshal(result)

			if err != nil {
				fmt.Println(err)
			}

			if err = c.Conn.WriteMessage(msgType, resultJson); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				return
			}

		case "registration":
			_, jsonData, err := c.Conn.ReadMessage()

			// TODO: improve error handling -> send err message to client
			if err != nil {
				return
			}

			var reg ws.Registration
			err = json.Unmarshal(jsonData, &reg)

			if err != nil {
				// send some message to the client
				return
			}

			// Check email
			if !IsValid(reg.Email) {

				// sent invalid message back to the user
				result := ws.RegistrationResult{
					BaseMessage: ws.BaseMessage{
						Command: "regResult",
					},
					ResponseCode: EMAIL_INVALID,
				}

				returnJSON, _ := json.Marshal(result)

				if err = c.Conn.WriteMessage(msgType, returnJSON); err != nil {
					return
				}

				// don't do any more checks -> email should have been checked on frontend
				continue
			}

			// Check password
			if !IsValidPassword(reg.Password) {
				// sent invalid message back to the user
				result := ws.RegistrationResult{
					BaseMessage: ws.BaseMessage{
						Command: "regResult",
					},
					ResponseCode: PASSWORD_INVALID,
				}

				returnJSON, _ := json.Marshal(result)

				if err = c.Conn.WriteMessage(msgType, returnJSON); err != nil {
					return
				}

				// don't do any more checks -> email should have been checked on frontend
				continue
			}

			// Create the profile from parameters
			err = (*c.DB).CreateProfile(&reg.Username, &reg.Email, &reg.Password)

			if err != nil {
				// by default it is unknown
				var errorCode byte = UNKNOWN

				// extract the error being referenced from message
				errorType := strings.Split(err.Error(), ":")[0]

				switch errorType {

				case "email":
					errorCode = EMAIL_IN_USE

				case "username":
					errorCode = USERAME_IN_USE

				}

				result := ws.RegistrationResult{
					BaseMessage: ws.BaseMessage{
						Command: "regResult",
					},
					ResponseCode: errorCode,
				}

				returnJSON, _ := json.Marshal(result)

				if err = c.Conn.WriteMessage(msgType, returnJSON); err != nil {
					return
				}

				continue

			}

			expirationTime := time.Now().Add(24 * time.Hour)
			token, _ := CreateToken(reg.Username, expirationTime)

			http.SetCookie(c.Web, &http.Cookie{
				Name:     "token",
				Value:    token,
				Expires:  expirationTime,
				HttpOnly: true,
			})

			result := ws.RegistrationResult{
				BaseMessage: ws.BaseMessage{
					Command: "regResult",
				},
				ResponseCode: SUCCESS,
			}

			returnJSON, _ := json.Marshal(result)
			if err = c.Conn.WriteMessage(msgType, returnJSON); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				return
			}

		}

	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request, db *WebDataProxy) {
	conn, err := upgrader.Upgrade(w, r, nil)
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{Hub: hub, Conn: conn, Send: make(chan []byte, 256), Web: w, DB: db}
	client.Hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
