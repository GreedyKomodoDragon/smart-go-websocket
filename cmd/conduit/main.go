package main

import (
	"flag"
	"go-websocket/pkg/db"
	"go-websocket/pkg/ws"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

var addr = flag.String("addr", ":5000", "http service address")

func main() {
	flag.Parse()

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

	hub := ws.NewHub()
	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWs(hub, w, r, &dbProxy)
	})

	err = http.ListenAndServe(*addr, nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
