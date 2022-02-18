package dbhandler

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/go-redis/redis"
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
