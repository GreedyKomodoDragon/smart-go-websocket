package db

type IDBHandler interface {
	establishConnection(url string, username string, password string) (string, error)
}
