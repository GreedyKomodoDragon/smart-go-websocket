package websocket

import (
	"fmt"
	"go-websocket/pkg/db"

	socketio "github.com/googollee/go-socket.io"
)

func SocketFactory(meditor *webDataBridge) *socketio.Server {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
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

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		fmt.Println("closed", reason)
	})

	return server
}
