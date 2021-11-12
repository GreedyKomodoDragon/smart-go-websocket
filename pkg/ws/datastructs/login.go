package ws

type Login struct {
	Email    string
	Password string
}

type LoginResult struct {
	BaseMessage
	Result   bool
	Username string
}
