package ws

import (
	"encoding/json"
	"fmt"
	ws "go-websocket/pkg/ws/datastructs"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func AssignRoutes(proxy *WebDataProxy, mux *http.ServeMux) error {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	// TODO: make it check over a list of valid origins
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// TODO: error ignored -> needs to be handled
		conn, err := upgrader.Upgrade(w, r, nil)

		fmt.Println("User connected to websocket")

		if err != nil {
			fmt.Println("some problem happened here")
			fmt.Println(err.Error())
		}

		for {
			// Read message from browser
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			// Convert to more readable type
			msgStr := string(msg)

			switch msgStr {
			case "login":
				fmt.Println("User attempted login")

				_, jsonData, _ := conn.ReadMessage()

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
				http.SetCookie(w, &http.Cookie{
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
				fmt.Println(result)

				if err != nil {
					fmt.Println(err)
				}

				if err = conn.WriteMessage(msgType, resultJson); err != nil {
					return
				}

			case "registration":
				_, jsonData, err := conn.ReadMessage()

				// TODO: improve error handling -> send err message to client
				if err != nil {
					return
				}

				var reg ws.Registration
				err = json.Unmarshal(jsonData, &reg)

				if err != nil {
					// send some message to the client
					continue
				}

				//TODO: Check that parameters are valid
				fmt.Println(reg)

				// Create the profile from parameters
				//err = (*proxy).CreateProfile(&reg.Username, &reg.Email, &reg.Password)

				/*if err != nil {
					result := ws.RegistrationResult{
						BaseMessage: ws.BaseMessage{
							Command: "regResult",
						},
						Message: "Failed but this should be more informative",
					}

					returnJSON, _ := json.Marshal(result)

					if err = conn.WriteMessage(msgType, returnJSON); err != nil {
						return
					}

					return
				}
				*/

				expirationTime := time.Now().Add(24 * time.Hour)
				token, _ := CreateToken("golangServer", expirationTime)

				http.SetCookie(w, &http.Cookie{
					Name:     "token",
					Value:    token,
					Expires:  expirationTime,
					HttpOnly: true,
				})

				result := ws.RegistrationResult{
					BaseMessage: ws.BaseMessage{
						Command: "regResult",
					},
					Message:  "success",
					Username: reg.Username,
				}

				returnJSON, _ := json.Marshal(result)
				conn.WriteMessage(msgType, returnJSON)

			case "logout":
				// delete cookie on log out
				http.SetCookie(w, &http.Cookie{
					Name:     "token",
					Value:    "",
					Expires:  time.Unix(0, 0),
					HttpOnly: true,
				})

				fmt.Printf("%s logout: %s\n", conn.RemoteAddr(), string(msg))

			default:
				fmt.Printf("Unknown Message: %s\n", msgStr)
			}

		}
	})

	return nil
}

/*
	server.OnConnect("/", func(s socketio.Conn) error {
		fmt.Println("connected in here")
		s.SetContext("")
		fmt.Println("connected:", s.ID())
		return nil
	})

	server.OnEvent("/", "login", func(s socketio.Conn, username, password *string) {
		// Cannot log in twice
		if isLoggedIn := (*meditor).IsLoggedIn(s.ID()); isLoggedIn {
			isValid, err := (*meditor).CheckLogin(username, password)

			// Do better error handling in future -> more informative handling
			if err != nil {
				s.Emit("loginResult", false)
			} else if isValid {
				if err = (*meditor).ConnectUsernameToID(username, s.ID()); err == nil {
					s.Emit("loginResult", true)
				} else {
					s.Emit("loginResult", false)
				}
			} else {
				s.Emit("loginResult", false)
			}

		} else {
			s.Emit("loginResult", false)
		}
	})

	server.OnEvent("/", "example", func(s socketio.Conn, msg string) {
		fmt.Println(msg)
	})

	server.OnEvent("/", "logout", func(s socketio.Conn, msg string) {
		err := (*meditor).LogoutID(s.ID())

		if err != nil {
			s.Emit("logout", false)
		} else {
			s.Emit("logout", true)
		}

		fmt.Println("notice:", msg)
		s.Emit("reply", "have "+msg)
	})

	server.OnEvent("/", "uploadMessage", func(s socketio.Conn, msg string, reciever string) {

		if isLoggedIn := (*meditor).IsLoggedIn(s.ID()); isLoggedIn {

			err := (*meditor).CreateMessage(s.ID(), &reciever, &msg)

			// Improve err message handling here
			if err != nil {
				s.Emit("messageUpload", false)
			} else {
				s.Emit("messageUpload", true)
			}

		} else {
			s.Emit("messageUpload", false)
		}

	})

	server.OnEvent("/", "uploadListing", func(s socketio.Conn, title, description, sym string, images []string, price int64) {
		if isLoggedIn := (*meditor).IsLoggedIn(s.ID()); isLoggedIn {

			// Create the listing from parameters
			listing := db.Listing{
				Title:      title,
				Decription: description,
				Sym:        sym,
				Images:     images,
				Price:      price,
			}

			err := (*meditor).UploadListing(s.ID(), &listing)

			// Improve err message handling here
			if err != nil {
				s.Emit("listingUpload", false)
			} else {
				s.Emit("listingUpload", true)
			}

		} else {
			s.Emit("messageUpload", false)
		}
	})

	server.OnEvent("/", "buyListing", func(s socketio.Conn, listingID int64, amount int64) {
		if isLoggedIn := (*meditor).IsLoggedIn(s.ID()); isLoggedIn {

			// Create the listing from parameters
			err := (*meditor).BuyListing(s.ID(), &listingID, &amount)

			// Improve err message handling here
			if err != nil {
				s.Emit("broughtOnServer", false)
			} else {
				s.Emit("broughtOnServer", true)
			}

		} else {
			s.Emit("messageUpload", false)
		}

	})

	server.OnEvent("/", "registration", func(s socketio.Conn, email, username, password string) {

		// Cannot be logged in to create an account
		if isLoggedIn := (*meditor).IsLoggedIn(s.ID()); !isLoggedIn {

			// Create the listing from parameters
			err := (*meditor).CreateProfile(&username, &email, &password)

			// Improve err message handling here
			if err != nil {
				s.Emit("listingUpload", false)
			} else {
				// Login them in
				err = (*meditor).ConnectUsernameToID(&username, s.ID())

				// Improve err message handling here
				// Account has been created but could not log in
				// User needs to be aware of this if this happens
				if err != nil {
					s.Emit("listingUpload", false)
				} else {
					// created and logged in
					s.Emit("listingUpload", true)
				}
				s.Emit("listingUpload", true)
			}

		} else {
			s.Emit("messageUpload", false)
		}

	})

	server.OnEvent("/", "Create", func(s socketio.Conn, email, username, password string) {})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("closed", reason)
	})

	return server
}

*/
