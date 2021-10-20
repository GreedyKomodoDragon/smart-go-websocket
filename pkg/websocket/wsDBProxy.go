package websocket

import (
	"fmt"
	"go-websocket/pkg/db"
)

type WSDBProxy struct {
	databaseManager *db.ISmartDBWriterReader
	idToUsername    map[string]string
}

func (ws WSDBProxy) ConnectUsernameToID(username *string, id string) error {

	if ws.idToUsername == nil {
		return fmt.Errorf("map has not been intialised")
	}

	if ws.idToUsername[id] == "" {
		ws.idToUsername[id] = *username

		return nil
	}

	return fmt.Errorf("id already has a connecting username")

}

func (ws WSDBProxy) IsLoggedIn(id string) bool {
	return ws.idToUsername != nil && ws.idToUsername[id] != ""
}

func (ws WSDBProxy) IsIDLinkedToUsername(id string, username *string) bool {
	return ws.idToUsername != nil && ws.idToUsername[id] != *username
}

func (ws WSDBProxy) LogoutID(id string) error {
	if ws.idToUsername == nil {
		return fmt.Errorf("map has not been intialised")
	}

	if ws.idToUsername[id] == "" {
		return fmt.Errorf("id has no logged in value")
	}

	// set back to the default value
	ws.idToUsername[id] = ""

	return nil
}

func (ws WSDBProxy) CheckLogin(username, password *string) (bool, error) {

	if ws.databaseManager == nil {
		return false, fmt.Errorf("databaseManager has not been intialised")
	}

	validDetails, err := (*ws.databaseManager).CheckLogin(username, password)

	if err != nil {
		return false, err
	}

	return validDetails, nil
}

func (ws WSDBProxy) GetMessages(socketID string, otherUsername *string, time *int64) ([]db.Messages, error) {

	if ws.databaseManager == nil {
		return nil, fmt.Errorf("databaseManager has not been intialised")
	}

	if isLoggedIn := ws.IsLoggedIn(socketID); isLoggedIn {

		username := ws.idToUsername[socketID]

		messages, err := (*ws.databaseManager).GetMessages(&username, otherUsername, time)

		if err != nil {
			return nil, err
		}

		return messages, nil

	}

	return nil, fmt.Errorf("user is not logged in")
}

func (ws WSDBProxy) CreateMessage(socketID string, receiverUsername, message *string) error {

	if ws.databaseManager == nil {
		return fmt.Errorf("databaseManager has not been intialised")
	}

	if isLoggedIn := ws.IsLoggedIn(socketID); isLoggedIn {

		username := ws.idToUsername[socketID]

		// TODO: Add Validation to check message length and validation for attacks

		err := (*ws.databaseManager).CreateMessage(&username, receiverUsername, message)

		if err != nil {
			return err
		}

		return nil

	}

	return fmt.Errorf("user is not logged in")
}

func (ws WSDBProxy) UploadListing(socketID string, listing *db.Listing) error {

	if ws.databaseManager == nil {
		return fmt.Errorf("databaseManager has not been intialised")
	}

	if isLoggedIn := ws.IsLoggedIn(socketID); isLoggedIn {

		username := ws.idToUsername[socketID]

		// TODO: Add Validation to listing before it is unloaded

		_, err := (*ws.databaseManager).UploadListing(&username, listing)

		if err != nil {
			return err
		}

		return nil

	}

	return fmt.Errorf("user is not logged in")
}

func (ws WSDBProxy) BuyListing(socketID string, listingID *int64, amount *int64) error {

	if ws.databaseManager == nil {
		return fmt.Errorf("databaseManager has not been intialised")
	}

	if isLoggedIn := ws.IsLoggedIn(socketID); isLoggedIn {

		username := ws.idToUsername[socketID]

		// TODO: Add Validation to check amount makes sense

		err := (*ws.databaseManager).BuyListing(&username, listingID, amount)

		if err != nil {
			return err
		}

		return nil

	}

	return fmt.Errorf("user is not logged in")
}

func (ws WSDBProxy) CreateProfile(username, email, password *string) error {

	if ws.databaseManager == nil {
		return fmt.Errorf("databaseManager has not been intialised")
	}

	// TODO: Add Validation to check amount makes sense

	err := (*ws.databaseManager).CreateProfile(username, email, password)

	if err != nil {
		return err
	}

	return nil
}
