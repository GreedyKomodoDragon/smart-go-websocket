# Websocket Protocol Table

This websocket should use a secure connection as it will be senting sensitive information

Client -> Server
|Command   |Description   | JSON Data   | Server Emits  |
|---|---|---|---|
|"login"|This is used to allow the client to login in|email:string <br/> password:string  | "loginResult" |
|"logout"|Used to log the address out of the websocket|N/A|N/A|
|"registration|Allows a user to register an account|username:string <br/> email:string <br/> password:string <br/>|"regResult"|

Server -> Client
|Command   |Description   | JSON Data   | Client Emits  |
|---|---|---|---|
|"loginResult"|Used to tell the client how login information resulted |Result:bool |N/A|
|"regResult"|Sends registration result|Message:string|N/A|