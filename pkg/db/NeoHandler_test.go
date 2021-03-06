package db

import (
	"context"
	"fmt"
	"go-websocket/pkg/mocks"
	"time"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("NeoDB", func() {

	const username = "neo4j"
	const password = "s3cr3t"

	var ctx context.Context = context.Background()
	neo4jContainer, _ := mocks.StartNeo4jContainer(ctx, username, password)
	port, _ := neo4jContainer.MappedPort(ctx, "7687")
	address := fmt.Sprintf("bolt://localhost:%d", port.Int())
	var driver neo4j.Driver

	var smartDB ISmartDBWriterReader

	BeforeEach(func() {
		driver, _ = neo4j.NewDriver(address, neo4j.BasicAuth(username, password, ""))

		session := driver.NewSession(neo4j.SessionConfig{})
		defer mocks.Close(session, "Session")

		session.WriteTransaction(func(transaction neo4j.Transaction) (interface{}, error) {
			// Has not been brought
			_, err := transaction.Run(
				`
				MATCH (n)
				DETACH DELETE n
				`,
				map[string]interface{}{})

			// Check that transaction worked
			if err != nil {
				return nil, err
			}

			return nil, nil

		})

	})

	AfterEach(func() {
		mocks.Close(driver, "Driver")
	})

	It("Registration: Single free username and email", func() {
		username := "some-user"
		email := "some-user@example.com"
		initialPassword := "some-password"

		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoHandler{
			Session: session,
		}

		defer mocks.Close(session, "Session")

		err := smartDB.CreateProfile(&username, &email, &initialPassword)

		Expect(err).To(BeNil(), "Transaction should successfully run")
	})

	It("Registration: same account twice should result in an error", func() {
		username := "some-user"
		email := "some-user@example.com"
		initialPassword := "some-password"

		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoHandler{
			Session: session,
		}

		defer mocks.Close(session, "Session")

		smartDB.CreateProfile(&username, &email, &initialPassword)
		err := smartDB.CreateProfile(&username, &email, &initialPassword)

		Expect(err).ToNot(BeNil(), "Transaction should unsuccessfully run")
	})

	It("Registration: two account with same username but different email", func() {
		username := "some-user"
		email := "some-user@example.com"
		initialPassword := "some-password"

		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoHandler{
			Session: session,
		}

		defer mocks.Close(session, "Session")

		smartDB.CreateProfile(&username, &email, &initialPassword)

		emailNew := "some-user@examples.com"

		err := smartDB.CreateProfile(&username, &emailNew, &initialPassword)

		Expect(err).ToNot(BeNil(), "Transaction should unsuccessfully run")
	})

	It("Registration: two account with same email but different username", func() {
		username := "some-user"
		email := "some-user@example.com"
		initialPassword := "some-password"

		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoHandler{
			Session: session,
		}

		defer mocks.Close(session, "Session")

		smartDB.CreateProfile(&username, &email, &initialPassword)

		usernameNew := "another-user"

		err := smartDB.CreateProfile(&usernameNew, &email, &initialPassword)

		Expect(err).ToNot(BeNil(), "Transaction should unsuccessfully run")
	})

	It("Registration: two account with different username & email", func() {
		username := "some"
		email := "some@example.com"
		usernameTwo := "some-user"
		emailTwo := "some-user@example.com"
		initialPassword := "some-password"

		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoHandler{
			Session: session,
		}

		defer mocks.Close(session, "Session")

		err := smartDB.CreateProfile(&username, &email, &initialPassword)

		Expect(err).To(BeNil(), "Transaction should successfully run")

		err = smartDB.CreateProfile(&usernameTwo, &emailTwo, &initialPassword)

		Expect(err).To(BeNil(), "Transaction should successfully run")
	})

	It("Check Login: Same login, same password", func() {
		username := "some"
		email := "some@example.com"
		initialPassword := "some-password"

		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoHandler{
			Session: session,
		}

		defer mocks.Close(session, "Session")

		err := smartDB.CreateProfile(&username, &email, &initialPassword)

		Expect(err).To(BeNil(), "Transaction should successfully run")

		returnedUsername, err := smartDB.CheckLogin(&email, &initialPassword)

		Expect(err).To(BeNil(), "Transaction should successfully run")
		Expect(returnedUsername).To(Equal(username), "The username")
	})

	It("Check Login: known login, different password", func() {
		username := "some"
		email := "some@example.com"
		initialPassword := "some-password"

		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoHandler{
			Session: session,
		}

		defer mocks.Close(session, "Session")

		newPassword := "djnhalkdasd"

		err := smartDB.CreateProfile(&username, &email, &initialPassword)

		Expect(err).To(BeNil(), "Transaction should successfully run")

		returnUsername, err := smartDB.CheckLogin(&email, &newPassword)

		Expect(err).ToNot(BeNil(), "Transaction should not successfully run")
		Expect(returnUsername).To(Equal(""), "Username returned should be the same")
	})

	It("Check Login: unknown login, different password", func() {
		email := "some@gmail.com"
		initialPassword := "some-password"

		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoHandler{
			Session: session,
		}

		defer mocks.Close(session, "Session")

		username, err := smartDB.CheckLogin(&email, &initialPassword)

		Expect(err).NotTo(BeNil(), "Transaction should not successfully run")
		Expect(username).To(Equal(""), "Username should be empty")
	})

	It("Upload Listing: Registered Account", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoHandler{
			Session: session,
		}

		defer mocks.Close(session, "Session")

		username := "some"
		email := "some@example.com"
		initialPassword := "some-password"

		err := registerUser(smartDB, &username, &email, &initialPassword)

		if err != nil {
			Panic().NegatedFailureMessage("This should not have happened (check previous tests)")
		}

		listing := Listing{
			Title:      "Example Listing",
			Decription: "This is a description of a listing",
			Images:     []string{"http://img.server.com/12", "http://img.server.com/13", "http://img.server.com/12"},
			Price:      12,
			Sym:        "ETH",
		}

		_, err = smartDB.UploadListing(&username, &listing)

		Expect(err).To(BeNil(), "Transaction should successfully run")
	})

	It("Uploaded Listing: Can retrieve", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoHandler{
			Session: session,
		}

		defer mocks.Close(session, "Session")

		username := "some"
		email := "some@example.com"
		initialPassword := "some-password"

		err := registerUser(smartDB, &username, &email, &initialPassword)

		if err != nil {
			Panic().NegatedFailureMessage("This should not have happened (check previous tests)")
		}

		listing := Listing{
			Title:      "Example Listing",
			Decription: "This is a description of a listing",
			Images:     []string{"http://img.server.com/12", "http://img.server.com/13", "http://img.server.com/12"},
			Price:      12,
			Sym:        "ETH",
		}

		id, err := smartDB.UploadListing(&username, &listing)
		Expect(err).To(BeNil(), "Transaction should successfully run")

		listingFound, err := smartDB.GetListing(&id)
		Expect(err).To(BeNil(), "Transaction should successfully run")

		Expect(id).To(Equal(listingFound.Id), "IDs should be the same")
		Expect(listing.Title).To(Equal(listingFound.Title), "Titles should be the same")
		Expect(listing.Decription).To(Equal(listingFound.Decription), "Decription should be the same")
		Expect(listing.Images).To(Equal(listingFound.Images), "Images should be the same")
		Expect(listing.Sym).To(Equal(listingFound.Sym), "Sym should be the same")
		Expect(listingFound.Owner).To(Equal(username), "Owner should be the same as the username that created it")
	})

	It("Messages: can create a message", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoHandler{
			Session: session,
		}

		defer mocks.Close(session, "Session")

		username := "some"
		email := "some@example.com"
		usernameTwo := "some-user"
		emailTwo := "some-user@example.com"
		initialPassword := "some-password"

		err := registerUser(smartDB, &username, &email, &initialPassword)

		if err != nil {
			Panic().NegatedFailureMessage("This should not have happened (check previous tests)")
		}

		err = registerUser(smartDB, &usernameTwo, &emailTwo, &initialPassword)

		if err != nil {
			Panic().NegatedFailureMessage("This should not have happened (check previous tests)")
		}

		message := "The test message"

		err = smartDB.CreateMessage(&username, &usernameTwo, &message)

		Expect(err).To(BeNil(), "Message should be able to be uploaded")
	})

	It("Messages: can get messages", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoHandler{
			Session: session,
		}

		defer mocks.Close(session, "Session")

		username := "some"
		email := "some@example.com"
		usernameTwo := "some-user"
		emailTwo := "some-user@example.com"
		initialPassword := "some-password"

		err := registerUser(smartDB, &username, &email, &initialPassword)

		if err != nil {
			Panic().NegatedFailureMessage("This should not have happened (check previous tests)")
		}

		err = registerUser(smartDB, &usernameTwo, &emailTwo, &initialPassword)

		if err != nil {
			Panic().NegatedFailureMessage("This should not have happened (check previous tests)")
		}

		message := "The test message"

		err = smartDB.CreateMessage(&username, &usernameTwo, &message)

		if err != nil {
			Panic().NegatedFailureMessage("This should not have happened (check previous tests)")
		}

		var time int64 = 0
		messages, err := smartDB.GetMessages(&usernameTwo, &username, &time)

		Expect(err).To(BeNil(), "Message should be able to be uploaded")
		Expect(len(messages)).To(Equal(1), "Message should get a single message")
		Expect(messages[0].Contents).To(Equal("The test message"), "Message should not be changed")
		Expect(messages[0].Sender).To(Equal(false), "person who accessed to the message did not send it")

	})

	It("Message: Can get messages after a certain date", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoHandler{
			Session: session,
		}

		defer mocks.Close(session, "Session")

		username := "some"
		email := "some@example.com"
		usernameTwo := "some-user"
		emailTwo := "some-user@example.com"
		initialPassword := "some-password"

		err := registerUser(smartDB, &username, &email, &initialPassword)

		if err != nil {
			Panic().NegatedFailureMessage("This should not have happened (check previous tests)")
		}

		err = registerUser(smartDB, &usernameTwo, &emailTwo, &initialPassword)

		if err != nil {
			Panic().NegatedFailureMessage("This should not have happened (check previous tests)")
		}

		message := "Message One"

		err = smartDB.CreateMessage(&username, &usernameTwo, &message)

		if err != nil {
			Panic().NegatedFailureMessage("This should not have happened (check previous tests)")
		}

		//So messages don't have the same time
		time.Sleep(time.Second)

		message = "Message Two"
		err = smartDB.CreateMessage(&username, &usernameTwo, &message)

		if err != nil {
			Panic().NegatedFailureMessage("This should not have happened (check previous tests)")
		}

		var timeOfLastMessage int64 = 0
		messages, err := smartDB.GetMessages(&usernameTwo, &username, &timeOfLastMessage)

		Expect(err).To(BeNil(), "Message should be able to be uploaded")
		Expect(len(messages)).To(Equal(2), "Should have two messages in the chat log")

		if err != nil {
			Panic().NegatedFailureMessage("This should not have happened (check previous tests)")
		}

		messages, err = smartDB.GetMessages(&usernameTwo, &username, &messages[0].Time)

		Expect(err).To(BeNil(), "Message should be able to be uploaded")
		Expect(len(messages)).To(Equal(1), "Message should get a single message")
		Expect(messages[0].Contents).To(Equal("Message Two"), "Message should not be changed")
		Expect(messages[0].Sender).To(Equal(false), "person who accessed to the message did not send it")

	})

	It("Brought Item: Item Exists and Owned by someone else ", func() {

		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoHandler{
			Session: session,
		}

		defer mocks.Close(session, "Session")

		username := "some"
		email := "some@example.com"

		usernameTwo := "some-user"
		emailTwo := "some-user@example.com"

		initialPassword := "some-password"

		err := registerUser(smartDB, &username, &email, &initialPassword)

		if err != nil {
			Panic().NegatedFailureMessage("This should not have happened (check previous tests)")
		}

		err = registerUser(smartDB, &usernameTwo, &emailTwo, &initialPassword)

		if err != nil {
			Panic().NegatedFailureMessage("This should not have happened (check previous tests)")
		}

		listing := Listing{
			Title:      "Example Listing",
			Decription: "This is a description of a listing",
			Images:     []string{"http://img.server.com/12", "http://img.server.com/13", "http://img.server.com/12"},
			Price:      12,
			Sym:        "ETH",
		}

		id, err := smartDB.UploadListing(&username, &listing)
		Expect(err).To(BeNil(), "Should be able to upload listing")

		var amount int64 = 50 //int64 to remove any int32 or int64 uncertainty for different systems
		err = smartDB.BuyListing(&usernameTwo, &id, &amount)

		Expect(err).To(BeNil(), "Should be able to buy listing")

	})

	It("Brought Item: Item Exists and Owned by same ", func() {

		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoHandler{
			Session: session,
		}

		defer mocks.Close(session, "Session")

		username := "some"
		email := "some@example.com"

		initialPassword := "some-password"

		err := registerUser(smartDB, &username, &email, &initialPassword)

		if err != nil {
			Panic().NegatedFailureMessage("This should not have happened (check previous tests)")
		}

		listing := Listing{
			Title:      "Example Listing",
			Decription: "This is a description of a listing",
			Images:     []string{"http://img.server.com/12", "http://img.server.com/13", "http://img.server.com/12"},
			Price:      12,
			Sym:        "ETH",
		}

		id, err := smartDB.UploadListing(&username, &listing)
		Expect(err).To(BeNil(), "Should be able to upload listing")

		var amount int64 = 50
		err = smartDB.BuyListing(&username, &id, &amount)

		Expect(err).NotTo(BeNil(), "Should not be able to buy listing")
	})

	It("Brought Item: Cannot buy something that has already been brought", func() {

		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoHandler{
			Session: session,
		}

		defer mocks.Close(session, "Session")

		username := "some"
		email := "some@example.com"
		usernameAnother := "another"
		emailAnother := "another@example.com"

		initialPassword := "some-password"

		err := registerUser(smartDB, &username, &email, &initialPassword)

		if err != nil {
			Panic().NegatedFailureMessage("This should not have happened (check previous tests)")
		}

		err = registerUser(smartDB, &usernameAnother, &emailAnother, &initialPassword)

		if err != nil {
			Panic().NegatedFailureMessage("This should not have happened (check previous tests)")
		}

		listing := Listing{
			Title:      "Example Listing",
			Decription: "This is a description of a listing",
			Images:     []string{"http://img.server.com/12", "http://img.server.com/13", "http://img.server.com/12"},
			Price:      12,
			Sym:        "ETH",
		}

		id, err := smartDB.UploadListing(&username, &listing)
		Expect(err).To(BeNil(), "Should be able to upload listing")

		var amount int64 = 50
		err = smartDB.BuyListing(&usernameAnother, &id, &amount)

		Expect(err).To(BeNil(), "Should be able to buy the first time")
		err = smartDB.BuyListing(&usernameAnother, &id, &amount)

		Expect(err).NotTo(BeNil(), "Should be able to buy the second time")
	})

	It("Contacts: Can obtain contacts", func() {

		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoHandler{
			Session: session,
		}

		defer mocks.Close(session, "Session")

		username := "some"
		email := "some@example.com"
		usernameAnother := "another"
		emailAnother := "another@example.com"

		initialPassword := "some-password"

		_ = registerUser(smartDB, &username, &email, &initialPassword)
		_ = registerUser(smartDB, &usernameAnother, &emailAnother, &initialPassword)

		message := "This is an example"

		contacts, err := smartDB.GetContacts(&username)
		Expect(err).To(BeNil(), "Should be able to get contacts")
		Expect(len(contacts)).To(Equal(0), "Should be no contacts")

		err = smartDB.CreateMessage(&username, &usernameAnother, &message)
		Expect(err).To(BeNil(), "Should send a message")

		contacts, err = smartDB.GetContacts(&username)
		Expect(err).To(BeNil(), "Should be able to get contacts")
		Expect(len(contacts)).To(Equal(1), "Should only be one contact")
		Expect(contacts[0].Username).To(Equal(usernameAnother), "Should only be one contact")

	})

})

func registerUser(db ISmartDBWriter, username, email, intialPassword *string) error {

	err := db.CreateProfile(username, email, intialPassword)

	return err
}
