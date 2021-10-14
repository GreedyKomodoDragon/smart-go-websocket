package db

import (
	"context"
	"fmt"
	"go-websocket/pkg/mocks"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/testcontainers/testcontainers-go"
)

var _ = Describe("SmartDB", func() {

	const username = "neo4j"
	const password = "s3cr3t"

	var ctx context.Context
	var neo4jContainer testcontainers.Container
	var driver neo4j.Driver
	var smartDB ISmartDBWriter

	BeforeEach(func() {
		ctx = context.Background()
		var err error
		neo4jContainer, err = mocks.StartNeo4jContainer(ctx, username, password)

		Expect(err).To(BeNil(), "Container should start")
		port, err := neo4jContainer.MappedPort(ctx, "7687")
		Expect(err).To(BeNil(), "Port should be resolved")
		address := fmt.Sprintf("bolt://localhost:%d", port.Int())
		driver, err = neo4j.NewDriver(address, neo4j.BasicAuth(username, password, ""))
		Expect(err).To(BeNil(), "Driver should be created")

	})

	AfterEach(func() {
		mocks.Close(driver, "Driver")
		Expect(neo4jContainer.Terminate(ctx)).To(BeNil(), "Container should stop")
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

		username := registerUser(smartDB)

		listing := Listing{
			Title:      "Example Listing",
			Decription: "This is a description of a listing",
			Images:     []string{"http://img.server.com/12", "http://img.server.com/13", "http://img.server.com/12"},
			Price:      12,
			Sym:        "ETH",
		}

		err := smartDB.UploadListing(username, &listing)

		Expect(err).To(BeNil(), "Transaction should successfully run")
	})

	It("Upload Listing: Unregistered Account", func() {
		session := driver.NewSession(neo4j.SessionConfig{})
		smartDB = NeoSmartHandler{
			Session: session,
		}

		defer mocks.Close(session, "Session")

		listing := Listing{
			Title:      "Example Listing",
			Decription: "This is a description of a listing",
			Images:     []string{"http://img.server.com/12", "http://img.server.com/13", "http://img.server.com/12"},
			Price:      12,
			Sym:        "ETH",
		}

		username := "some"
		err := smartDB.UploadListing(&username, &listing)

		Expect(err).NotTo(BeNil(), "Transaction should unsuccessfully run")
	})

})

func registerUser(db ISmartDBWriter) *string {
	username := "some"
	email := "some@example.com"
	initialPassword := "some-password"

	err := db.CreateProfile(&username, &email, &initialPassword)

	if err != nil {
		Panic().NegatedFailureMessage("This should not have happened (check previous tests)")
	}

	return &username
}
