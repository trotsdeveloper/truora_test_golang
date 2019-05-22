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
	Id                   int    // SERIAL PRIMARY KEY
	Domain               string // VARCHAR[100]
	EvaluationHour       string // VARCHAR[30]
	EvaluationInProgress bool   // boolean
	Servers              []Server
	SslGrade             string // VARCHAR [5]
	IsDown               bool   // boolean
}

func (se *ServerEvaluation) Copy(sec ServerEvaluation) {
	se.Id = sec.Id
	se.Domain = sec.Domain
	se.EvaluationHour = sec.EvaluationHour
	se.EvaluationInProgress = sec.EvaluationInProgress
	se.Servers = sec.Servers
	se.SslGrade = sec.SslGrade
	se.IsDown = sec.IsDown
}

func CompareServer(s1, s2 Server) bool {
	return s1.Address == s2.Address && s1.SslGrade == s2.SslGrade &&
		s1.Country == s2.Country && s1.Owner == s2.Owner
}

func CompareServerList(sl1 []Server, sl2 []Server) bool {
	if len(sl1) == len(sl2) {
		if len(sl1) > 0 {
			for i := 0; i < len(sl1); i++ {
				if !CompareServer(sl1[i], sl2[i]) {
					return false
				}
			}
		}
		return true
	}
	return false
}

func CompareServerEvaluation(se1, se2 ServerEvaluation) bool {
	return se1.Domain == se2.Domain && se1.EvaluationHour == se2.EvaluationHour &&
		se1.EvaluationInProgress == se2.EvaluationInProgress && se1.SslGrade == se2.SslGrade &&
		se1.IsDown == se2.IsDown && CompareServerList(se1.Servers, se2.Servers)
}

// ServerEvaluationComplete ...
/*
type ServerEvaluationComplete struct {
	EvaluationInProgress   bool     `json:"-"`
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
					domain VARCHAR(100), EvaluationHour VARCHAR(30), EvaluationInProgress boolean,
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

func CleanDataInDB(dbc *sql.DB) error {
	sqlStatement1 := `DELETE FROM serverevaluation WHERE id > 0;`
	_, err := dbc.Exec(sqlStatement1)
	if err != nil {
		return err
	}
	sqlStatement2 := `DELETE FROM server WHERE id > 0;`
	_, err = dbc.Exec(sqlStatement2)
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
	sqlStatement := `SELECT domain, EvaluationHour, EvaluationInProgress, sslGrade, isDown FROM
	serverEvaluation WHERE id=$1;`
	row, err := QueryRow(dbc, sqlStatement, se.Id)
	err = row.Scan(&se.Domain, &se.EvaluationHour, &se.EvaluationInProgress, &se.SslGrade, &se.IsDown)
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
	sqlStatement := `INSERT INTO serverEvaluation (domain, EvaluationHour, EvaluationInProgress, sslGrade, isDown)
	VALUES ($1, $2, $3, $4, $5) RETURNING id;`
	row, err := QueryRow(dbc, sqlStatement, se.Domain, se.EvaluationHour, se.EvaluationInProgress, se.SslGrade, se.IsDown)
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
	sqlStatement := `UPDATE serverEvaluation SET domain = $2, EvaluationHour = $3, EvaluationInProgress = $4,
	sslGrade = $5, isDown = $6 WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, se.Id, se.Domain,
		se.EvaluationHour, se.EvaluationInProgress, se.SslGrade, se.IsDown)
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
	sqlStatement := `UPDATE serverEvaluation SET EvaluationHour = $2 WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, se.Id, se.EvaluationHour)
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
	sqlStatement := `SELECT id, domain, EvaluationHour, EvaluationInProgress, sslGrade, isDown FROM serverEvaluation;`
	rows, err := Query(dbc, sqlStatement)

	if err != nil {
		return serverEvaluations, err
	}

	for rows.Next() {
		var se ServerEvaluation
		if err = rows.Scan(&se.Id, &se.Domain, &se.EvaluationHour, &se.EvaluationInProgress,
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

func SearchLastEvaluations(domainName string, EvaluationInProgress bool,
	lastNItems int, dbc interface{}) (lastServerEvaluations []ServerEvaluation, err error) {

	if lastNItems <= 0 {
		err = &CustomError{`lastNItems needs to be a positive integer`}
		return
	}

	var serverEvaluations []ServerEvaluation
	sqlStatement := `SELECT id, domain, EvaluationHour, EvaluationInProgress, sslGrade, isDown
		FROM serverEvaluation WHERE domain = $1 AND EvaluationInProgress = $2;`
	var rows *sql.Rows
	rows, err = Query(dbc, sqlStatement, domainName, EvaluationInProgress)

	if err != nil {
		return
	}

	for rows.Next() {
		var seTmp ServerEvaluation
		if err = rows.Scan(&seTmp.Id, &seTmp.Domain, &seTmp.EvaluationHour, &seTmp.EvaluationInProgress,
			&seTmp.SslGrade, &seTmp.IsDown); err != nil {
			return
		}
		serverEvaluations = append(serverEvaluations, seTmp)
	}

	if err = rows.Err(); err != nil {
		return
	}

	if len(serverEvaluations) == 0 {
		return
	}

	lowestBound := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	highest := lowestBound
	highestID := 0

	for _, v := range serverEvaluations {
		var d time.Time
		d, err = time.Parse(time.RFC3339, v.EvaluationHour)
		if d.After(highest) {
			highest = d
			highestID = v.Id
		}
	}

	se1 := ServerEvaluation{Id: highestID}
	se1.selectInDB(dbc)
	lastServerEvaluations = append(lastServerEvaluations, se1)

	for i := 2; i <= lastNItems; i++ {
		if len(serverEvaluations)-i < 0 {
			return
		}
		higher := lowestBound
		higherID := 0

		for _, v := range serverEvaluations {
			var d time.Time
			d, err = time.Parse(time.RFC3339, v.EvaluationHour)
			if d.After(higher) && d.Before(highest) {
				higher = d
				higherID = v.Id
			}
		}

		seTmp := ServerEvaluation{Id: higherID}
		seTmp.selectInDB(dbc)
		lastServerEvaluations = append(lastServerEvaluations, seTmp)
		highest = higher
	}
	return
}

func (se *ServerEvaluation) SearchLastEvaluation(domainName string, EvaluationInProgress bool,
	dbc interface{}) error {
	lastEvaluations, err := SearchLastEvaluations(domainName, EvaluationInProgress, 1, dbc)
	if err == nil {
		if len(lastEvaluations) > 0 {
			se.Copy(lastEvaluations[0])
		}
	}
	return err
}

func MakeEvaluationInDomain(domainName string, currentHour time.Time, makeEvaluation func(string) (ServerEvaluation, error),
	dbc interface{}) (st ServerEvaluation, err error) {

	// 1) In the database, is there a server Evaluation in process
	// with the same given domain?
	pendingEvaluation := ServerEvaluation{}
	err = pendingEvaluation.SearchLastEvaluation(domainName, true, dbc)
	if err != nil {
		fmt.Println(err)
		return
	}
	if pendingEvaluation.Id != 0 {
		// 1.1) YES: Is difference between the current hour
		// and the pending Evaluation lower than 20 seconds?
		var pendingEvaluationHour time.Time
		pendingEvaluationHour, err = time.Parse(time.RFC3339, pendingEvaluation.EvaluationHour)
		if err != nil {
			return
		}
		pendingEvaluationHourA20 := pendingEvaluationHour.Add(time.Second * 20)
		if pendingEvaluationHourA20.After(currentHour) {
			// 1.1.1) YES: In the database, data will remain unchanged
			// return the pending Evaluation
			return pendingEvaluation, err
		} else {
			// 1.1.2) Update the hour of the pending Evaluation with the current hour
			pendingEvaluation.EvaluationHour = currentHour.Format(time.RFC3339)
			err = pendingEvaluation.updateHourInDb(dbc)
			if err != nil {
				return
			}

			var currentEvaluation ServerEvaluation
			currentEvaluation, err = makeEvaluation(domainName)
			if err != nil {
				return
			}
			// 1.1.2) NO: In SSLabs, Is the server Evaluation in process?
			if currentEvaluation.EvaluationInProgress {
				// 1.1.2.1) YES: Return the pending Evaluation with the new hour
				return pendingEvaluation, err
			} else {
				// 1.1.2.2) NO: Update the pending Evaluation in the database, with
				// the information of the current Evaluation
				currentEvaluation.Id = pendingEvaluation.Id
				err = currentEvaluation.updateInDB(dbc)
				return currentEvaluation, err
			}
		}
	} else {
		// 1.2) NO: Is there a past server Evaluation, ready, with the same given domain?
		pastEvaluation := ServerEvaluation{}
		err = pastEvaluation.SearchLastEvaluation(domainName, false, dbc)
		if pastEvaluation.Id != 0 {
			// 1.2.1) YES: Is difference lower than 20 seconds?
			var pastEvaluationHour time.Time
			pastEvaluationHour, err = time.Parse(time.RFC3339, pastEvaluation.EvaluationHour)
			if err != nil {
				return
			}
			pastEvaluationHourA20 := pastEvaluationHour.Add(time.Second * 20)
			if pastEvaluationHourA20.After(currentHour) {
				// 1.2.1.1) YES: In the database, data will remain unchanged,
				// return the past Evaluation
				return pastEvaluation, err
			} else {
				// 1.2.1.2) NO: Make a server Evaluation from SSLabs, save it in DB.
				var currentEvaluation ServerEvaluation
				currentEvaluation, err = makeEvaluation(domainName)
				if err != nil {
					return
				}
				err = currentEvaluation.createInDB(dbc)
				return currentEvaluation, err
			}
		} else {
			// 1.2.2) NO: Make a server Evaluation from SSLabs, save it in DB.
			var currentEvaluation ServerEvaluation
			currentEvaluation, err = makeEvaluation(domainName)
			if err != nil {
				return
			}
			err = currentEvaluation.createInDB(dbc)
			return currentEvaluation, err
		}
	}
}

func main() {}
