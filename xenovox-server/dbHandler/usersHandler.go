package dbhandler

import (
	"database/sql"
	"log"

	"github.com/lib/pq"
)

func GetUserInfo(id *int) (row *sql.Rows, status bool) {
	row, err := db.Query(`SELECT id, name, username, email, picture FROM users WHERE id = $1`, id)
	if err != nil {
		status = false
	} else {
		status = true
	}
	return
}

func GetUserInfoMin(id *int) (row *sql.Rows, status bool) {
	row, err := db.Query(`SELECT id, username, picture FROM users WHERE id = $1`, id)
	if err != nil {
		status = false
	} else {
		status = true
	}
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

func GetUsernameAndPicture(id *int, getPic bool) (row *sql.Rows, status bool) {
	q := `SELECT username `
	if getPic {
		q += `, picture `
	}
	row, err := db.Query(q+`FROM users WHERE id = $1`, id)
	if err != nil {
		log.Fatal(err)
		status = false
	} else {
		status = true
	}

	return
}
