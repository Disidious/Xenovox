package dbhandler

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"

	structs "github.com/Disidious/Xenovox/Structs"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
)

var db *sql.DB
var rdb *redis.Client

const (
	host     = "localhost"
	port     = 5432
	dbuser   = "fady"
	password = "123"
	dbname   = "xenovox"
)

const (
	rHost = "localhost:6379"
	rPass = "123"
	rDB   = 0
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

func ConnectDb() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, dbuser, password, dbname)

	var dberr interface{}
	db, dberr = sql.Open("postgres", psqlInfo)

	if dberr != nil {
		panic(dberr)
	}

	dberr = db.Ping()
	if dberr != nil {
		panic(dberr)
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     rHost,
		Password: rPass,
		DB:       rDB,
	})

	log.Println("Successfully connected!")
}

func CloseDb() {
	db.Close()
}

func Login(user *structs.User) (string, string) {

	//Get id, username, password and token of the user logging in by his/her username.
	getUserQ := `SELECT id, username, password, token FROM users WHERE username = $1`

	var id, username, password, token string
	row := db.QueryRow(getUserQ, user.Username)
	err := row.Scan(&id, &username, &password, &token)

	//If any errors occured, the user was not found or the password was wrong then authentication is failed
	if (err != sql.ErrNoRows && password != user.Password) || err == sql.ErrNoRows {
		return "AUTH_FAILED", ""
	}

	//Generate Token flag.
	var generate bool

	/*
		Set Generate Token flag to true if:
		-	The token is empty.
		-	There is an error during the token retreival from Redis.
		-	The stored user id in Redis doesn't match the current user logging in.

		Deletes value of the token if there is a stored id that doesn't match.
	*/
	if token == "" {
		generate = true
	} else {
		storedId := getIdRedis(&token)
		if storedId != id {
			generate = true
		}

		if storedId != id {
			rdb.Del(token)
		}
	}

	//Create a random string with length 50 and putting it in Redis and the database
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

	//Check if the username or email already exists.
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

	//Generate token for new user.
	token := generateToken()

	//Insert new user in the database and store the id with the token as the key on Redis.
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

func GetUserInfo(id *int) (row *sql.Row) {
	row = db.QueryRow(`SELECT id, name, username, email, picture FROM users WHERE id = $1`, id)
	return
}

func GetFriends(id *int) (rows *sql.Rows, status bool) {
	rows, err := db.Query(`SELECT users.id, users.username, users.picture 
	FROM users 
	INNER JOIN relations 
	ON ((relations.user1id=$1 AND users.id = relations.user2id) or (relations.user2id=$1 AND users.id = relations.user1id)) 
	AND relations.relation = 1;`, id)

	if err != nil {
		log.Fatal(err)
		status = false
	} else {
		status = true
	}

	return
}

func GetChat(id *int, friendId *int) (rows *sql.Rows, status bool) {
	rows, err := db.Query(`SELECT message, senderId, receiverId FROM private_messages 
	WHERE (senderId = $1 AND receiverId = $2) OR (senderId = $2 AND receiverId = $1)`, id, friendId)

	if err != nil {
		log.Fatal(err)
		status = false
	} else {
		status = true
	}

	return
}

func InsertPrivateMessage(message *structs.Message) string {
	// Check if the relation is Blocked before sending the message
	checkQ := `SELECT relation FROM relations WHERE (user1id = $1 AND user2id = $2) OR (user1id = $2 AND user2id = $1)`
	row := db.QueryRow(checkQ, message.SenderId, message.ReceiverId)
	var relation int
	row.Scan(&relation)
	if relation == 2 {
		return "BLOCKED"
	}

	addQ := `INSERT INTO private_messages (senderId, receiverId, message) VALUES($1, $2, $3)`
	log.Println(message.SenderId, message.ReceiverId, message.Message)
	_, err := db.Exec(addQ, message.SenderId, message.ReceiverId, message.Message)
	if err != nil {
		log.Println(err)
		return "FAILED"
	}
	return "SUCCESS"
}

func GetFriendRequests(id *int) (rows *sql.Rows, status bool) {
	rows, err := db.Query(`SELECT r.id as relationId, u.username, u.id as userId FROM relations r INNER JOIN users u ON r.user1id = u.id 
							WHERE r.user2id = $1 AND r.relation = 0`, id)

	if err != nil {
		log.Fatal(err)
		status = false
	} else {
		status = true
	}

	return
}

func UpsertRelation(relation *structs.Relation) string {
	/*
		-1 = Remove Relation
		0  = Friend Request
		1  = Friends
		2  = Blocked
	*/

	// Delete relation
	// Only blocker can delete a Blocked relation
	if relation.Relation == -1 {
		q := `DELETE FROM relations 
		WHERE ( ((user1id = $1 AND user2id = $2) OR (user1id = $2 AND user2id = $1)) AND relation != 2 )
		OR ((user1id = $1 AND user2id = $2) AND relation = 2)`
		_, err := db.Exec(q, relation.User1Id, relation.User2Id)
		if err != nil {
			log.Println(err)
			return "FAILED"
		}
		return "SUCCESS"
	}

	// Only send a Friend Request relation if there is no relation that already exists between the two users
	if relation.Relation == 0 {
		checkQ := `SELECT id FROM relations WHERE (user1id = $1 AND user2id = $2) OR (user1id = $2 AND user2id = $1)`
		row := db.QueryRow(checkQ, relation.User1Id, relation.User2Id)
		var temp interface{}
		err := row.Scan(&temp)
		if err != sql.ErrNoRows {
			return "FAILED"
		}

		q := `INSERT INTO relations (user1id, user2id, relation) VALUES($1, $2, $3)`
		_, err = db.Exec(q, relation.User1Id, relation.User2Id, relation.Relation)
		if err != nil {
			log.Println(err)
			return "FAILED"
		}
		return "SUCCESS"
	}

	// Only the receiver can accept the Friend Request relation
	if relation.Relation == 1 {
		q := `UPDATE relations SET relation = $1 WHERE user2id = $2 AND user1id = $3 AND relation = 0`
		_, err := db.Exec(q, relation.Relation, relation.User1Id, relation.User2Id)
		if err != nil {
			return "FAILED"
		}
		return "SUCCESS"
	}

	// Check if there is a relation that already exists and change it to Blocked, if not insert a new relation
	// Set the sender and the receiver along with the Blocked relation
	if relation.Relation == 2 {
		checkQ := `SELECT id FROM relations WHERE (user1id = $1 AND user2id = $2) OR (user1id = $2 AND user2id = $1)`
		row := db.QueryRow(checkQ, relation.User1Id, relation.User2Id)
		var id string
		err := row.Scan(&id)

		if err == sql.ErrNoRows {
			q := `INSERT INTO relations (user1id, user2id, relation) VALUES($1, $2, $3)`
			_, err = db.Exec(q, relation.User1Id, relation.User2Id, relation.Relation)
			if err != nil {
				log.Println(err)
				return "FAILED"
			}
			return "SUCCESS"
		}

		q := `UPDATE relations SET user1id = $1, user2id = $2, relation = $3 WHERE id = $4`
		_, err = db.Exec(q, relation.User1Id, relation.User2Id, relation.Relation, id)
		if err != nil {
			log.Println(err)
			return "FAILED"
		}
		return "SUCCESS"
	}

	return "FAILED"
}
