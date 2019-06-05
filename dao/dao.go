// Package for the declaration of the main structures
// for data handling in the project.
package dao

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	_ "github.com/lib/pq"
)

// Postgresql Settings
const (
	USER        = "manuelams"
	PASSWORD    = "11235813"
	HOST        = "localhost"
	PORT        = 26257
	DATABASE    = "servers"
	SSL         = true
	SSLMODE     = "require"
	SSLROOTCERT = "../../certs/ca.crt"
	SSLKEY      = "../../certs/client.manuelams.key"
	SSLCERT     = "../../certs/client.manuelams.crt"
)

// Declaration of global DB controller
var (
	DBConf *sql.DB
)

// Function for the inicialization of the global DB controller
func InitDB() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?ssl=%s&sslmode=%s&sslrootcert=%s&sslkey=%s&sslcert=%s",
		USER, PASSWORD, HOST, strconv.Itoa(PORT), DATABASE, strconv.FormatBool(SSL), SSLMODE, SSLROOTCERT, SSLKEY, SSLCERT)
	db, err := sql.Open("postgres", psqlInfo)
	return db, err
}

// Auxiliar function to add quotes to a string
func addQuotes(word string) string {
	return fmt.Sprintf(`'%v'`, word)
}

// Server - Struct for the representation of servers in a domain.
// It contains the ip and ssl from the sslab test, and the
// country and owner info from whois test
type Server struct {
	Id       int    `json:"-"`         // SERIAL PRIMARY KEY
	Address  string `json:"address"`   // VARCHAR[16]
	SslGrade string `json:"ssl_grade"` // VARCHAR[5]
	Country  string `json:"country"`   // VARCHAR[20]
	Owner    string `json:"owner"`     // VARCHAR[50]
}

// DomainEvaluation: Struct for the representation of a SSLabs test in
// a specific domain.
type DomainEvaluation struct {
	Id                   int      `json:"-"`           // SERIAL PRIMARY KEY
	Domain               string   `json:"domain"`      // VARCHAR[100]
	EvaluationHour       string   `json:"hour"`        // VARCHAR[30]
	EvaluationInProgress bool     `json:"in_progress"` // boolean
	Servers              []Server `json:"-"`
	SslGrade             string   `json:"ssl_grade"` // VARCHAR [5]
	Logo                 string   `json:"logo"`      // VARCHAR[20]
	Title                string   `json:"title"`     // VARCHAR[20]
	IsDown               bool     `json:"is_down"`   // boolean
}

// DomainEvaluationComplete: Struct for the representation of all data
// about a specific domain using different scrapers and db queries.
// It's used for displaying information in the API
type DomainEvaluationComplete struct {
	Servers          []Server `json:"servers"`
	ServersChanged   bool     `json:"servers_changed"`
	SslGrade         string   `json:"ssl_grade"` // VARCHAR [5]
	PreviousSslGrade string   `json:"previous_ssl_grade"`
	Logo             string   `json:"logo"`
	Title            string   `json:"title"`
	IsDown           bool     `json:"is_down"` // boolean
}

// Function self-explanatory, it allows to copy information from one structure
// for database manipulation to another structure for data visualization
func (dec *DomainEvaluationComplete) Copy(de DomainEvaluation) {
	dec.Servers = de.Servers
	dec.SslGrade = de.SslGrade
	dec.IsDown = de.IsDown
	dec.Logo = de.Logo
	dec.Title = de.Title
}

// Compares two server structures
func CompareServer(s1, s2 Server) bool {
	return s1.Address == s2.Address && s1.SslGrade == s2.SslGrade &&
		s1.Country == s2.Country && s1.Owner == s2.Owner
}

// Compare two server lists
func CompareServerList(sl1 []Server, sl2 []Server) (b bool) {
	b = true
	if len(sl1) == len(sl2) {
		if len(sl1) > 0 {
			for i := 0; i < len(sl1); i++ {
				if !CompareServer(sl1[i], sl2[i]) {
					b = false
					return
				}
			}
		}
		return
	}
	return
}

// Compare two DomainEvaluation structures.
func CompareDomainEvaluation(de1, de2 DomainEvaluation) bool {
	return de1.Domain == de2.Domain && de1.EvaluationHour == de2.EvaluationHour &&
		de1.EvaluationInProgress == de2.EvaluationInProgress && de1.SslGrade == de2.SslGrade &&
		de1.IsDown == de2.IsDown && CompareServerList(de1.Servers, de2.Servers)
}

// DAO interface: Declaration of interface for data manipulation in database
type DAO interface {
	SelectInDB(dbc interface{}) error
	CreateInDB(dbc interface{}) error
	UpdateInDB(dbc interface{}) error
	DeleteInDB(dbc interface{}) error
}

// Function for calling Exec in either a *sql.DB or a *sql.Tx controller.
func Exec(dbc interface{}, sqlString string, args ...interface{}) (r sql.Result, err error) {
	switch v := dbc.(type) {
	case *sql.DB:
		r, err = v.Exec(sqlString, args...)
	case *sql.Tx:
		r, err = v.Exec(sqlString, args...)
	default:
		err = errors.New("No valid DB Controller")
	}
	return
}

// Funcion for calling QueryRow in either a *sql.DB or a *sql.Tx controller.
func QueryRow(dbc interface{}, sqlString string, args ...interface{}) (r *sql.Row, err error) {
	switch v := dbc.(type) {
	case *sql.DB:
		r = v.QueryRow(sqlString, args...)
	case *sql.Tx:
		r = v.QueryRow(sqlString, args...)
	default:
		err = errors.New("No valid DB Controller")
	}
	return
}

// Function for calling Query in either a *sql.DB or a *sql.Tx controller.
func Query(dbc interface{}, sqlString string, args ...interface{}) (r *sql.Rows, err error) {
	switch v := dbc.(type) {
	case *sql.DB:
		r, err = v.Query(sqlString, args...)
	case *sql.Tx:
		r, err = v.Query(sqlString, args...)
	default:
		err = errors.New("No valid DB Controller")
	}
	return
}

// Function for creating the domainEvaluation table.
// WARNING: USAGE ONLY IN TESTS ENVIRONMENTS, NOT RECOMMENDED FOR PRODUCTION
func InitDomainEvaluationTable(dbc *sql.DB) error {
	sqlStatement := `CREATE TABLE domainEvaluation (id SERIAL PRIMARY KEY,
					domain VARCHAR(100), EvaluationHour VARCHAR(30), EvaluationInProgress boolean,
					sslGrade VARCHAR(5), logo VARCHAR(80), title VARCHAR(80), isDown boolean);`
	_, err := dbc.Exec(sqlStatement)
	return err
}

// Function for creating the server table.
// WARNING: USAGE ONLY IN TESTS ENVIRONMENTS, NOT RECOMMENDED FOR PRODUCTION
func InitServerTable(dbc *sql.DB) error {
	sqlStatement := `CREATE TABLE server (id SERIAL PRIMARY KEY,
		domainEvaluationId integer,	address VARCHAR(50), sslGrade VARCHAR(5), country VARCHAR(20),
		owner VARCHAR(50), FOREIGN KEY(domainEvaluationId) REFERENCES domainEvaluation(id));`
	_, err := dbc.Exec(sqlStatement)
	return err
}

// Function for droping the server table.
// WARNING: USAGE ONLY IN TESTS ENVIRONMENTS, NOT RECOMMENDED FOR PRODUCTION
func DropServerTable(dbc *sql.DB) error {
	sqlStatement := `DROP TABLE server;`
	_, err := dbc.Exec(sqlStatement)
	return err
}

// Function for droping the domainEvaluation table.
// WARNING: USAGE ONLY IN TESTS ENVIRONMENTS, NOT RECOMMENDED FOR PRODUCTION
func DropDomainEvaluationTable(dbc *sql.DB) error {
	sqlStatement := `DROP TABLE domainEvaluation;`
	_, err := dbc.Exec(sqlStatement)
	return err
}

// Function for cleaning data in DB.
// WARNING: USABLE BUT NOT RECOMMENDED FOR PRODUCTION
func CleanDataInDB(dbc *sql.DB) error {
	sqlStatement1 := `DELETE FROM server WHERE id > 0;`
	_, err := dbc.Exec(sqlStatement1)
	if err != nil {
		return err
	}
	sqlStatement2 := `DELETE FROM domainevaluation WHERE id > 0;`
	_, err = dbc.Exec(sqlStatement2)
	return err
}

// SelectInDB
// Implementation of the method SelectInDB from the DAO interface
// for the Server structure.
func (s *Server) SelectInDB(dbc interface{}) error {
	sqlStatement := `SELECT address, sslGrade, country, owner FROM server WHERE id=$1;`
	row, err := QueryRow(dbc, sqlStatement, s.Id)
	err = row.Scan(&s.Address, &s.SslGrade, &s.Country, &s.Owner)
	switch err {
	case sql.ErrNoRows:
		return errors.New("No rows were returned.")
	case nil:
		return nil
	default:
		return err
	}
}

// CreateInDB
// Implementation of the method CreateInDB from the DAO interface
// for the Server structure.
func (s *Server) CreateInDB(dbc interface{}) error {
	sqlStatement := `INSERT INTO server (address, sslGrade, country, owner)
	VALUES ($1, $2, $3, $4) RETURNING id;`
	row, err := QueryRow(dbc, sqlStatement, s.Address, s.SslGrade, s.Country, s.Owner)
	err = row.Scan(&s.Id)
	return err
}

// updateDomainEvaluationInDB
// Method for updating the domainEvaluationId of a server structure.
// Since the server structure doesn't store the id, the update is just in
// the database
func (s *Server) updateDomainEvaluationInDB(domainEvaluationId int, dbc interface{}) error {
	sqlStatement := `UPDATE server SET domainEvaluationId = $2 WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, s.Id, domainEvaluationId)
	return err
}

// UpdateInDB
// Implementation of the method UpdateInDB from the DAO interface
// for the Server structure.
func (s *Server) UpdateInDB(dbc interface{}) error {
	sqlStatement := `UPDATE server SET address = $2, sslGrade = $3, country = $4,
	owner = $5 WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, s.Id, s.Address,
		s.SslGrade, s.Country, s.Owner)
	return err
}

// DeleteInDB
// Implementation of the method DeleteInDB from the DAO interface
// for the Server structure.
func (s *Server) DeleteInDB(dbc interface{}) error {
	sqlStatement := `DELETE FROM server WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, s.Id)
	return err
}

// SelectInDB
// Implementation of the method SelectInDB from the DAO interface
// for the DomainEvaluation structure.
func (de *DomainEvaluation) SelectInDB(dbc interface{}) error {
	sqlStatement := `SELECT domain, EvaluationHour, EvaluationInProgress, sslGrade,
	logo, title, isDown FROM domainEvaluation WHERE id=$1;`
	row, err := QueryRow(dbc, sqlStatement, de.Id)
	err = row.Scan(&de.Domain, &de.EvaluationHour, &de.EvaluationInProgress, &de.SslGrade,
		&de.Logo, &de.Title, &de.IsDown)
	switch err {
	case sql.ErrNoRows:
		return errors.New("No rows were returned.")
	case nil:
		return nil
	default:
		return err
	}
}

// CreateInDB
// Implementation of the method CreateInDB from the DAO interface
// for the DomainEvaluation structure.
// In case the lists of servers in the structure is not empty,
// the method create all the servers in the list in the db.
func (de *DomainEvaluation) CreateInDB(dbc interface{}) error {
	sqlStatement := `INSERT INTO domainEvaluation (domain, EvaluationHour, EvaluationInProgress, sslGrade,
		logo, title, isDown) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;`
	row, err := QueryRow(dbc, sqlStatement, de.Domain, de.EvaluationHour,
		de.EvaluationInProgress, de.SslGrade, de.Logo, de.Title, de.IsDown)
	err = row.Scan(&de.Id)
	if err != nil {
		return err
	}

	//
	if len(de.Servers) > 0 {
		for _, v := range de.Servers {
			if err = v.CreateInDB(dbc); err != nil {
				return err
			}
			if err = v.updateDomainEvaluationInDB(de.Id, dbc); err != nil {
				return err
			}
		}
	}
	return err
}

// UpdateInDB
// Implementation of the method UpdateInDB from the DAO interface
// for the DomainEvaluation structure.

// In case the lists of servers in the structure is not empty,
// the method delete all servers in the db and create the servers
// in the list.

// Partial updates of the server list are not implemented in this method
func (de *DomainEvaluation) UpdateInDB(dbc interface{}) error {
	sqlStatement := `UPDATE domainEvaluation SET domain = $2, EvaluationHour = $3, EvaluationInProgress = $4,
	sslGrade = $5, logo = $6, title = $7, isDown = $8 WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, de.Id, de.Domain, de.EvaluationHour,
		de.EvaluationInProgress, de.SslGrade, de.Logo, de.Title, de.IsDown)
	if err != nil {
		return err
	}
	if len(de.Servers) > 0 {
		err = de.deleteAllServersInDB(dbc)
		if err != nil {
			return err
		}
		for _, v := range de.Servers {
			if err = v.CreateInDB(dbc); err != nil {
				return err
			}
			if err = v.updateDomainEvaluationInDB(de.Id, dbc); err != nil {
				return err
			}
		}
	}
	return err
}

// UpdateLogoInDb
// Method for updating only the logo in a domainEvaluation structure.
func (de *DomainEvaluation) UpdateLogoInDb(dbc interface{}) error {
	sqlStatement := `UPDATE domainEvaluation SET logo = $2 WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, de.Id, de.Logo)
	return err
}

// UpdateTitleInDb
// Method for updating only the title in a domainEvaluation structure.
func (de *DomainEvaluation) UpdateTitleInDb(dbc interface{}) error {
	sqlStatement := `UPDATE domainEvaluation SET title = $2 WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, de.Id, de.Title)
	return err
}

// UpdateHourInDb
// Method for updating only the hour in a domainEvaluation structure.
func (de *DomainEvaluation) UpdateHourInDb(dbc interface{}) error {
	sqlStatement := `UPDATE domainEvaluation SET EvaluationHour = $2 WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, de.Id, de.EvaluationHour)
	return err
}

// deleteAllServersInDB
// Method for deleting all servers in db corresponding to a specific
// domainEvaluation structure.
func (de *DomainEvaluation) deleteAllServersInDB(dbc interface{}) error {
	sqlStatement := `DELETE FROM server WHERE domainEvaluationId = $1;`
	_, err := Exec(dbc, sqlStatement, de.Id)
	return err
}

// DeleteInDB
// Implementation of the method DeleteInDB from the DAO interface for
// the domainEvaluation structure.
func (de *DomainEvaluation) DeleteInDB(dbc interface{}) error {
	sqlStatement := `DELETE FROM domainEvaluation WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, de.Id)
	if err == nil {
		return de.deleteAllServersInDB(dbc)
	}
	return err
}

// ListDomainEvaluations
// Function for listing all domain evaluations in the database, without any
// kind of grouping
func ListDomainEvaluations(dbc interface{}) ([]DomainEvaluation, error) {
	var domainEvaluations []DomainEvaluation
	sqlStatement := `SELECT id, domain, EvaluationHour, EvaluationInProgress, sslGrade, logo, title, isDown FROM domainEvaluation;`
	rows, err := Query(dbc, sqlStatement)

	if err != nil {
		return domainEvaluations, err
	}

	for rows.Next() {
		var de DomainEvaluation
		if err = rows.Scan(&de.Id, &de.Domain, &de.EvaluationHour, &de.EvaluationInProgress,
			&de.SslGrade, &de.Logo, &de.Title, &de.IsDown); err != nil {
			return domainEvaluations, err
		}
		domainEvaluations = append(domainEvaluations, de)
	}

	if err = rows.Err(); err != nil {
		return domainEvaluations, err
	}

	return domainEvaluations, err
}

// Function for listing the last domain evaluations for each unique domain name
// in the database.
func ListRecentDomainEvaluations(dbc interface{}) ([]DomainEvaluation, error) {
	del, err := ListDomainEvaluations(dbc)
	if err != nil {
		return nil, err
	}

	domains := make(map[string][]DomainEvaluation)
	for i, _ := range del {
		if arr, ok := domains[del[i].Domain]; ok {
			arr = append(arr, del[i])
			domains[del[i].Domain] = arr
		} else {
			domains[del[i].Domain] = make([]DomainEvaluation, 0)
			arr = append(arr, del[i])
			domains[del[i].Domain] = arr
		}
	}

	recentDomainEvaluations := make([]DomainEvaluation, 0)
	for _, delByDomain := range domains {

		if len(delByDomain) >= 1 {
			lowestBound := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
			highest := lowestBound
			highestDE := delByDomain[0]

			for i := 1; i < len(delByDomain); i++ {
				v := delByDomain[i]
				var d time.Time
				d, err = time.Parse(time.RFC3339, v.EvaluationHour)
				if err != nil {
					return nil, err
				}
				if d.After(highest) {
					highest = d
					highestDE = v
				}
			}
			recentDomainEvaluations = append(recentDomainEvaluations, highestDE)
		}

	}
	return recentDomainEvaluations, err
}

// Function for listing the servers corresponding to a specific idDomainEvaluation
func ListServersID(idDomainEvaluation int, dbc interface{}) ([]Server, error) {
	var servers []Server
	sqlStatement := `SELECT id, address, sslGrade, country, owner FROM server
						WHERE domainEvaluationId = $1;`
	rows, err := Query(dbc, sqlStatement, idDomainEvaluation)

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

// Method for listing the servers of a DomainEvaluation
func (de *DomainEvaluation) ListServers(dbc interface{}) error {
	servers, err := ListServersID(de.Id, dbc)
	de.Servers = servers
	return err
}

// SearchLastEvaluation
// Method for searching the last evaluation done before a given time.

// Additional to the time, the method receives the domainName of the evaluation,
// and the status of the evaluation, allowing for search for either the last
// evaluation in progress, or the last evaluation ready

func (de *DomainEvaluation) SearchLastEvaluation(domainName string, EvaluationInProgress bool,
	upperBound time.Time, dbc interface{}) error {

	var domainEvaluations []DomainEvaluation
	sqlStatement := `SELECT id, domain, EvaluationHour, EvaluationInProgress, sslGrade, logo,
		title, isDown	FROM domainEvaluation WHERE domain = $1 AND EvaluationInProgress = $2;`
	rows, err := Query(dbc, sqlStatement, domainName, EvaluationInProgress)
	if err != nil {
		return err
	}

	for rows.Next() {
		var deTmp DomainEvaluation
		if err = rows.Scan(&deTmp.Id, &deTmp.Domain, &deTmp.EvaluationHour, &deTmp.EvaluationInProgress,
			&deTmp.SslGrade, &deTmp.Logo, &deTmp.Title, &deTmp.IsDown); err != nil {
			return err
		}
		domainEvaluations = append(domainEvaluations, deTmp)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	if len(domainEvaluations) == 0 {
		return err
	}

	lowestBound := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	highest := lowestBound
	highestID := 0

	for _, v := range domainEvaluations {
		var d time.Time
		d, err = time.Parse(time.RFC3339, v.EvaluationHour)
		if err != nil {
			return err
		}
		if d.After(highest) && d.Before(upperBound) {
			highest = d
			highestID = v.Id
		}
	}

	de.Id = highestID
	de.SelectInDB(dbc)
	return err
}


var SLStatus = newSLSRegistry()
func newSLSRegistry() *slsRegistry {
	return &slsRegistry{
		NoPastEvaluation: 0,
		Unchanged:        1,
		Changed:          2,
	}
}

type slsRegistry struct {
	NoPastEvaluation int
	Unchanged        int
	Changed          int
}

// HaveServersChanged
// Method for evaluating if the servers of a domain evaluation have changed.

// The method compares the servers of the current DomainEvaluation structure
// with the servers of the previous DomainEvaluation(one hour before)
// in the database.

func (de *DomainEvaluation) HaveServersChanged(dbc interface{}) (int, error) {
	EvaluationHour, err := time.Parse(time.RFC3339, de.EvaluationHour)
	if err != nil {
		return SLStatus.NoPastEvaluation, err
	}
	EvaluationHourS1H := EvaluationHour.Add(time.Hour * -1)

	deTmp := DomainEvaluation{}
	err = deTmp.SearchLastEvaluation(de.Domain, false, EvaluationHourS1H, dbc)
	if err != nil {
		return SLStatus.NoPastEvaluation, err
	}
	if deTmp.Id == 0 {
		return SLStatus.NoPastEvaluation, nil
	}
	err = deTmp.ListServers(dbc)
	if err != nil {
		return SLStatus.NoPastEvaluation, err
	}

	if !CompareServerList(deTmp.Servers, de.Servers) {
		return SLStatus.Changed, nil
	}

	return SLStatus.Unchanged, nil
}

// PreviousSSLgrade
// Method for evaluating the previous sslgrade of a given domain evaluation.

// The method compares the current DomainEvaluation structure with the previous
// one (one hour before) in the database.
func (de *DomainEvaluation) PreviousSSLgrade(dbc interface{}) (string, error) {
	EvaluationHour, err := time.Parse(time.RFC3339, de.EvaluationHour)
	if err != nil {
		return `NO EVALUATION`, err
	}
	EvaluationHourS1H := EvaluationHour.Add(time.Hour * -1)

	deTmp := DomainEvaluation{}
	err = deTmp.SearchLastEvaluation(de.Domain, false, EvaluationHourS1H, dbc)
	if err != nil {
		return `NO EVALUATION`, err
	}
	if deTmp.Id == 0 {
		return `NO EVALUATION`, nil
	}

	return deTmp.SslGrade, nil
}
