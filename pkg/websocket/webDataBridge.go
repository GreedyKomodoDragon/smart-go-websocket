package websocket

import "go-websocket/pkg/db"

type webDataBridge interface {
	// Socket methods
	ConnectUsernameToID(username *string, id string) error
	IsLoggedIn(id string) bool
	IsIDLinkedToUsername(id string, username *string) bool
	LogoutID(id string) error

	// DB based methods
	CheckLogin(username, password *string) (bool, error)
	GetMessages(socketID string, otherUsername *string, time *int64) ([]db.Messages, error)
	CreateMessage(socketID string, receiverUsername, message *string) error
	UploadListing(socketID string, listing *db.Listing) error
	BuyListing(socketID string, listingID *int64, amount *int64) error
	CreateProfile(username, email, password *string) error
}
