package main

import (
	"fmt"
	"go-websocket/pkg/db"
	"os"

	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err.Error())
	}

	url := os.Getenv("NEO4J_URI")
	username := os.Getenv("NEO4J_USERNAME")
	password := os.Getenv("NEO4J_PASSWORD")

	driver, err := neo4j.NewDriver(url, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return
	}

	defer driver.Close()

	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	var smartDB db.ISmartDBWriter = db.NeoSmartHandler{
		Session: session,
	}

	email := "jksdshdhksdlddsasdasd"

	err = smartDB.CreateProfile(&username, &email, &password)
	fmt.Println(err)
}
