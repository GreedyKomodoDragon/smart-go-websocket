package cryptograph

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	memory  = 64 * 1024
	threads = 4
	time    = 1
	keyLen  = 32
)

func GenerateRandomSalt(saltSize int) []byte {
	var salt = make([]byte, saltSize)

	_, err := rand.Read(salt[:])

	if err != nil {
		panic(err)
	}

	return salt
}

func HashPassword(password *string, salt *[]byte) (string, string) {
	// Convert password string to byte slice
	var passwordBytes = []byte(*password)

	var hashedPasswordBytes = argon2.IDKey(passwordBytes, *salt, time, memory, threads, keyLen)

	// Convert the hashed password to a base64 encoded string -> easier to store in DB
	var base64EncodedPasswordHash = base64.RawStdEncoding.EncodeToString(hashedPasswordBytes)
	var base64EncodedPasswordSalt = base64.RawStdEncoding.EncodeToString(*salt)

	return base64EncodedPasswordHash, base64EncodedPasswordSalt
}

func ComparePassword(passwordOne, passwordHashDB, salt *string) (bool, error) {

	saltBytes, err := base64.RawStdEncoding.DecodeString(*salt)

	if err != nil {
		return false, err
	}

	// Using recieved salt, hash the password
	hashedAttempt, _ := HashPassword(passwordOne, &saltBytes)

	fmt.Println(hashedAttempt, *passwordHashDB)

	return hashedAttempt == *passwordHashDB, nil
}
