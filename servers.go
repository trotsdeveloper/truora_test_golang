package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// Server ...
type Server struct {
	address  string
	sslGrade string
	country  string
	owner    string
}

// ServerTest ...
type ServerTest struct {
	testHour       string // MM - DD - YYYY hh:mm:ss
	testInProgress bool
	servers        []Server
	sslGrade       string
	isDown         bool
}

// ServerTestComplete ...
type ServerTestComplete struct {
	servers          []Server
	sslGrade         string
	previousSslGrade string
	logo             string
	title            string
	isDown           bool
}

// DAOInterface ...
type DAOInterface interface {
	initTable()
	selectInDB()
	createInDB()
	updateInDB()
	deleteInDB()
}

func main() {
	// Connect to the "servers" database.
	db, err := sql.Open("postgres",
		"postgresql://manuelams@localhost:26257/servers?ssl=true&sslmode=require&sslrootcert=../certs/ca.crt&sslkey=../certs/client.maxroach.key&sslcert=../certs/client.maxroach.crt")
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	fmt.Printf("Hello world %T\n", db)
	defer db.Close()

}
