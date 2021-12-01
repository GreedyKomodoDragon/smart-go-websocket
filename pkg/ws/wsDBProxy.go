package ws

import (
	"fmt"
	"go-websocket/pkg/db"
)

type WSDBProxy struct {
	DatabaseManager *db.ISmartDBWriterReader
	IdToUsername    map[string]string
}

func (ws WSDBProxy) ConnectUsernameToID(username *string, id string) error {

	if ws.IdToUsername == nil {
		return fmt.Errorf("map has not been intialised")
	}

	ws.IdToUsername[id] = *username

	fmt.Println("id: ", id, " username", *username)

	return nil

}

func (ws WSDBProxy) IsLoggedIn(id string) bool {
	fmt.Println("id is: ", id)
	return ws.IdToUsername != nil && ws.IdToUsername[id] != ""
}

func (ws WSDBProxy) IsIDLinkedToUsername(id string, username *string) bool {
	return ws.IdToUsername != nil && ws.IdToUsername[id] != *username
}

func (ws WSDBProxy) LogoutID(id string) error {
	if ws.IdToUsername == nil {
		return fmt.Errorf("map has not been intialised")
	}

	if ws.IdToUsername[id] == "" {
		return fmt.Errorf("id has no logged in value")
	}

	// set back to the default value
	ws.IdToUsername[id] = ""

	return nil
}

func (ws WSDBProxy) CheckLogin(email, password *string) (string, error) {

	if ws.DatabaseManager == nil {
		return "", fmt.Errorf("DatabaseManager has not been intialised")
	}

	username, err := (*ws.DatabaseManager).CheckLogin(email, password)

	if err != nil {
		return "", err
	}

	return username, nil
}

func (ws WSDBProxy) GetMessages(socketID string, otherUsername *string, time *int64) ([]db.Messages, error) {

	if ws.DatabaseManager == nil {
		return nil, fmt.Errorf("DatabaseManager has not been intialised")
	}

	if isLoggedIn := ws.IsLoggedIn(socketID); isLoggedIn {

		username := ws.IdToUsername[socketID]

		messages, err := (*ws.DatabaseManager).GetMessages(&username, otherUsername, time)

		if err != nil {
			return nil, err
		}

		return messages, nil

	}

	return nil, fmt.Errorf("user is not logged in")
}

func (ws WSDBProxy) CreateMessage(socketID string, receiverUsername, message *string) error {

	if ws.DatabaseManager == nil {
		return fmt.Errorf("DatabaseManager has not been intialised")
	}

	if isLoggedIn := ws.IsLoggedIn(socketID); isLoggedIn {

		username := ws.IdToUsername[socketID]

		// TODO: Add Validation to check message length and validation for attacks

		err := (*ws.DatabaseManager).CreateMessage(&username, receiverUsername, message)

		if err != nil {
			return err
		}

		return nil

	}

	return fmt.Errorf("user is not logged in")
}

func (ws WSDBProxy) UploadListing(socketID string, listing *db.Listing) error {

	if ws.DatabaseManager == nil {
		return fmt.Errorf("DatabaseManager has not been intialised")
	}

	if isLoggedIn := ws.IsLoggedIn(socketID); isLoggedIn {

		username := ws.IdToUsername[socketID]

		// TODO: Add Validation to listing before it is unloaded

		_, err := (*ws.DatabaseManager).UploadListing(&username, listing)

		if err != nil {
			return err
		}

		return nil

	}

	return fmt.Errorf("user is not logged in")
}

func (ws WSDBProxy) BuyListing(socketID string, listingID *int64, amount *int64) error {

	if ws.DatabaseManager == nil {
		return fmt.Errorf("DatabaseManager has not been intialised")
	}

	if isLoggedIn := ws.IsLoggedIn(socketID); isLoggedIn {

		username := ws.IdToUsername[socketID]

		// TODO: Add Validation to check amount makes sense

		err := (*ws.DatabaseManager).BuyListing(&username, listingID, amount)

		if err != nil {
			return err
		}

		return nil

	}

	return fmt.Errorf("user is not logged in")
}

func (ws WSDBProxy) CreateProfile(username, email, password *string) error {

	if ws.DatabaseManager == nil {
		return fmt.Errorf("DatabaseManager has not been intialised")
	}

	// TODO: Add Validation to check amount makes sense

	err := (*ws.DatabaseManager).CreateProfile(username, email, password)

	if err != nil {
		return err
	}

	return nil
}

func (ws WSDBProxy) GetContacts(socketID string) ([]db.Contact, error) {

	if ws.DatabaseManager == nil {
		return nil, fmt.Errorf("DatabaseManager has not been intialised")
	}

	if isLoggedIn := ws.IsLoggedIn(socketID); isLoggedIn {

		username := ws.IdToUsername[socketID]

		contacts, err := (*ws.DatabaseManager).GetContacts(&username)

		if err != nil {
			return nil, err
		}

		return contacts, nil

	}

	return nil, fmt.Errorf("user is not logged in")

}
