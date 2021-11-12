package main

import (
	"go-websocket/pkg/db"
	"go-websocket/pkg/ws"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/rs/cors"
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

	var smartDB db.ISmartDBWriterReader = db.NeoHandler{
		Session: session,
	}

	var dbProxy ws.WebDataProxy = ws.WSDBProxy{
		DatabaseManager: &smartDB,
		IdToUsername:    make(map[string]string),
	}

	mux := http.NewServeMux()
	ws.AssignRoutes(&dbProxy, mux)

	// provide default cors to the mux
	handler := cors.Default().Handler(mux)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})

	// decorate existing handler with cors functionality set in c
	handler = c.Handler(handler)

	log.Println("Serving at localhost:5000...")
	log.Fatal(http.ListenAndServe(":5000", handler))
}
