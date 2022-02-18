package dbhandler

import (
	"database/sql"
	"log"
	"strconv"

	structs "github.com/Disidious/Xenovox/Structs"
	"github.com/go-redis/redis"
)

func InsertGroupMessage(message *structs.GroupMessage) (id int, status string) {
	// Checks if sender is a member of the group
	checkQ := `SELECT COUNT(*) FROM group_members WHERE userid = $1 AND groupid = $2`
	row := db.QueryRow(checkQ, message.SenderId, message.GroupId)
	var count int
	row.Scan(&count)
	if count != 1 {
		status = "FAILED"
		return
	}

	addQ := `INSERT INTO group_messages (senderid, groupid, message, issystem) VALUES($1, $2, $3, $4) RETURNING id`
	rowId := db.QueryRow(addQ, message.SenderId, message.GroupId, message.Message, message.IsSystem)

	rowId.Scan(&id)
	status = "SUCCESS"
	return
}

func InsertDirectMessage(message *structs.Message) (id int, status string) {
	// Check if the relation is Blocked before sending the message
	checkQ := `SELECT relation FROM relations WHERE (user1id = $1 AND user2id = $2) OR (user1id = $2 AND user2id = $1)`
	row := db.QueryRow(checkQ, message.SenderId, message.ReceiverId)
	var relation int
	row.Scan(&relation)
	if relation == 2 {
		status = "BLOCKED"
		return
	}

	addQ := `INSERT INTO private_messages (senderId, receiverId, message) VALUES($1, $2, $3) RETURNING id`
	rowId := db.QueryRow(addQ, message.SenderId, message.ReceiverId, message.Message)
	rowId.Scan(&id)
	status = "SUCCESS"

	return
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

func GetPrivateChat(id *int, id2 *int) (rows *sql.Rows, status bool) {
	rows, err := db.Query(`
	SELECT private_messages.id as "id", message, senderId, username, picture FROM private_messages INNER JOIN users ON users.id = senderId
	WHERE (senderId = $1 AND receiverId = $2) OR (senderId = $2 AND receiverId = $1) ORDER BY private_messages.id`, id, id2)

	if err != nil {
		log.Fatal(err)
		status = false
	} else {
		status = true
	}

	return
}

func GetGroupChat(id *int, groupId *int) (rows *sql.Rows, status bool) {
	// Checks if user is a member of the group
	checkQ := `SELECT COUNT(*) FROM group_members WHERE userid = $1 AND groupid = $2`
	row := db.QueryRow(checkQ, id, groupId)
	var count int
	row.Scan(&count)
	if count == 0 {
		status = false
		return
	}

	rows, err := db.Query(`
	SELECT group_messages.id as "id", message, senderid, username, picture, issystem 
	FROM group_messages INNER JOIN users ON users.id = senderid WHERE groupid = $1`, groupId)
	if err != nil {
		log.Fatal(err)
		status = false
		return
	}

	status = true
	return
}