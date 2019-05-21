package main

import (
	"encoding/json"
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"
	_ "github.com/lib/pq"
	"github.com/cockroachdb/cockroach-go/crdb"
)

// Constants
const (
	USER        = "manuelams"
	HOST        = "localhost"
	PORT        = 26257
	DATABASE    = "servers"
	SSL         = true
	SSLMODE     = "require"
	SSLROOTCERT = "../certs/ca.crt"
	SSLKEY      = "../certs/client.manuelams.key"
	SSLCERT     = "../certs/client.manuelams.crt"
)

func initDB() (*sql.DB, error) {
	/*db, err := sql.Open("postgres",
	"postgresql://manuelams@localhost:26257/servers?ssl=true&sslmode=require&sslrootcert=../certs/ca.crt&sslkey=../certs/client.manuelams.key&sslcert=../certs/client.manuelams.crt")
	*/
	psqlInfo := fmt.Sprintf("posgresql://%s@%s:%s/%s?ssl=%s&sslmode=%s&sslrootcert=%s&sslkey=%s&sslcert=%s",
		USER, HOST, strconv.Itoa(PORT), DATABASE, strconv.FormatBool(SSL), SSLMODE, SSLROOTCERT, SSLKEY, SSLCERT)
	db, err := sql.Open("postgres", psqlInfo)
	return db, err
}

// CustomError ...
type CustomError struct {
	customMessage string
}

func (e *CustomError) Error() string {
	return fmt.Sprintf("%s", e.customMessage)
}

func addQuotes(word string) string {
	return fmt.Sprintf(`'%v'`, word)
}

// Server ...
type Server struct {
	id       int    `json:"-"`// SERIAL PRIMARY KEY
	address  string `json:"address"`// VARCHAR[16]
	sslGrade string `json:"ssl_grade"`// VARCHAR[5]
	country  string `json:"country"` // VARCHAR[20]
	owner    string `json:"owner"`// VARCHAR[50]
}

// ServerTest ...
type ServerTest struct {
	id             int    // SERIAL PRIMARY KEY
	domain         string // VARCHAR[100]
	testHour       string // VARCHAR[30]
	testInProgress bool		// boolean
	servers        []Server
	sslGrade       string	// VARCHAR [5]
	isDown         bool		// boolean
}

// ServerTestComplete ...
type ServerTestComplete struct {
	testInProgress bool `json:"-"`
	servers []Server	`json:"servers"`
	serversChanged bool `json:"servers_changed"`
	sslGrade string `json:"ssl_grade"`
	previousSslGrade string `json:"previous_ssl_grade"`
	logo string `json:"logo"`
	title string `json:"title"`
	isDown bool	 `json:"is_down"`
}

// DAO..
type DAO interface {
	existsInDb(dbc interface{}) (bool, error)
	selectInDB(dbc interface{}) error
	createInDB(dbc interface{}) error
	updateInDB(dbc interface{}) error
	deleteInDB(dbc interface{}) error
}

func Exec(dbc interface{}, sqlString string, args ...interface{}) (r *sql.Result, err error) {
		switch v := dbc.(type) {
		case *sql.DB:
			r, err = v.Exec(sqlString, args...)
		case *sql.Tx:
			r, err = v.Exec(sqlString, args...)
		default:
			err = &CustomError{"No valid DB Controller"}
		}
		return
}

func QueryRow(dbc interface{}, sqlString string, args ...interface{}) (r *sql.Row, err error) {
		switch v := dbc.(type) {
		case *sql.DB:
			r, err = v.QueryRow(sqlString, args...)
		case *sql.Tx:
			r, err = v.QueryRow(sqlString, args...)
		default:
			err = &CustomError{"No valid DB Controller"}
		}
		return
}

func Query(dbc interface{}, sqlString string, args ...interface{}) (r *sql.Rows, err error) {
		switch v := dbc.(type) {
		case *sql.DB:
		r, err = v.Query(sqlString, args...)
		case *sql.Tx:
		r, err = v.Query(sqlString, args...)
		default:
		err = &CustomError{"No valid DB Controller"}
		}
		return
}

func InitServerTestTable(dbc *sql.DB) error {
		sqlStatement := `CREATE TABLE serverTest IF NOT EXISTS (id SERIAL PRIMARY KEY,
					domain VARCHAR[100], testHour VARCHAR[30], testInProgress boolean,
					sslGrade VARCHAR[5], isDown boolean);`
		_, err := dbc.Exec(sqlStatement)
		return err
}
func InitServerTable(dbc *sql.DB) error {
		sqlStatement := `CREATE TABLE server IF NOT EXISTS (id SERIAL PRIMARY KEY,
			serverTestId integer REFERENCES serverTest(id),	address VARCHAR[16],
			sslGrade VARCHAR[5], country VARCHAR[20], owner VARCHAR[50]);`
		_, err := dbc.Exec(sqlStatement)
		return err
}

func DropServerTable(dbc *sql.DB) error {
	sqlStatement := `DROP TABLE IF EXISTS server;`
	_, err := dbc.Exec(sqlStatement)
	return err
}

func DropServerTestTable(dbc *sql.DB) error {
	sqlStatement := `DROP TABLE IF EXISTS serverTable;`
	_, err := dbc.Exec(sqlStatement)
	return err
}

func (s *Server) existsInDB(dbc interface{}) (bool, error) {
	sqlStatement := `SELECT id FROM server WHERE id =$1;`
	row := QueryRow(dbc, sqlStatement, s.id)
	err := row.Scan(&s.id)
	switch err {
	case sql.ErrNoRows:
		return false, nil
	case nil:
		return true, nil
	default:
		return false, err
	}
}

func (s *Server) selectInDB(dbc interface{}) error {
	sqlStatement := "SELECT address, sslGrade, country, owner FROM server WHERE id=$1;"
	row := QueryRow(dbc, sqlStatement, s.id)
	err := row.Scan(&s.address, &s.sslGrade, &s.country, &s.owner)
	switch err {
	case sql.ErrNoRows:
		return &CustomError{"No rows were returned."}
	case nil:
		return nil
	default:
		return err
	}
}

func (s *Server) createInDB(dbc interface{}) error {
	sqlStatement := `INSERT INTO server (address, sslGrade, country, owner)
	VALUES ($1, $2, $3, $4) RETURNING id;`
	row := QueryRow(dbc, sqlStatement, s.address, s.sslGrade, s.country, s.owner)
	err := row.Scan(&s.id)
	return err
}


func (s *Server) updateInDB(dbc interface{}) error {
	sqlStatement := `UPDATE server SET address = $2, sslGrade = $3, country = $4,
	owner = $5 WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, addQuotes(strconv.Itoa(s.id)), addQuotes(s.address),
			addQuotes(s.sslGrade), addQuotes(s.country), addQuotes(s.owner))
	return err
}

func (s *Server) deleteInDB(dbc interface{}) error {
	sqlStatement := `DELETE FROM server WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, s.id)
	return err
}

func (st *ServerTest) existsInDB(dbc interface{}) (bool, error) {
	sqlStatement := `SELECT id FROM serverTest WHERE id =$1;`
	row := QueryRow(dbc, sqlStatement, st.id)
	err := row.Scan(&st.id)
	switch err {
	case sql.ErrNoRows:
		return false, nil
	case nil:
		return true, nil
	default:
		return false, err
	}
}

func (st *ServerTest) selectInDB(dbc interface{}) error {
	sqlStatement := `SELECT domain, testHour, testInProgress, sslGrade, isDown FROM
	serverTest WHERE id=$1;`
	row := QueryRow(dbc, sqlStatement, st.id)
	err := row.Scan(&st.domain, &st.testHour, &st.testInProgress, &st.sslGrade, &st.isDown)
	switch err {
	case sql.ErrNoRows:
		return &CustomError{"No rows were returned."}
	case nil:
		return nil
	default:
		return err
	}
}

func (st *ServerTest) createInDB(dbc interface{}) error {
	sqlStatement := `INSERT INTO serverTest (domain, testHour, testInProgress, sslGrade, isDown)
	VALUES ($1, $2, $3, $4, $5) RETURNING id;`
	row := QueryRow(dbc, sqlStatement, st.domain, st.testHour, st.testInProgress, st.sslGrade, st.isDown)
	err := row.Scan(&st.id)
	return err
}

func (st *ServerTest) updateInDB(dbc interface{}) error {
	sqlStatement := `UPDATE serverTest SET domain = $2, testHour = $3, testInProgress = $4,
	sslGrade = $5, isDown = $6 WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, strconv.Itoa(st.id), addQuotes(st.domain),
		addQuotes(st.testHour), addQuotes(strconv.FormatBool(st.testInProgress)),
		addQuotes(st.sslGrade), addQuotes(strconv.FormatBool(st.isDown)))
	return err
}

func (st *ServerTest) updateHourInDb(dbc interface{}) error {
	sqlStatement := `UPDATE serverTest SET testHour = $2 WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, strconv.Itoa(st.id), addQuotes(st.testHour))
	return err
}

func (st *ServerTest) deleteInDB(dbc interface{}) error {
	sqlStatement := `DELETE FROM serverTest WHERE id = $1;`
	_, err := Exec(sqlStatement, st.id)
	return err
}

func ServerListFactory(idServerTest int, dbc interface{}) ([]Server, error) {
	var servers []Server
	sqlStatement := `SELECT id, address, sslGrade, country, owner FROM server
						WHERE serverTestId = $1;`
	rows, err := Query(dbc, sqlStatement, idServerTest)

	if err != nil {
		return servers, err
	}

	for rows.Next() {
		var s Server
		if err = rows.Scan(&s.id, &s.address, &s.sslGrade, &s.country, &s.owner); err != nil {
			return servers, err
		}
		servers = append(servers, s)
	}

	if err = rows.Err(); err != nil {
		return servers, err
	}

	return servers, err
}

func ServerTestListFactory(dbc interface{}) ([]ServerTest, error) {
	var serverTests []ServerTest
	sqlStatement := `SELECT id, domain, testHour, testInProgress, sslGrade, isDown FROM serverTest;`
	rows, err := Query(dbc, sqlStatement)

	if err != nil {
		return serverTests, err
	}

	for rows.Next() {
		var st ServerTest
		if err = rows.Scan(&st.id, &st.domain, &st.testHour, &st.testInProgress,
			&st.sslGrade, &st.isDown); err != nil {
			return serverTests, err
		}
		serverTests = append(serverTests, st)
	}

	if err = rows.Err(); err != nil {
		return serverTests, err
	}

	return serverTests, err
}

func (st *ServerTest) listServers(dbc interface{}) ([]Server, error) {
	return ServerListFactory(st.id, dbc)
}


// ENDPOINT domain_name -> test_complete

// 1) Search domain in database
// 2) Is there a server test in process (in database) with the same given domain?
// 2.1) YES: Is difference lower than 20 seconds?
// 2.1.1) YES: In the database, the server test will remain unchanged
//						Show the results using the ServerCompleteTest structure
// 2.1.2) NO: Connect to the SSLabs App, is the server test in process in SSLabs?
// 2.1.2.1) YES: Update the server test time in the database,
//						Show the results using the ServerCompleteTest structure
// 2.1.2.2) NO: Save the server test in the database
//						Show the results using the ServerCompleteTest structure
// 2.2) NO: Save the server test in the database
// 						Show the results using the ServerCompleteTest structure

// ENDPOINT -> domain_names

func(st *ServerTest) searchPendingTest(domainName string, testInProgress bool,
	dbc interface{}) error {
	sqlStatement := `SELECT id, domain, testHour, testInProgress, sslGrade, isDown
	FROM serverTest WHERE domain=$1 AND testInProgress=$2;`
	rows, err := Query(dbc, sqlStatement, domainName, testInProgress)
	if err != nil {
		return err
	}

	var serverTests []ServerTest
	for rows.Next() {
		var stTmp ServerTest
		if err = rows.Scan(&stTmp.id, &stTmp.domain, &stTmp.testHour, &stTmp.testInProgress,
			&stTmp.sslGrade, &stTmp.isDown); err != nil {
			return err
		}
		serverTests = append(serverTests, stTmp)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	if len(serverTests) > 0 {
			higher := Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
			higherId := 0
			for _, v := range serverTests {
				d := Parse("2006-01-02 15:04:05", v.testHour)
				if d.After(higher) {
					higher = d
					higherId = v.id
				}
			}
			v.id = higherId
			st.selectInDB(dbc)
	}

	return err
}

/*
type ServerTestComplete struct {
	testInProgress bool `json:"-"`
	servers []Server	`json:"servers"`
	serversChanged bool `json:"servers_changed"`
	sslGrade string `json:"ssl_grade"`
	previousSslGrade string `json:"previous_ssl_grade"`
	logo string `json:"logo"`
	title string `json:"title"`
	isDown bool	 `json:"is_down"`
}
 */
func MakeTestInDomain(domainName string,tx *sql.Tx) (ServerTestComplete, error) {
	//  testHour string, testInProgress bool,servers []Server, sslGrade string, isDown bool,
	returnTest := ServerTestComplete{}

	pendingTest := ServerTest{}
	serverCompleteTest := ServerCompleteTest

	// 1) Is there a server test in process with the same given domain?
	pendingTest.searchPendingTest(domainName, true, tx)
	if(pendingTest.id != 0) {
		pendingTestHour := Parse("2006-10-10 15:04:05", pendingTest.testHour)
		currentTestHour := Parse("2006-10-10 15:04:05", testHour)
		pendingTestHourA20 := pendingTestHour.Add(time.Second * 20)
		// 1.1) Is difference lower than 20 seconds?
		if(pendingTestHourA10.After(currentTestHour)) {
			// 1.1.1) In the database, data will remain unchanged
			// return the partial test.
			returnTest.testInProgress = testInProgress
			returnTest.servers = servers
			returnTest.sslGrade = servers
			returnTest.isDown = isDown
			return ServerTestComplete
		} else {
			// 1.1.2) N

		}

	}
	// 2.1.2) NO: Connect to the SSLabs App, is the server test in process in SSLabs?
	// 2.1.2.1) YES: Update the server test time in the database,
	//						Show the results using the ServerCompleteTest structure
	// 2.1.2.2) NO: Save the server test in the database
	//						Show the results using the ServerCompleteTest structure
	// 2.2) NO: Is there a past server test, ready, with the same domain?
	// 2.2.1) YES: Is difference lower than 20 seconds?
	// 2.2.1.1) YES: In the database, the data will remain unchanged, show the results
	// 2.2.1.2) NO:  Connect to the SSLabs App, is the server test in process in SSLabs?
	// 2.2.1.2.1) YES: Update the server test time in the database,
	//						Show the results using the ServerCompleteTest structure
	// 2.2.1.2.2) NO: Save the server test in the database
	//						Show the results using the ServerCompleteTest structure









}

func main() {
	// Connect to the "servers" database.
	db, err := initDB()
	err = crdb.ExecuteTx(context.Background(), db, nil, func(tx *sql.Tx) error {
			return ()
	})

	if err == nil {
		fmt.Println("Success")
	}	else {
		return err
	}
	defer db.Close()

}
