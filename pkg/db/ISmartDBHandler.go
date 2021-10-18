package db

type ISmartDBReader interface {
	GetMessages(recieverUsername, senderUsername *string, timeOfLastMessage *int64) ([]Messages, error)
	GetListing(listingID *int64) (Listing, error)
	CheckLogin(username, password *string) (bool, error)
	//GetProfile(profileUsername *string) (string, error)
	//GetUnreadNotifications(profileID *string) (string, error)
}

type ISmartDBWriter interface {

	// Writer Methods
	CreateMessage(senderUsername, receiverUsername, message *string) error
	UploadListing(username *string, listing *Listing) (int64, error)
	BuyListing(buyerID *string, listingID *int64, amount *int) error
	CreateProfile(username, email, password *string) error
}

type ISmartDBWriterReader interface {
	ISmartDBReader
	ISmartDBWriter
}
