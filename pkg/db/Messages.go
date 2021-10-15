package db

type Messages struct {
	Contents string
	Time     int64
	Read     int64
	Sender   bool
}
