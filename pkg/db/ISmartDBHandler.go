package db

type ISmartDBReader interface {
	getMessages(recieverID, sendID, lastUnreadMessageID *string) (string, error)
	getListing(listingID *string) (string, error)
	getProfile(profileID *string) (string, error)
	getUnreadNotifications(profileID *string) (string, error)
	checkLogin()
}

type ISmartDBWriter interface {

	// Writer Methods
	CreateMessage(senderID, recieverID, message *string) error
	UploadListing(username *string, listing *Listing) error
	BuyListing(buyerID, listingID *string, amount *int) error
	CreateProfile(username, email, password *string) error
}

type ISmartDBWriterReader interface {
	ISmartDBReader
	ISmartDBWriter
}
