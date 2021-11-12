package ws

import (
	"fmt"
	ws "go-websocket/pkg/ws/datastructs"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func CreateToken(username string, expirationTime time.Time) (string, error) {
	var err error

	//Creating Access Token
	claims := &ws.Claims{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(os.Getenv("ACCESS_SECRET")))

	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func IsValidToken(tokenStr string) bool {

	//Creating Access Token
	claims := &ws.Claims{}

	// Parse the JWT string and store the result in `claims`.
	_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return os.Getenv("ACCESS_SECRET"), nil
	})

	//False if any err is returned
	return err == nil

}

func RefreshToken(w http.ResponseWriter, r *http.Request, expirationTime time.Time) error {

	c, err := r.Cookie("token")
	if err != nil {
		return err
	}

	tknStr := c.Value
	claims := &ws.Claims{}

	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return os.Getenv("ACCESS_SECRET"), nil
	})

	if err != nil {
		return err
	}

	if !tkn.Valid {
		return fmt.Errorf("token is invalid")
	}

	// We ensure that a new token is not issued until enough time has elapsed
	// In this case, a new token will only be issued if the old token is within
	// 30 seconds of expiry. Otherwise, return a bad request status
	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Minute {
		return fmt.Errorf("token is not valid")
	}

	// Now, create a new token for the current use, with a renewed expiration time
	claims.ExpiresAt = expirationTime.Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(os.Getenv("ACCESS_SECRET"))

	if err != nil {
		return err
	}

	// Set the new token as the users `token` cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenString,
		Expires:  expirationTime,
		HttpOnly: true,
	})

	return nil

}
