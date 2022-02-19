package dbhandler

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/Disidious/Xenovox/structs"
	"github.com/lib/pq"
)

func GetGroups(id *int) (rows *sql.Rows, status bool) {
	rows, err := db.Query(`SELECT groups.id, groups.name, groups.ownerid, groups.picture 
	FROM groups
	INNER JOIN group_members
	ON userid = $1 AND groupid = groups.id`, id)

	if err != nil {
		log.Fatal(err)
		status = false
	} else {
		status = true
	}

	return
}

func GetGroupMembers(groupId *int) (rows *sql.Rows, status bool) {
	query := `SELECT gm.id as "id", userid, username, picture FROM users 
	INNER JOIN group_members gm ON userid = users.id AND groupid = $1`

	rows, err := db.Query(query, groupId)
	if err != nil {
		log.Fatal(err)
		status = false
		return
	}

	status = true
	return
}

func CreateGroup(id *int, group *structs.Group) (status bool) {
	var groupId int
	err := db.QueryRow(`INSERT INTO groups (name, picture) VALUES($1, $2) RETURNING id`, group.Name, group.Picture).Scan(&groupId)
	if err != nil {
		log.Fatal(err)
		status = false
		return
	}

	var gmId int
	err = db.QueryRow(`INSERT INTO group_members (userid, groupid) VALUES($1, $2) RETURNING id`, id, groupId).Scan(&gmId)
	if err != nil {
		log.Fatal(err)
		status = false
	}

	_, err = db.Exec(`UPDATE groups SET ownerid = $1 WHERE id = $2`, gmId, groupId)
	if err != nil {
		log.Fatal(err)
		status = false
	}

	group.Id = groupId
	status = true
	return
}

func AddToGroup(id *int, friendIds *[]int, groupId *int) string {
	if len(*friendIds) == 0 {
		return "FAILED"
	}

	// Checks if the user to add is already a member and if the user that is adding is a member
	checkQ := `SELECT userid FROM group_members WHERE groupid = $1 AND (userid = ANY($2) OR userid = $3)`
	rows, err := db.Query(checkQ, groupId, pq.Array(*friendIds), id)
	if err == nil && rows.Next() {
		var userId int
		rows.Scan(&userId)
		if userId != *id || rows.Next() {
			return "EXISTS"
		}
	} else {
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

	addQ := `INSERT INTO group_members (userid, groupid) VALUES`
	groupIdStr := strconv.Itoa(*groupId)
	for _, friendId := range *friendIds {
		addQ += "(" + strconv.Itoa(friendId) + "," + groupIdStr + "),"
	}
	addQ = addQ[:len(addQ)-1]

	_, err = db.Exec(addQ)
	if err != nil {
		log.Println(err)
		return "FAILED"
	}

	return "SUCCESS"
}

func changeGroupOwner(ownerId *int, groupId *int) bool {
	if *ownerId == -1 {
		ownerId = nil
	}
	updateOwnerQ := `UPDATE groups SET ownerid = $1 WHERE id = $2`
	_, err := db.Exec(updateOwnerQ, ownerId, groupId)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func ChangeGroupOwnerRand(id *int, groupId *int) bool {
	// Change owner to be the first found member if the user that left is the owner
	ownerQ := `SELECT gm.userid FROM group_members gm INNER JOIN groups g ON gm.id = ownerid AND g.id = $1`
	row := db.QueryRow(ownerQ, groupId)
	var ownerId int
	err := row.Scan(&ownerId)
	if err == sql.ErrNoRows {
		return true
	}

	if ownerId == *id {
		getMemberQ := `SELECT id FROM group_members WHERE groupid = $1 AND userid != $2 LIMIT 1`
		row := db.QueryRow(getMemberQ, groupId, id)
		var newOwnerId int
		err = row.Scan(&newOwnerId)
		if err != nil && err == sql.ErrNoRows {
			newOwnerId = -1
		} else if err != nil {
			log.Println(err)
			return false
		}
		return changeGroupOwner(&newOwnerId, groupId)
	}

	return true
}

func LeaveGroup(id *int, groupId *int) bool {
	remQ := `DELETE FROM group_members WHERE userid = $1 AND groupid = $2`
	_, err := db.Exec(remQ, id, groupId)

	if err != nil || !removeUnreadMsg(groupId, id, true) {
		log.Println(err)
		return false
	}

	return true
}

func GetGroupInfo(groupId *int) (row *sql.Rows, status bool) {
	q := `SELECT id, name, ownerid, picture FROM groups WHERE id = $1`
	row, err := db.Query(q, groupId)
	if err != nil {
		status = false
	} else {
		status = true
	}
	return
}
