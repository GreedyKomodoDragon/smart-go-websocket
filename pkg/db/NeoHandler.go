package db

import (
	"fmt"
	cryptograph "go-websocket/pkg/Cryptograph"
	"strings"
	"time"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type NeoSmartHandler struct {
	Session neo4j.Session
}

func (db NeoSmartHandler) CreateProfile(username, email, password *string) error {

	_, err := db.Session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {

		// Check if account exists for email submitted
		result, err := transaction.Run(
			`
			MATCH (n:Person)
			WHERE n.email = $email 
			RETURN COUNT(n)
			`,
			map[string]interface{}{
				"email": *email,
			})

		// Check that transaction worked
		if err != nil {
			return nil, err
		}

		if !result.Next() {
			return nil, fmt.Errorf("unexpected error: neo4j account could not be counted")
		}

		countEmail, okEmail := result.Record().Values[0].(int64) // count email

		if !okEmail {
			return nil, fmt.Errorf("unexpected error: neo4j account could not be counted")
		}

		if countEmail != 0 {
			return nil, fmt.Errorf("cannot create account with that email")
		}

		// Check if account exists
		result, err = transaction.Run(
			`
			MATCH (n:Person)
			WHERE n.username = $username
			RETURN COUNT(n)
			`,
			map[string]interface{}{
				"username": *username,
				"email":    *email,
			})

		// Check that transaction worked
		if err != nil {
			return nil, err
		}

		if !result.Next() {
			return nil, fmt.Errorf("unexpected error: neo4j account could not be counted")
		}

		countUsername, okUsername := result.Record().Values[0].(int64) // count username

		if !okUsername {
			return nil, fmt.Errorf("unexpected error: neo4j account could not be counted")
		}

		if countUsername != 0 {
			return nil, fmt.Errorf("canot create account that username")
		}

		// Create the account if no account exists
		salt := cryptograph.GenerateRandomSalt(32)
		passwordStr, saltStr := cryptograph.HashPassword(password, &salt)

		result, err = transaction.Run(
			`
			CREATE (n:Person {username: $username, email: $email, password: $password, salt: $salt})
			`,
			map[string]interface{}{
				"username":    *username,
				"email":       *email,
				"password":    passwordStr,
				"salt":        saltStr,
				"currentTime": time.Now().Unix(),
			})

		if err != nil {
			return nil, err
		}

		return nil, result.Err()
	})

	return err
}

func (db NeoSmartHandler) CreateMessage(senderUsername, receiverUsername, message *string) error {

	_, err := db.Session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			`
			MATCH (sender: Person {username: $usernameOne}), (reciever: Person {username: $usernameTwo})
			CREATE (sender)-[r:message {message: $message, time: $currentTime, timeOfRead: 0}]->(reciever)
			`,
			map[string]interface{}{
				"message":     *message,
				"usernameOne": *senderUsername,
				"usernameTwo": *receiverUsername,
				"currentTime": time.Now().Unix(),
			})

		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record().Values[0], nil
		}

		return nil, result.Err()
	})

	return err
}

func (db NeoSmartHandler) UploadListing(username *string, listing *Listing) (int64, error) {

	value, err := db.Session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			`
			MATCH (seller: Person)
			WHERE seller.username = $username
			RETURN COUNT(seller)
			`,
			map[string]interface{}{
				"username": *username,
			})

		if err != nil {
			return nil, err
		}

		if !result.Next() {
			return nil, fmt.Errorf("unexpected error: neo4j account could not be counted")
		}

		countUsername, okUsername := result.Record().Values[0].(int64) // count username

		if !okUsername {
			return nil, fmt.Errorf("unexpected error: neo4j account could not be counted")
		}

		if countUsername == 0 {
			return nil, fmt.Errorf("account does not exist")
		}

		result, err = transaction.Run(
			`
			MATCH (seller: Person)
			WHERE seller.username = $username
			CREATE (listing: Listing 
				{title: $title, 
				description: $description, images: $images,
				price: $price, sym: $symbol, active: true
				})
			CREATE (seller)-[s:Selling {timeOfUpload: $uploadTime}]->(listing)
			RETURN id(listing)
			`,
			map[string]interface{}{
				"username":    *username,
				"title":       &listing.Title,
				"description": &listing.Decription,
				"images":      &listing.Images,
				"price":       &listing.Price,
				"symbol":      &listing.Sym,
				"uploadTime":  time.Now().Unix(),
			})

		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record().Values[0], nil
		}

		return nil, result.Err()
	})

	if err != nil {
		return -1, err
	}

	// Cast to string
	if i, ok := value.(int64); ok {
		return i, nil
	}

	return -1, nil
}

func (db NeoSmartHandler) BuyListing(buyerUsername *string, listingID *int64, amount *int64) error {

	_, err := db.Session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		// Has not been brought or no longer for sale
		result, err := transaction.Run(
			`
			MATCH (listing: Listing)
			WHERE id(listing) = $listingID AND listing.active = false
			RETURN COUNT(listing)
			`,
			map[string]interface{}{
				"listingID": *listingID,
			})

		// Check that transaction worked
		if err != nil {
			return nil, err
		}

		if !result.Next() {
			return nil, fmt.Errorf("unexpected error: neo4j relationship could not be checked")
		}

		count, ok := result.Record().Values[0].(int64)

		if !ok {
			return nil, fmt.Errorf("unexpected error: neo4j relationship could not be checked")
		}

		if count > 0 {
			return nil, fmt.Errorf("cannot buy an item that has already been brought")
		}

		// Is not the owner of the item
		result, err = transaction.Run(
			`
			MATCH (listing: Listing)
			WHERE id(listing) = $listingID
			RETURN EXISTS ((: Person {username: $buyerUsername})-[:Selling]->(listing))
			`,
			map[string]interface{}{
				"listingID":     *listingID,
				"buyerUsername": *buyerUsername,
				"purchaseTime":  time.Now().Unix(),
			})

		// Check that transaction worked
		if err != nil {
			return nil, err
		}

		if !result.Next() {
			return nil, fmt.Errorf("unexpected error: neo4j relationship could not be checked")
		}

		isOwner, ok := result.Record().Values[0].(bool)

		if !ok {
			return nil, fmt.Errorf("unexpected error: neo4j relationship could not be checked")
		}

		if isOwner {
			return nil, fmt.Errorf("cannot buy your own item")
		}

		result, err = transaction.Run(
			`
			MATCH (buyer: Person {username: $buyerUsername}), (listing: Listing)
			WHERE id(listing) = $listingID
			CREATE (buyer)-[b:brought {timeOfPurchrase: $purchaseTime}]->(listing) 
			SET listing.active = false
			`,
			map[string]interface{}{
				"listingID":     *listingID,
				"buyerUsername": *buyerUsername,
				"purchaseTime":  time.Now().Unix(),
			})

		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record().Values[0], nil
		}

		return nil, result.Err()
	})

	return err
}

// Should not be able to call if the thread that calls is not the same as recieverUsername
func (db NeoSmartHandler) GetMessages(recieverUsername, senderUsername *string, timeOfLastMessage *int64) ([]Messages, error) {
	value, err := db.Session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		var result neo4j.Result
		var err error

		// check if there is a time point to start from
		if *timeOfLastMessage > 0 {
			result, err = transaction.Run(
				`
			MATCH (one: Person {username: $usernameOne})-[m:message]-(two: Person {username: $usernameTwo})
			WHERE m.time > $time
			RETURN m.message, m.time, m.timeOfRead, (startNode(m) = one)
			ORDER BY m.time
			LIMIT 10
			`,
				map[string]interface{}{
					"usernameOne": *recieverUsername,
					"usernameTwo": *senderUsername,
					"time":        *timeOfLastMessage,
				})

		} else {
			result, err = transaction.Run(
				`
			MATCH (one: Person {username: $usernameOne})-[m:message]-(two: Person {username: $usernameTwo})
			RETURN m.message, m.time, m.timeOfRead, (startNode(m) = one)
			ORDER BY m.time
			LIMIT 10
			`,
				map[string]interface{}{
					"usernameOne": *recieverUsername,
					"usernameTwo": *senderUsername,
				})
		}

		// Check that transaction worked
		if err != nil {
			return nil, err
		}

		//Get all the messages you have access to
		messages := []Messages{}

		for result.Next() {
			messages = append(messages, Messages{
				Contents: result.Record().Values[0].(string),
				Time:     result.Record().Values[1].(int64),
				Read:     result.Record().Values[2].(int64),
				Sender:   result.Record().Values[3].(bool),
			})
		}

		return messages, nil

	})

	if err != nil {
		return nil, err
	}

	// Cast to []Messages
	messages, ok := value.([]Messages)

	if !ok {
		return nil, fmt.Errorf("cannot cast to []Messages")
	}

	return messages, nil
}

func (db NeoSmartHandler) GetListing(listingID *int64) (Listing, error) {

	value, err := db.Session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {

		result, err := transaction.Run(
			`
			MATCH (p:Person)-[s:Selling]->(item: Listing)
			WHERE id(item) = $listingID
			RETURN item.title, item.description, item.images, item.price, item.sym, item.active, p.username
			`,
			map[string]interface{}{
				"listingID": *listingID,
			})

		// Check that transaction worked
		if err != nil {
			return nil, err
		}

		// Listing my not exist, not an error though
		if !result.Next() {
			return Listing{}, nil
		}

		// Get listing info
		images := strings.Fields(strings.Trim(fmt.Sprint(result.Record().Values[2]), "[ ]"))

		listing := Listing{
			Id:         *listingID,
			Title:      result.Record().Values[0].(string),
			Decription: result.Record().Values[1].(string),
			Images:     images,
			Price:      result.Record().Values[3].(int64),
			Sym:        result.Record().Values[4].(string),
			Active:     result.Record().Values[5].(bool),
			Owner:      result.Record().Values[6].(string),
		}

		return listing, nil

	})

	if err != nil {
		return Listing{}, err
	}

	// Cast to Listing
	if listing, ok := value.(Listing); ok {
		return listing, nil
	}

	return Listing{}, fmt.Errorf("cannot cast to listing")

}

func (db NeoSmartHandler) CheckLogin(username, password *string) (bool, error) {

	value, err := db.Session.ReadTransaction(func(transaction neo4j.Transaction) (interface{}, error) {

		// Get Salt and Password for someone with the same username
		result, err := transaction.Run(
			`
			MATCH (n:Person {username: $username})
			WHERE n.username = $username 
			RETURN n.salt, n.password
			`,
			map[string]interface{}{
				"username": *username,
			})

		// Check that transaction worked
		if err != nil {
			return false, err
		}

		if !result.Next() {
			return false, fmt.Errorf("unexpected error: neo4j account could not be found")
		}

		salt, _ := result.Record().Values[0].(string)
		passwordDB, _ := result.Record().Values[1].(string)

		match, err := cryptograph.ComparePassword(password, &passwordDB, &salt)

		if err != nil {
			return false, err
		}

		return match, nil
	})

	// Fails if it does not work
	if err != nil {
		return false, err
	}

	if match, ok := value.(bool); ok {
		return match, nil
	}

	return false, fmt.Errorf("could not case to bool")
}
