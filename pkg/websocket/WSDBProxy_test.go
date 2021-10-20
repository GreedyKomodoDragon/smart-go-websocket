package websocket

import (
	"context"
	"fmt"
	"go-websocket/pkg/db"
	"go-websocket/pkg/mocks"
	"time"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("WSDB", func() {

	const username = "neo4j"
	const password = "s3cr3t"

	// Account details
	accountUsernameOne := "some-user"
	accountUsernameTwo := "some"
	emailOne := "some-user@example.com"
	emailTwo := "some@example.com"
	initialPasswordOne := "some-password"
	initialPasswordTwo := "another-password"

	// Product information
	var productIDOne int64
	var productIDTwo int64

	// Neo4J Container
	var ctx context.Context = context.Background()
	neo4jContainer, _ := mocks.StartNeo4jContainer(ctx, username, password)
	port, _ := neo4jContainer.MappedPort(ctx, "7687")
	address := fmt.Sprintf("bolt://localhost:%d", port.Int())
	var driver neo4j.Driver

	// Interfaces
	var smartDB db.ISmartDBWriterReader
	var bridge webDataBridge

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

		smartDB = db.NeoSmartHandler{
			Session: session,
		}

		// Add accounts
		err := smartDB.CreateProfile(&accountUsernameOne, &emailOne, &initialPasswordOne)
		Expect(err).To(BeNil(), "Transaction should successfully run")

		err = smartDB.CreateProfile(&accountUsernameTwo, &emailTwo, &initialPasswordTwo)
		Expect(err).To(BeNil(), "Transaction should successfully run")

		// Add messages
		message := "first message"
		secondMessage := "second message"
		thirdMessage := "third message"

		err = smartDB.CreateMessage(&accountUsernameOne, &accountUsernameTwo, &message)
		Expect(err).To(BeNil(), "Transaction should successfully run")
		time.Sleep(time.Second)

		err = smartDB.CreateMessage(&accountUsernameOne, &accountUsernameTwo, &secondMessage)
		Expect(err).To(BeNil(), "Transaction should successfully run")
		time.Sleep(time.Second)

		err = smartDB.CreateMessage(&accountUsernameTwo, &accountUsernameOne, &thirdMessage)
		Expect(err).To(BeNil(), "Transaction should successfully run")
		time.Sleep(time.Second)

		// Add products
		productOne := db.Listing{
			Title:      "Example Listing",
			Decription: "This is a description of a listing",
			Images:     []string{"http://img.server.com/12", "http://img.server.com/13", "http://img.server.com/12"},
			Price:      12,
			Sym:        "ETH",
		}

		productIDOne, err = smartDB.UploadListing(&accountUsernameOne, &productOne)
		Expect(err).To(BeNil(), "Should be able to upload listing")

		productIDTwo, err = smartDB.UploadListing(&accountUsernameOne, &productOne)
		Expect(err).To(BeNil(), "Should be able to upload listing")

		// Buy one product, leave one free to the brought
		var amount int64 = 50
		err = smartDB.BuyListing(&accountUsernameTwo, &productIDOne, &amount)
		Expect(err).To(BeNil(), "Should be able to buy listing")

	})

	AfterEach(func() {
		mocks.Close(driver, "Driver")
	})

	It("SocketID to login: socket is not logged in", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		defer mocks.Close(session, "Session")
		smartDB = db.NeoSmartHandler{
			Session: session,
		}

		bridge = WSDBProxy{
			databaseManager: &smartDB,
			idToUsername:    make(map[string]string),
		}

		isLoggedIn := bridge.IsLoggedIn("socketOne")
		Expect(isLoggedIn).To(Equal(false), "Socket should not have a username")
	})

	It("SocketID to login: socket is logged in", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		defer mocks.Close(session, "Session")
		smartDB = db.NeoSmartHandler{
			Session: session,
		}

		bridge = WSDBProxy{
			databaseManager: &smartDB,
			idToUsername:    make(map[string]string),
		}

		err := bridge.ConnectUsernameToID(&accountUsernameOne, "socketOne")
		Expect(err).To(BeNil(), "Socket should not having a username")

		isLoggedIn := bridge.IsLoggedIn("socketOne")
		Expect(isLoggedIn).To(Equal(true), "Socket should have username")
	})

	It("Log out: ID should be disconnected to username", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		defer mocks.Close(session, "Session")
		smartDB = db.NeoSmartHandler{
			Session: session,
		}

		bridge = WSDBProxy{
			databaseManager: &smartDB,
			idToUsername:    make(map[string]string),
		}

		err := bridge.ConnectUsernameToID(&accountUsernameOne, "socketOne")
		Expect(err).To(BeNil(), "Socket should be able to connect to the username")

		isLoggedIn := bridge.IsLoggedIn("socketOne")
		Expect(isLoggedIn).To(Equal(true), "Socket should be connected username")

		err = bridge.LogoutID("socketOne")
		Expect(err).To(BeNil(), "Should be able to be logged out")

		isLoggedIn = bridge.IsLoggedIn("socketOne")
		Expect(isLoggedIn).To(Equal(false), "Socket should not have a username")
	})

	It("Log out: Cannot log out if you are not logged in", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		defer mocks.Close(session, "Session")
		smartDB = db.NeoSmartHandler{
			Session: session,
		}

		bridge = WSDBProxy{
			databaseManager: &smartDB,
			idToUsername:    make(map[string]string),
		}

		isLoggedIn := bridge.IsLoggedIn("socketOne")
		Expect(isLoggedIn).To(Equal(false), "Socket should not be connected username")

		err := bridge.LogoutID("socketOne")
		Expect(err).NotTo(BeNil(), "Should be unable to be logged out")

	})

	It("Messages: Cannot get messages if not logged in", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		defer mocks.Close(session, "Session")
		smartDB = db.NeoSmartHandler{
			Session: session,
		}

		bridge = WSDBProxy{
			databaseManager: &smartDB,
			idToUsername:    make(map[string]string),
		}

		isLoggedIn := bridge.IsLoggedIn("socketOne")
		Expect(isLoggedIn).To(Equal(false), "Socket should not be connected username")

		var time int64 = 0
		_, err := bridge.GetMessages("socketOne", &accountUsernameTwo, &time)
		Expect(err).NotTo(BeNil(), "Should be unable to be get messages if not logged in")
	})

	It("Messages: Can get messages if logged in", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		defer mocks.Close(session, "Session")
		smartDB = db.NeoSmartHandler{
			Session: session,
		}

		bridge = WSDBProxy{
			databaseManager: &smartDB,
			idToUsername:    make(map[string]string),
		}

		err := bridge.ConnectUsernameToID(&accountUsernameOne, "socketOne")
		Expect(err).To(BeNil(), "Socket should not having a username")

		isLoggedIn := bridge.IsLoggedIn("socketOne")
		Expect(isLoggedIn).To(Equal(true), "Socket should be connected username")

		var time int64 = 0
		messages, err := bridge.GetMessages("socketOne", &accountUsernameTwo, &time)

		Expect(err).To(BeNil(), "Should be able to be get messages")
		Expect(len(messages)).To(Equal(3), "3 messages should be found")
	})

	It("Messages: Can create messages if logged in", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		defer mocks.Close(session, "Session")
		smartDB = db.NeoSmartHandler{
			Session: session,
		}

		bridge = WSDBProxy{
			databaseManager: &smartDB,
			idToUsername:    make(map[string]string),
		}

		err := bridge.ConnectUsernameToID(&accountUsernameOne, "socketOne")
		Expect(err).To(BeNil(), "Socket should not have already username")

		isLoggedIn := bridge.IsLoggedIn("socketOne")
		Expect(isLoggedIn).To(Equal(true), "Socket should be connected username")

		message := "new message"

		err = bridge.CreateMessage("socketOne", &accountUsernameTwo, &message)
		Expect(err).To(BeNil(), "Should be able to be create a new message")

		var time int64 = 0
		messages, err := bridge.GetMessages("socketOne", &accountUsernameTwo, &time)

		Expect(err).To(BeNil(), "Should be able to be get messages")
		Expect(len(messages)).To(Equal(4), "4 messages should be found")
	})

	It("Messages: Cannot create messages if not logged in", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		defer mocks.Close(session, "Session")
		smartDB = db.NeoSmartHandler{
			Session: session,
		}

		bridge = WSDBProxy{
			databaseManager: &smartDB,
			idToUsername:    make(map[string]string),
		}

		isLoggedIn := bridge.IsLoggedIn("socketOne")
		Expect(isLoggedIn).To(Equal(false), "Socket should be connected username")

		message := "new message"

		err := bridge.CreateMessage("socketOne", &accountUsernameTwo, &message)
		Expect(err).NotTo(BeNil(), "Should be unable to be create a new message")

	})

	It("Listing: Can upload listing if logged in", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		defer mocks.Close(session, "Session")
		smartDB = db.NeoSmartHandler{
			Session: session,
		}

		bridge = WSDBProxy{
			databaseManager: &smartDB,
			idToUsername:    make(map[string]string),
		}

		err := bridge.ConnectUsernameToID(&accountUsernameOne, "socketOne")
		Expect(err).To(BeNil(), "Socket should not have already username")

		isLoggedIn := bridge.IsLoggedIn("socketOne")
		Expect(isLoggedIn).To(Equal(true), "Socket should be connected username")

		product := db.Listing{
			Title:      "Example Listing",
			Decription: "This is a description of a listing",
			Images:     []string{"http://img.server.com/12", "http://img.server.com/13", "http://img.server.com/12"},
			Price:      12,
			Sym:        "ETH",
		}

		err = bridge.UploadListing("socketOne", &product)
		Expect(err).To(BeNil(), "Should be able to upload a listing")

	})

	It("Listing: Cannot upload listing if not logged in", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		defer mocks.Close(session, "Session")

		smartDB = db.NeoSmartHandler{
			Session: session,
		}

		bridge = WSDBProxy{
			databaseManager: &smartDB,
			idToUsername:    make(map[string]string),
		}

		isLoggedIn := bridge.IsLoggedIn("socketOne")
		Expect(isLoggedIn).To(Equal(false), "Socket should not be connected username")

		product := db.Listing{
			Title:      "Example Listing",
			Decription: "This is a description of a listing",
			Images:     []string{"http://img.server.com/12", "http://img.server.com/13", "http://img.server.com/12"},
			Price:      12,
			Sym:        "ETH",
		}

		err := bridge.UploadListing("socketOne", &product)
		Expect(err).NotTo(BeNil(), "Should not be able to upload a listing")

	})

	It("Buying: Cannot buy a for-sale item if not logged in", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		defer mocks.Close(session, "Session")

		smartDB = db.NeoSmartHandler{
			Session: session,
		}

		bridge = WSDBProxy{
			databaseManager: &smartDB,
			idToUsername:    make(map[string]string),
		}

		isLoggedIn := bridge.IsLoggedIn("socketOne")

		Expect(isLoggedIn).To(Equal(false), "Socket should not be connected username")

		var amount int64 = 10

		err := bridge.BuyListing("socketOne", &productIDTwo, &amount)
		Expect(err).NotTo(BeNil(), "Should not be able to buy if not logged in")

	})

	It("Buying: Cannot buy for-sale item if logged in but your own", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		defer mocks.Close(session, "Session")

		smartDB = db.NeoSmartHandler{
			Session: session,
		}

		bridge = WSDBProxy{
			databaseManager: &smartDB,
			idToUsername:    make(map[string]string),
		}

		err := bridge.ConnectUsernameToID(&accountUsernameOne, "socketOne")
		Expect(err).To(BeNil(), "Socket should not having a username")

		isLoggedIn := bridge.IsLoggedIn("socketOne")
		Expect(isLoggedIn).To(Equal(true), "Socket should be connected username")

		var amount int64 = 10

		err = bridge.BuyListing("socketOne", &productIDTwo, &amount)
		Expect(err).NotTo(BeNil(), "Should be able to buy if logged in")

	})

	It("Buying: Cannot buy not-for-sale item", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		defer mocks.Close(session, "Session")

		smartDB = db.NeoSmartHandler{
			Session: session,
		}

		bridge = WSDBProxy{
			databaseManager: &smartDB,
			idToUsername:    make(map[string]string),
		}

		err := bridge.ConnectUsernameToID(&accountUsernameTwo, "socketTwo")
		Expect(err).To(BeNil(), "Socket should not having a username")

		isLoggedIn := bridge.IsLoggedIn("socketTwo")
		Expect(isLoggedIn).To(Equal(true), "Socket should be connected to username")

		var amount int64 = 10

		err = bridge.BuyListing("socketTwo", &productIDOne, &amount)
		Expect(err).NotTo(BeNil(), "Should not be able to buy already brought")

	})

	It("Buying: Can be an item that is for sale", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		defer mocks.Close(session, "Session")

		smartDB = db.NeoSmartHandler{
			Session: session,
		}

		bridge = WSDBProxy{
			databaseManager: &smartDB,
			idToUsername:    make(map[string]string),
		}

		err := bridge.ConnectUsernameToID(&accountUsernameTwo, "socketTwo")
		Expect(err).To(BeNil(), "Socket should not having a username")

		isLoggedIn := bridge.IsLoggedIn("socketTwo")
		Expect(isLoggedIn).To(Equal(true), "Socket should be connected to username")

		var amount int64 = 10

		err = bridge.BuyListing("socketTwo", &productIDTwo, &amount)
		Expect(err).To(BeNil(), "Should not be able to buy already brought")

	})

	It("Buying: Can be an item that is for sale", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		defer mocks.Close(session, "Session")

		smartDB = db.NeoSmartHandler{
			Session: session,
		}

		bridge = WSDBProxy{
			databaseManager: &smartDB,
			idToUsername:    make(map[string]string),
		}

		newEmail := "test@gmail.com"
		newUsername := "username"
		newPassword := "adhkashdlahd"

		err := bridge.CreateProfile(&newUsername, &newEmail, &newPassword)
		Expect(err).To(BeNil(), "Should be able to create an account")
	})

})
