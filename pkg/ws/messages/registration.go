package ws

type Registration struct {
	Username string
	Email    string
	Password string
}

type RegistrationResult struct {
	BaseMessage
	ResponseCode byte
}
