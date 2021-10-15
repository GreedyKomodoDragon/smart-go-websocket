package db

type Listing struct {
	Id         int64
	Title      string
	Decription string
	Images     []string
	Price      int64  // if token is decimal it will be moved up to be an integer
	Sym        string // which token is being used
	Active     bool
	Owner      string
}
