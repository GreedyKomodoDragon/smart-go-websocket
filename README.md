# smart-go-websocket
A backend written in Go, Websocket for clients to access Databases

## Prerequisites

Make sure to install a recent [Golang](https://golang.org/) version.

## Build

As simple as:

```
go build ./cmd/conduit
```

## Run

First, make sure to configure the application to target your specific Neo4j instance.
All settings are mandatory.

| Environment variable  | Description |
| --------------------- | ----------- |
| NEO4J_URI             | [Connection URI](https://neo4j.com/docs/driver-manual/current/client-applications/#driver-connection-uris) of the instance (e.g. `bolt://localhost`, `neo4j+s://example.org`) |
| NEO4J_USERNAME        | Username of the account to connect with (must have read & write permissions) |
| NEO4J_PASSWORD        | Password of the account to connect with (must have read & write permissions)|

Then, just execute:
```
./conduit
```

You can also skip the build command and directly execute:

```
go run ./cmd/conduit/
```
