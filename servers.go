package main

import (
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/lib/pq"
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

// Server ...
type Server struct {
	id       int    // SERIAL PRIMARY KEY
	address  string // VARCHAR[16]
	sslGrade string // VARCHAR[5]
	country  string // VARCHAR[20]
	owner    string // VARCHAR[50]
}

// ServerTest ...
type ServerTest struct {
	id             int    // SERIAL PRIMARY KEY
	domain         string // VARCHAR[100]
	testHour       string // VARCHAR[30]
	testInProgress bool
	servers        []Server
	sslGrade       string
	isDown         bool
}

// DAOInterface ...
type DAOInterface interface {
	initTable(db *sql.DB) error
	selectInDB(db *sql.DB) error
	createInDB(db *sql.DB) error
	updateInDB(db *sql.DB) error
	deleteInDB(db *sql.DB) error
}

func (s *Server) initTable(db *sql.DB) error {
	sqlStatement := `CREATE TABLE server IF NOT EXISTS (id SERIAL PRIMARY KEY,
		serverTestId integer REFERENCES serverTest(id),	address VARCHAR[16],
		sslGrade VARCHAR[5], country VARCHAR[20], owner VARCHAR[50]);`
	_, err := db.Exec(sqlStatement)
	return err
}

func (s *Server) selectInDB(db *sql.DB) error {
	sqlStatement := "SELECT address, sslGrade, country, owner FROM server WHERE id=$1;"
	row := db.QueryRow(sqlStatement, s.id)
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

func (s *Server) createInDB(db *sql.DB) error {
	sqlStatement := `INSERT INTO server (address, sslGrade, country, owner)
	VALUES ($1, $2, $3, $4) RETURNING id;`
	row := db.QueryRow(sqlStatement, s.address, s.sslGrade, s.country, s.owner)
	err := row.Scan(&s.id)
	return err
}

func addQuotes(word string) string {
	return fmt.Sprintf(`'%v'`, word)
}

func (s *Server) updateInDB(db *sql.DB) error {
	sqlStatement := `UPDATE server SET address = $2, sslGrade = $3, country = $4, owner = $5 WHERE
	id = $1;`
	_, err := db.Exec(sqlStatement, addQuotes(strconv.Itoa(s.id)), addQuotes(s.address),
		addQuotes(s.sslGrade), addQuotes(s.country), addQuotes(s.owner))
	return err
}

func (s *Server) deleteInDB(db *sql.DB) error {
	sqlStatement := `DELETE FROM server WHERE id = $1;`
	_, err := db.Exec(sqlStatement, s.id)
	return err
}

func (st *ServerTest) initTable(db *sql.DB) error {
	sqlStatement := `CREATE TABLE serverTest IF NOT EXISTS (id SERIAL PRIMARY KEY,
		domain VARCHAR[100], testHour VARCHAR[30], testInProgress boolean,
		sslGrade VARCHAR[5], isDown boolean);`
	_, err := db.Exec(sqlStatement)
	return err
}

func (st *ServerTest) listServers(db *sql.DB) ([]Server, error) {
	var servers []Server
	sqlStatement := `SELECT id, address, sslGrade, country, owner FROM server
						WHERE serverTestId = $1;`
	rows, err := db.Query(sqlStatement, st.id)
	if err != nil {
		return servers, err
	}

	for rows.Next() {
		var s Server
		if err := rows.Scan(&s.id, &s.address, &s.sslGrade, &s.country, &s.owner); err != nil {
			return servers, err
		}
		servers = append(servers, s)
	}

	if err := rows.Err(); err != nil {
		return servers, err
	}

	return servers, err
}

func (st *ServerTest) selectInDB(db *sql.DB) error {
	sqlStatement := `SELECT domain, testHour, testInProgress, sslGrade, isDown FROM
	serverTest WHERE id=$1;`
	row := db.QueryRow(sqlStatement, st.id)
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
func (st *ServerTest) createInDB(db *sql.DB) error {
	sqlStatement := `INSERT INTO serverTest (domain, testHour, testInProgress, sslGrade, isDown)
	VALUES ($1, $2, $3, $4, $5) RETURNING id;`
	row := db.QueryRow(sqlStatement, st.domain, st.testHour, st.testInProgress, st.sslGrade, st.isDown)
	err := row.Scan(&st.id)
	return err
}
func (st *ServerTest) updateInDB(db *sql.DB) error {
	sqlStatement := `UPDATE serverTest SET domain = $2, testHour = $3, testInProgress = $4,
	sslGrade = $5, isDown = $6 WHERE id = $1`
	_, err := db.Exec(sqlStatement, strconv.Itoa(st.id), addQuotes(st.domain),
		addQuotes(st.testHour), strconv.FormatBool(st.testInProgress),
		addQuotes(st.sslGrade), strconv.FormatBool(st.isDown))
	return err
}

func (st *ServerTest) deleteInDB(db *sql.DB) error {
	sqlStatement := `DELETE FROM serverTest WHERE id = $1;`
	_, err := db.Exec(sqlStatement, st.id)
	return err
}

func main() {
	// Connect to the "servers" database.
	db, err := initDB()
	s := Server{}
	st := ServerTest{}

	defer db.Close()

}
