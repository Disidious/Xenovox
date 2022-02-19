package dbhandler

import (
	"database/sql"
	"log"

	structs "github.com/Disidious/Xenovox/structs"
)

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

func GetAllConnections(id *int) (frows *sql.Rows, grows *sql.Rows, status bool) {
	frows, status = GetFriends(id)
	if !status {
		return
	}

	grows, status = GetGroups(id)
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
		if err != nil ||
			!removeUnreadMsg(&relation.User1Id, &relation.User2Id, false) ||
			!removeUnreadMsg(&relation.User2Id, &relation.User1Id, false) {
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

		if !removeUnreadMsg(&relation.User1Id, &relation.User2Id, false) ||
			!removeUnreadMsg(&relation.User2Id, &relation.User1Id, false) {
			log.Println(err)
			return "FAILED"
		}

		return "SUCCESS"
	}

	return "FAILED"
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
