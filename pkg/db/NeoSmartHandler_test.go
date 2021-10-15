package db

import (
	"context"
	"fmt"
	"go-websocket/pkg/mocks"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SmartDB", func() {

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
		smartDB = NeoSmartHandler{
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
		smartDB = NeoSmartHandler{
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
		smartDB = NeoSmartHandler{
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
		smartDB = NeoSmartHandler{
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
		smartDB = NeoSmartHandler{
			Session: session,
		}

		defer mocks.Close(session, "Session")

		err := smartDB.CreateProfile(&username, &email, &initialPassword)

		Expect(err).To(BeNil(), "Transaction should successfully run")

		err = smartDB.CreateProfile(&usernameTwo, &emailTwo, &initialPassword)

		Expect(err).To(BeNil(), "Transaction should successfully run")
	})

	It("Upload Listing: Registered Account", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoSmartHandler{
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
		smartDB = NeoSmartHandler{
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
		smartDB = NeoSmartHandler{
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
		smartDB = NeoSmartHandler{
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

		var time uint = 0
		messages, err := smartDB.GetMessages(&usernameTwo, &username, &time)

		Expect(err).To(BeNil(), "Message should be able to be uploaded")
		Expect(len(messages)).To(Equal(1), "Message should get a single message")
		Expect(messages[0].Contents).To(Equal("The test message"), "Message should not be changed")
		Expect(messages[0].Sender).To(Equal(false), "person who accessed to the message did not send it")

	})

	It("Brought Item: Item Exists and Owned by someone else ", func() {

		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoSmartHandler{
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

		amount := 50
		err = smartDB.BuyListing(&usernameTwo, &id, &amount)

		Expect(err).To(BeNil(), "Should be able to buy listing")

	})

	It("Brought Item: Item Exists and Owned by same ", func() {

		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoSmartHandler{
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

		amount := 50
		err = smartDB.BuyListing(&username, &id, &amount)

		Expect(err).NotTo(BeNil(), "Should not be able to buy listing")
	})

	It("Brought Item: Cannot buy something that has already been brought", func() {

		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoSmartHandler{
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

		amount := 50
		err = smartDB.BuyListing(&username, &id, &amount)

		Expect(err).NotTo(BeNil(), "Should not be able to buy listing")
	})

})

func registerUser(db ISmartDBWriter, username, email, intialPassword *string) error {

	err := db.CreateProfile(username, email, intialPassword)

	return err
}
