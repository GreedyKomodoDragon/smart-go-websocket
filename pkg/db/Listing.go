package db

type Listing struct {
	Id         string
	Title      string
	Decription string
	Images     []string
	Price      int    // if token is decimal it will be moved up to be an integer
	Sym        string // which token is being used
}
