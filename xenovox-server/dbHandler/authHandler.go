package dbhandler

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"strconv"

	structs "github.com/Disidious/Xenovox/structs"
)

func generateToken() string {
	b := make([]byte, 50)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func setTokenRedis(token *string, id *string) bool {
	err := rdb.Set(*token, *id, 0)
	return err.Err() != nil
}

func getIdRedis(token *string) string {
	id, err := rdb.Get(*token).Result()
	if err != nil {
		return "FAILED"
	}

	return id
}

func Login(user *structs.User) (string, string) {

	// Get id, username, password and token of the user logging in by his/her username.
	getUserQ := `SELECT id, username, password, token FROM users WHERE username = $1`

	var id, username, password, token string
	row := db.QueryRow(getUserQ, user.Username)
	err := row.Scan(&id, &username, &password, &token)

	// If any errors occured, the user was not found or the password was wrong then authentication is failed
	if (err != sql.ErrNoRows && password != user.Password) || err == sql.ErrNoRows {
		return "AUTH_FAILED", ""
	}

	// Generate Token flag.
	var generate bool

	/*
		Set Generate Token flag to true if:
		-	The token is empty.
		-	There is an error during the token retreival from Redis.
		-	The stored user id in Redis doesn't match the current user logging in.
	*/
	if token == "" || getIdRedis(&token) != id {
		generate = true
	}

	// Create a random string with length 50 and putting it in Redis and the database
	if generate {
		token = generateToken()
		putTokenQ := `UPDATE users SET token = $1 WHERE id = $2`
		if _, err = db.Exec(putTokenQ, token, id); err != nil {
			return "AUTH_FAILED", ""
		}

		err := setTokenRedis(&token, &id)
		if err {
			return "AUTH_FAILED", ""
		}
	}

	return "SUCCESS", token
}

func Register(newUser *structs.User) string {
	// Check if the username or email already exists.
	checkQ := `SELECT username, email FROM users WHERE email = $1 OR username = $2`

	var email, username string
	row := db.QueryRow(checkQ, newUser.Email, newUser.Username)
	err := row.Scan(&username, &email)

	if err != sql.ErrNoRows {
		if email == newUser.Email {
			return "EMAIL_EXISTS"
		} else {
			return "USERNAME_EXISTS"
		}
	}

	// Generate token for new user.
	token := generateToken()

	// Insert new user in the database and store the id with the token as the key on Redis.
	var id string
	q := `INSERT INTO users (username, password, name, email, picture, token) VALUES($1,$2,$3,$4,$5,$6) RETURNING id`
	retId := db.QueryRow(q, newUser.Username, newUser.Password, newUser.Name, newUser.Email, newUser.Picture, token)
	err = retId.Scan(&id)

	if err != nil {
		return "FAILED"
	}

	idString := string(id)
	setTokenRedis(&token, &idString)

	return "SUCCESS"
}

func Authorize(token *string) string {
	id, err := rdb.Get(*token).Result()
	if err != nil || id == "" {
		return "UNAUTHORIZED"
	}
	return "AUTHORIZED"
}

func GetUserId(token *string) int {
	id, err := strconv.Atoi(getIdRedis(token))
	if err != nil {
		return -1
	}
	return id
}
