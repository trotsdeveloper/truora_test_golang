package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

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
	psqlInfo := fmt.Sprintf("postgresql://%s@%s:%s/%s?ssl=%s&sslmode=%s&sslrootcert=%s&sslkey=%s&sslcert=%s",
		USER, HOST, strconv.Itoa(PORT), DATABASE, strconv.FormatBool(SSL), SSLMODE, SSLROOTCERT, SSLKEY, SSLCERT)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Println("AAA")
	}
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
	Id       int    `json:"-"`         // SERIAL PRIMARY KEY
	Address  string `json:"address"`   // VARCHAR[16]
	SslGrade string `json:"ssl_grade"` // VARCHAR[5]
	Country  string `json:"country"`   // VARCHAR[20]
	Owner    string `json:"owner"`     // VARCHAR[50]
}

// ServerEvaluation ...
type ServerEvaluation struct {
	Id             int    // SERIAL PRIMARY KEY
	Domain         string // VARCHAR[100]
	TestHour       string // VARCHAR[30]
	TestInProgress bool   // boolean
	Servers        []Server
	SslGrade       string // VARCHAR [5]
	IsDown         bool   // boolean
}

func CompareServerEvaluation(se1, se2 ServerEvaluation) bool {
	return se1.Domain == se2.Domain && se1.TestHour == se2.TestHour &&
		se1.TestInProgress == se2.TestInProgress && se1.SslGrade == se2.SslGrade &&
		se1.IsDown == se2.IsDown
}

// ServerEvaluationComplete ...
/*
type ServerEvaluationComplete struct {
	testInProgress   bool     `json:"-"`
	servers          []Server `json:"servers"`
	serversChanged   bool     `json:"servers_changed"`
	sslGrade         string   `json:"ssl_grade"`
	previousSslGrade string   `json:"previous_ssl_grade"`
	logo             string   `json:"logo"`
	title            string   `json:"title"`
	isDown           bool     `json:"is_down"`
}
*/

// DAO..
type DAO interface {
	existsInDb(dbc interface{}) (bool, error)
	selectInDB(dbc interface{}) error
	createInDB(dbc interface{}) error
	updateInDB(dbc interface{}) error
	deleteInDB(dbc interface{}) error
}

func Exec(dbc interface{}, sqlString string, args ...interface{}) (r sql.Result, err error) {
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
		r = v.QueryRow(sqlString, args...)
	case *sql.Tx:
		r = v.QueryRow(sqlString, args...)
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

func InitServerEvaluationTable(dbc *sql.DB) error {
	sqlStatement := `CREATE TABLE serverEvaluation (id SERIAL PRIMARY KEY,
					domain VARCHAR(100), testHour VARCHAR(30), testInProgress boolean,
					sslGrade VARCHAR(5), isDown boolean);`
	_, err := dbc.Exec(sqlStatement)
	return err
}
func InitServerTable(dbc *sql.DB) error {
	sqlStatement := `CREATE TABLE server (id SERIAL PRIMARY KEY,
		serverEvaluationId integer,	address VARCHAR(16), sslGrade VARCHAR(5), country VARCHAR(20),
		owner VARCHAR(50), FOREIGN KEY(serverEvaluationId) REFERENCES serverEvaluation(id));`
	_, err := dbc.Exec(sqlStatement)
	return err
}

func DropServerTable(dbc *sql.DB) error {
	sqlStatement := `DROP TABLE server;`
	_, err := dbc.Exec(sqlStatement)
	return err
}

func DropServerEvaluationTable(dbc *sql.DB) error {
	sqlStatement := `DROP TABLE serverEvaluation;`
	_, err := dbc.Exec(sqlStatement)
	return err
}

func (s *Server) existsInDB(dbc interface{}) (bool, error) {
	sqlStatement := `SELECT id FROM server WHERE id =$1;`
	row, err := QueryRow(dbc, sqlStatement, s.Id)
	err = row.Scan(&s.Id)
	switch err {
	case sql.ErrNoRows:
		return false, nil
	case nil:
		return true, nil
	default:
		return false, err
	}
}

//
func (s *Server) selectInDB(dbc interface{}) error {
	sqlStatement := `SELECT address, sslGrade, country, owner FROM server WHERE id=$1;`
	row, err := QueryRow(dbc, sqlStatement, s.Id)
	err = row.Scan(&s.Address, &s.SslGrade, &s.Country, &s.Owner)
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
	row, err := QueryRow(dbc, sqlStatement, s.Address, s.SslGrade, s.Country, s.Owner)
	err = row.Scan(&s.Id)
	return err
}

func (s *Server) updateServerEvaluationInDB(serverEvaluationId int, dbc interface{}) error {
	sqlStatement := `UPDATE server SET serverEvaluationId = $2 WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, s.Id, serverEvaluationId)
	return err
}

func (s *Server) updateInDB(dbc interface{}) error {
	sqlStatement := `UPDATE server SET address = $2, sslGrade = $3, country = $4,
	owner = $5 WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, s.Id, s.Address,
		s.SslGrade, s.Country, s.Owner)
	return err
}

func (s *Server) deleteInDB(dbc interface{}) error {
	sqlStatement := `DELETE FROM server WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, s.Id)
	return err
}

func (se *ServerEvaluation) existsInDB(dbc interface{}) (bool, error) {
	sqlStatement := `SELECT id FROM serverEvaluation WHERE id =$1;`
	row, err := QueryRow(dbc, sqlStatement, se.Id)
	err = row.Scan(&se.Id)
	switch err {
	case sql.ErrNoRows:
		return false, nil
	case nil:
		return true, nil
	default:
		return false, err
	}
}

func (se *ServerEvaluation) selectInDB(dbc interface{}) error {
	sqlStatement := `SELECT domain, testHour, testInProgress, sslGrade, isDown FROM
	serverEvaluation WHERE id=$1;`
	row, err := QueryRow(dbc, sqlStatement, se.Id)
	err = row.Scan(&se.Domain, &se.TestHour, &se.TestInProgress, &se.SslGrade, &se.IsDown)
	switch err {
	case sql.ErrNoRows:
		return &CustomError{"No rows were returned."}
	case nil:
		return nil
	default:
		return err
	}
}

func (se *ServerEvaluation) createInDB(dbc interface{}) error {
	sqlStatement := `INSERT INTO serverEvaluation (domain, testHour, testInProgress, sslGrade, isDown)
	VALUES ($1, $2, $3, $4, $5) RETURNING id;`
	row, err := QueryRow(dbc, sqlStatement, se.Domain, se.TestHour, se.TestInProgress, se.SslGrade, se.IsDown)
	err = row.Scan(&se.Id)
	if err != nil {
		return err
	}
	if len(se.Servers) > 0 {
		for _, v := range se.Servers {
			if err = v.createInDB(dbc); err != nil {
				return err
			}
			if err = v.updateServerEvaluationInDB(se.Id, dbc); err != nil {
				return err
			}
		}
	}
	return err
}

func (se *ServerEvaluation) updateInDB(dbc interface{}) error {
	sqlStatement := `UPDATE serverEvaluation SET domain = $2, testHour = $3, testInProgress = $4,
	sslGrade = $5, isDown = $6 WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, se.Id, se.Domain,
		se.TestHour, se.TestInProgress, se.SslGrade, se.IsDown)
	if err != nil {
		return err
	}
	if len(se.Servers) > 0 {
		err = se.deleteAllServersInDB(dbc)
		if err != nil {
			return err
		}
		for _, v := range se.Servers {
			if err = v.createInDB(dbc); err != nil {
				return err
			}
			if err = v.updateServerEvaluationInDB(se.Id, dbc); err != nil {
				return err
			}
		}
	}
	return err
}

func (se *ServerEvaluation) updateHourInDb(dbc interface{}) error {
	sqlStatement := `UPDATE serverEvaluation SET testHour = $2 WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, se.Id, se.TestHour)
	return err
}

func (se *ServerEvaluation) deleteAllServersInDB(dbc interface{}) error {
	sqlStatement := `DELETE FROM server WHERE serverEvaluationId = $1;`
	_, err := Exec(dbc, sqlStatement, se.Id)
	return err
}

func (se *ServerEvaluation) deleteInDB(dbc interface{}) error {
	sqlStatement := `DELETE FROM serverEvaluation WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, se.Id)
	return err
}

func ServerListFactory(idServerEvaluation int, dbc interface{}) ([]Server, error) {
	var servers []Server
	sqlStatement := `SELECT id, address, sslGrade, country, owner FROM server
						WHERE serverEvaluationId = $1;`
	rows, err := Query(dbc, sqlStatement, idServerEvaluation)

	if err != nil {
		return servers, err
	}

	for rows.Next() {
		var s Server
		if err = rows.Scan(&s.Id, &s.Address, &s.SslGrade, &s.Country, &s.Owner); err != nil {
			return servers, err
		}
		servers = append(servers, s)
	}

	if err = rows.Err(); err != nil {
		return servers, err
	}

	return servers, err
}

func ServerEvaluationListFactory(dbc interface{}) ([]ServerEvaluation, error) {
	var serverEvaluations []ServerEvaluation
	sqlStatement := `SELECT id, domain, testHour, testInProgress, sslGrade, isDown FROM serverEvaluation;`
	rows, err := Query(dbc, sqlStatement)

	if err != nil {
		return serverEvaluations, err
	}

	for rows.Next() {
		var se ServerEvaluation
		if err = rows.Scan(&se.Id, &se.Domain, &se.TestHour, &se.TestInProgress,
			&se.SslGrade, &se.IsDown); err != nil {
			return serverEvaluations, err
		}
		serverEvaluations = append(serverEvaluations, se)
	}

	if err = rows.Err(); err != nil {
		return serverEvaluations, err
	}

	return serverEvaluations, err
}

func (se *ServerEvaluation) listServers(dbc interface{}) {
	se.Servers, _ = ServerListFactory(se.Id, dbc)
}

func (se *ServerEvaluation) searchPendingTest(domainName string, testInProgress bool,
	dbc interface{}) error {
	sqlStatement := `SELECT id, domain, testHour, testInProgress, sslGrade, isDown
	FROM serverEvaluation WHERE domain = $1 AND testInProgress = $2;`
	rows, err := Query(dbc, sqlStatement, domainName, testInProgress)
	if err != nil {
		return err
	}

	var serverEvaluations []ServerEvaluation
	for rows.Next() {
		var seTmp ServerEvaluation
		if err = rows.Scan(&seTmp.Id, &seTmp.Domain, &seTmp.TestHour, &seTmp.TestInProgress,
			&seTmp.SslGrade, &seTmp.IsDown); err != nil {
			return err
		}
		serverEvaluations = append(serverEvaluations, seTmp)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	if len(serverEvaluations) > 0 {
		higher := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		higherId := 0
		for _, v := range serverEvaluations {
			var d time.Time
			d, err = time.Parse(time.RFC3339, v.TestHour)
			if d.After(higher) {
				higher = d
				higherId = v.Id
			}
		}
		se.Id = higherId
		se.selectInDB(dbc)
	}
	return err
}

func MakeEvaluationInDomain(domainName string, currentHour time.Time, makeTest func(string) (ServerEvaluation, error),
	dbc interface{}) (st ServerEvaluation, err error) {

	// 1) In the database, is there a server test in process
	// with the same given domain?
	pendingTest := ServerEvaluation{}
	err = pendingTest.searchPendingTest(domainName, true, dbc)
	if err != nil {
		fmt.Println(err)
		return
	}
	if pendingTest.Id != 0 {
		// 1.1) YES: Is difference between the current hour
		// and the pending test lower than 20 seconds?
		var pendingTestHour time.Time
		pendingTestHour, err = time.Parse(time.RFC3339, pendingTest.TestHour)
		if err != nil {
			return
		}
		pendingTestHourA20 := pendingTestHour.Add(time.Second * 20)
		if pendingTestHourA20.After(currentHour) {
			// 1.1.1) YES: In the database, data will remain unchanged
			// return the pending test
			return pendingTest, err
		} else {
			// 1.1.2) Update the hour of the pending test with the current hour
			pendingTest.TestHour = currentHour.Format(time.RFC3339)
			err = pendingTest.updateHourInDb(dbc)
			if err != nil {
				return
			}

			var currentTest ServerEvaluation
			currentTest, err = makeTest(domainName)
			if err != nil {
				return
			}
			// 1.1.2) NO: In SSLabs, Is the server test in process?
			if currentTest.TestInProgress {
				// 1.1.2.1) YES: Return the pending test with the new hour
				return pendingTest, err
			} else {
				// 1.1.2.2) NO: Update the pending test in the database, with
				// the information of the current test
				currentTest.Id = pendingTest.Id
				err = currentTest.updateInDB(dbc)
				return currentTest, err
			}
		}
	} else {
		// 1.2) NO: Is there a past server test, ready, with the same given domain?
		pastTest := ServerEvaluation{}
		err = pastTest.searchPendingTest(domainName, false, dbc)
		if pastTest.Id != 0 {
			// 1.2.1) YES: Is difference lower than 20 seconds?
			var pastTestHour time.Time
			pastTestHour, err = time.Parse(time.RFC3339, pastTest.TestHour)
			if err != nil {
				return
			}
			pastTestHourA20 := pastTestHour.Add(time.Second * 20)
			if pastTestHourA20.After(currentHour) {
				// 1.2.1.1) YES: In the database, data will remain unchanged,
				// return the past test
				return pastTest, err
			} else {
				// 1.2.1.2) NO: Make a server test from SSLabs, save it in DB.
				var currentTest ServerEvaluation
				currentTest, err = makeTest(domainName)
				if err != nil {
					return
				}
				err = currentTest.createInDB(dbc)
				return currentTest, err
			}
		} else {
			// 1.2.2) NO: Make a server test from SSLabs, save it in DB.
			var currentTest ServerEvaluation
			currentTest, err = makeTest(domainName)
			if err != nil {
				return
			}
			err = currentTest.createInDB(dbc)
			return currentTest, err
		}
	}
}

func main() {}
