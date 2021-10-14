package db

import (
	"fmt"
	cryptograph "go-websocket/pkg/Cryptograph"
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

func (db NeoSmartHandler) CreateMessage(senderID, receiverID, message *string) error {

	_, err := db.Session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			`
			MATCH (sender: Person), (reciever: Person)
			WHERE sender.id = $senderID and receiver.id = $receiverID
			CREATE (sender)-[r:message {message: $message, time: $currentTime, timeOfRead: 0}]->(receiver)
			`,
			map[string]interface{}{
				"message":     *message,
				"senderID":    *senderID,
				"receiverID":  *receiverID,
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

func (db NeoSmartHandler) UploadListing(username *string, listing *Listing) error {

	_, err := db.Session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
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
				{id: $listingID, title: $title, 
				description: $description, images: $images,
				price: $price, sym: $symbol, active: true
				})
			CREATE (seller)-[s:selling {timeOfUpload: $uploadTime}]->(listing)
			`,
			map[string]interface{}{
				"username":    *username,
				"listingID":   &listing.Id,
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

	return err
}

func (db NeoSmartHandler) BuyListing(buyerID, listingID *string, amount *int) error {

	value, err := db.Session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
		result, err := transaction.Run(
			`
			MATCH (buyer: Person {id: $buyerID}), (listing: Listing {id: listingID})
			apoc.do.when(true,
				'CREATE (buyer)-[b:brought {timeOfPurchrase: $purchaseTime}]->(listing) SET listing.active = false RETURN true',
				'RETURN false', 
				{}
			)
			YIELD value
			RETURN value
			`,
			map[string]interface{}{
				"listingID":    *listingID,
				"buyerID":      *buyerID,
				"purchaseTime": time.Now().Unix(),
			})

		if err != nil {
			return nil, err
		}

		if result.Next() {
			return result.Record().Values[0], nil
		}

		return nil, result.Err()
	})

	fmt.Println(value)

	return err
}
