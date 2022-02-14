package dbhandler

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"reflect"
	"strconv"

	structs "github.com/Disidious/Xenovox/Structs"
	"github.com/go-redis/redis"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

var db *sql.DB
var rdb *redis.Client

//var rcache *redis.Client

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

// const (
// 	rcHost = "localhost:6380"
// 	rcPass = "123"
// 	rcDB = 0
// )

func generateToken() string {
	b := make([]byte, 50)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func InsertGroupMessage(message *structs.GroupMessage) string {
	checkQ := `SELECT COUNT(*) FROM group_members WHERE userid = $1 AND groupid = $2 AND "left" = false`
	row := db.QueryRow(checkQ, message.SenderId, message.GroupId)
	var count int
	row.Scan(&count)
	if count != 1 {
		return "FAILED"
	}

	addQ := `INSERT INTO group_messages (senderid, groupid, message, issystem) VALUES($1, $2, $3, $4)`
	_, err := db.Exec(addQ, message.SenderId, message.GroupId, message.Message, message.IsSystem)
	if err != nil {
		log.Println(err)
		return "FAILED"
	}

	return "SUCCESS"
}

func InsertDirectMessage(message *structs.Message) string {
	// Check if the relation is Blocked before sending the message
	checkQ := `SELECT relation FROM relations WHERE (user1id = $1 AND user2id = $2) OR (user1id = $2 AND user2id = $1)`
	row := db.QueryRow(checkQ, message.SenderId, message.ReceiverId)
	var relation int
	row.Scan(&relation)
	if relation == 2 {
		return "BLOCKED"
	}

	addQ := `INSERT INTO private_messages (senderId, receiverId, message) VALUES($1, $2, $3)`
	_, err := db.Exec(addQ, message.SenderId, message.ReceiverId, message.Message)
	if err != nil {
		log.Println(err)
		return "FAILED"
	}

	return "SUCCESS"
}

func AppendUnreadDM(senderId *int, receiverId *int) bool {
	return appendUnreadMsg(senderId, receiverId, false)
}
func AppendUnreadGM(groupId *int, users *[]structs.User) {
	for _, user := range *users {
		appendUnreadMsg(groupId, &user.Id, true)
	}
}
func UpdateGMsToRead(groupId *int, id *int) bool {
	return removeUnreadMsg(groupId, id, true)
}
func UpdateDMsToRead(senderId *int, receiverId *int) bool {
	return removeUnreadMsg(senderId, receiverId, false)
}
func GetUnreadMsgs(id *int) *[]redis.Z {
	key := "N-" + strconv.Itoa(*id)
	res, err := rdb.ZRangeWithScores(key, 0, -1).Result()
	if err != nil {
		log.Fatal(err)
		return nil
	} else {
		return &res
	}
}

func appendUnreadMsg(id *int, receiverId *int, group bool) bool {
	var member string
	if group {
		member = "GM-"
	} else {
		member = "DM-"
	}
	member += strconv.Itoa(*id)
	key := "N-" + strconv.Itoa(*receiverId)

	_, err := rdb.ZIncrBy(key, 1, member).Result()
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}

func removeUnreadMsg(id *int, receiverId *int, group bool) bool {
	var member string
	if group {
		member = "GM-"
	} else {
		member = "DM-"
	}
	member += strconv.Itoa(*id)
	key := "N-" + strconv.Itoa(*receiverId)

	_, err := rdb.ZRem(key, member).Result()
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
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

func GetUsernames(ids []int) (rows *sql.Rows, status bool) {
	rows, err := db.Query(`SELECT id, username FROM users WHERE id = ANY($1)`, pq.Array(ids))
	if err != nil {
		log.Fatal(err)
		status = false
	}

	status = true
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

func GetGroups(id *int) (rows *sql.Rows, status bool) {
	rows, err := db.Query(`SELECT groups.id, groups.name, groups.ownerid, groups.picture 
	FROM groups
	INNER JOIN group_members
	ON userid = $1 AND groupid = groups.id and "left" = false`, id)

	if err != nil {
		log.Fatal(err)
		status = false
	} else {
		status = true
	}

	return
}

func GetAllConnections(id *int) (frows *sql.Rows, grows *sql.Rows, status bool) {
	frows, status = GetFriends(id)
	if !status {
		return
	}

	grows, status = GetGroups(id)

	return
}

func GetPrivateChat(id *int, id2 *int) (rows *sql.Rows, status bool) {
	rows, err := db.Query(`SELECT message, senderId FROM private_messages 
	WHERE (senderId = $1 AND receiverId = $2) OR (senderId = $2 AND receiverId = $1) ORDER BY id`, id, id2)

	if err != nil {
		log.Fatal(err)
		status = false
	} else {
		status = true
	}

	return
}

func GetGroupMembers(groupId *int, includingLeftMembers *bool) (rows *sql.Rows, status bool) {
	query := `SELECT users.id as "id", username, picture FROM users 
	INNER JOIN group_members ON userid = users.id AND groupid = $1 `

	if !*includingLeftMembers {
		query += `AND "left" = false`
	}

	rows, err := db.Query(query, groupId)
	if err != nil {
		log.Fatal(err)
		status = false
		return
	}

	status = true
	return
}

func GetGroupChatAndMembers(id *int, groupId *int) (hrows *sql.Rows, mrows *sql.Rows, status bool) {
	checkQ := `SELECT COUNT(*) FROM group_members WHERE userid = $1 AND groupid = $2 and "left" = false`
	row := db.QueryRow(checkQ, id, groupId)
	var count int
	row.Scan(&count)
	if count == 0 {
		status = false
		return
	}

	hrows, err := db.Query(`SELECT message, senderid, issystem FROM group_messages WHERE groupid = $1`, groupId)
	if err != nil {
		log.Fatal(err)
		status = false
		return
	}

	includingLeftMembers := true
	mrows, ok := GetGroupMembers(groupId, &includingLeftMembers)
	if !ok {
		log.Fatal(err)
		status = false
		return
	}

	status = true
	return
}

func GetFriendRequests(id *int) (rows *sql.Rows, status bool) {
	rows, err := db.Query(`SELECT r.id as relationid, u.username, u.id as userid FROM relations r INNER JOIN users u ON r.user1id = u.id 
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

func GetUnreadDMs(id *int) (rows *sql.Rows, status bool) {
	rows, err := db.Query(`SELECT DISTINCT senderid FROM private_messages WHERE receiverid = $1 AND read = false`, id)
	if err != nil {
		log.Fatal(err)
		status = false
	} else {
		status = true
	}

	return
}

func FriendReqExists(id *int) (result bool) {
	row := db.QueryRow(`SELECT COUNT(*) as count FROM relations WHERE relation = 0 AND user2id = $1 LIMIT 1`, id)

	var count int
	row.Scan(&count)

	if count == 0 {
		result = false
	} else {
		result = true
	}

	return
}

func AddToGroup(id *int, friendIds *[]int, groupId *int) string {
	// Checks if the user to add is already a member
	checkQ := `SELECT COUNT(*) FROM group_members WHERE userid = ANY($1) AND groupid = $2 and "left" = false`
	row := db.QueryRow(checkQ, pq.Array(*friendIds), groupId)
	var count int
	row.Scan(&count)
	if count != 0 {
		return "EXISTS"
	}

	// Checks if the user that is adding is a member
	checkQ3 := `SELECT COUNT(*) FROM group_members WHERE userid = $1 AND groupid = $2 AND "left" = false`
	row3 := db.QueryRow(checkQ3, id, groupId)
	var count3 int
	row3.Scan(&count3)
	if count3 == 0 {
		return "NOT_MEMBER"
	}

	// Checks if both users are friends
	checkQ2 := `SELECT COUNT(*) FROM relations 
	WHERE ((user1id = $1 AND user2id = ANY($2)) OR (user1id = ANY($2) AND user2id = $1)) AND relation = 1`
	row2 := db.QueryRow(checkQ2, id, pq.Array(*friendIds))
	var count2 int
	row2.Scan(&count2)
	if count2 == 0 {
		return "NOT_FRIENDS"
	}

	leftMemsQ := `SELECT userid as "id" FROM group_members WHERE userid = ANY($1) AND groupid = $2 and "left" = true`
	rows, err := db.Query(leftMemsQ, pq.Array(*friendIds), groupId)

	var leftIds []int
	isLeft := make(map[int]bool)
	if err == nil {
		users, ok := structs.StructifyRows(rows, reflect.TypeOf(structs.User{}))
		if ok {
			for _, user := range users {
				userCasted := user.(*structs.User)
				leftIds = append(leftIds, userCasted.Id)
				isLeft[userCasted.Id] = true
			}
		}
	}

	if len(leftIds) != 0 {
		updateQ := `UPDATE group_members SET "left" = false WHERE userid = ANY($1)`
		_, err = db.Exec(updateQ, pq.Array(leftIds))
		if err != nil {
			log.Println(err)
			return "FAILED"
		}
	}

	if len(leftIds) != len(*friendIds) {
		addQ := `INSERT INTO group_members (userid, groupid) VALUES`
		groupIdStr := strconv.Itoa(*groupId)
		for _, friendId := range *friendIds {
			if isLeft[friendId] {
				continue
			}
			addQ += "(" + strconv.Itoa(friendId) + "," + groupIdStr + "),"
		}
		if addQ[len(addQ)-1] == ',' {
			addQ = addQ[:len(addQ)-1]
		}

		_, err = db.Exec(addQ)
		if err != nil {
			log.Println(err)
			return "FAILED"
		}
	}

	return "SUCCESS"
}

func LeaveGroup(id *int, groupId *int) bool {
	remQ := `UPDATE group_members SET "left" = true WHERE userid = $1 AND groupid = $2`
	_, err := db.Exec(remQ, id, groupId)

	if err != nil || !removeUnreadMsg(groupId, id, true) {
		log.Println(err)
		return false
	}

	ownerQ := `SELECT ownerid FROM groups WHERE id = $1`
	row := db.QueryRow(ownerQ, groupId)
	var ownerId int
	err = row.Scan(&ownerId)
	if err != nil {
		log.Println(err)
		return false
	}

	if ownerId == *id {
		getMemberQ := `SELECT userid FROM group_members WHERE groupid = $1 && userid != $2 LIMIT 1`
		row := db.QueryRow(getMemberQ, groupId, id)
		var newOwnerId int
		err = row.Scan(&ownerId)
		if err != nil {
			log.Println(err)
			return true
		}

		updateOwnerQ := `UPDATE groups SET ownerid = $1 WHERE id = $2`
		_, err = db.Exec(updateOwnerQ, newOwnerId, groupId)
		if err != nil {
			log.Println(err)
			return false
		}
	}

	return true
}
