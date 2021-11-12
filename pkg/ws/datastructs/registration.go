package ws

type Registration struct {
	Username string
	Email    string
	Password string
}

type RegistrationResult struct {
	BaseMessage
	Message string
}
