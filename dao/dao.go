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

// Postgresql settings
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

// ServerEvaluation: Struct for the representation of a SSLabs test in
// a specific domain.
type ServerEvaluation struct {
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

// ServerEvaluationComplete: Struct for the representation of all data
// about a specific domain using different scrapers and db queries.
// It's used for displaying information in the API
type ServerEvaluationComplete struct {
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
func (sec *ServerEvaluationComplete) Copy(se ServerEvaluation) {
	sec.Servers = se.Servers
	sec.SslGrade = se.SslGrade
	sec.IsDown = se.IsDown
	sec.Logo = se.Logo
	sec.Title = se.Title
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

// Compare two serverEvaluation structures.
func CompareServerEvaluation(se1, se2 ServerEvaluation) bool {
	return se1.Domain == se2.Domain && se1.EvaluationHour == se2.EvaluationHour &&
		se1.EvaluationInProgress == se2.EvaluationInProgress && se1.SslGrade == se2.SslGrade &&
		se1.IsDown == se2.IsDown && CompareServerList(se1.Servers, se2.Servers)
}

// DAO interface: Declaration of principal methods for data manipulation
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

func InitServerEvaluationTable(dbc *sql.DB) error {
	sqlStatement := `CREATE TABLE serverEvaluation (id SERIAL PRIMARY KEY,
					domain VARCHAR(100), EvaluationHour VARCHAR(30), EvaluationInProgress boolean,
					sslGrade VARCHAR(5), logo VARCHAR(80), title VARCHAR(80), isDown boolean);`
	_, err := dbc.Exec(sqlStatement)
	return err
}
func InitServerTable(dbc *sql.DB) error {
	sqlStatement := `CREATE TABLE server (id SERIAL PRIMARY KEY,
		serverEvaluationId integer,	address VARCHAR(50), sslGrade VARCHAR(5), country VARCHAR(20),
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
	sqlStatement1 := `DELETE FROM server WHERE id > 0;`
	_, err := dbc.Exec(sqlStatement1)
	if err != nil {
		return err
	}
	sqlStatement2 := `DELETE FROM serverevaluation WHERE id > 0;`
	_, err = dbc.Exec(sqlStatement2)
	return err
}

//
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

func (s *Server) CreateInDB(dbc interface{}) error {
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

func (s *Server) UpdateInDB(dbc interface{}) error {
	sqlStatement := `UPDATE server SET address = $2, sslGrade = $3, country = $4,
	owner = $5 WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, s.Id, s.Address,
		s.SslGrade, s.Country, s.Owner)
	return err
}

func (s *Server) DeleteInDB(dbc interface{}) error {
	sqlStatement := `DELETE FROM server WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, s.Id)
	return err
}

func (se *ServerEvaluation) SelectInDB(dbc interface{}) error {
	sqlStatement := `SELECT domain, EvaluationHour, EvaluationInProgress, sslGrade,
	logo, title, isDown FROM serverEvaluation WHERE id=$1;`
	row, err := QueryRow(dbc, sqlStatement, se.Id)
	err = row.Scan(&se.Domain, &se.EvaluationHour, &se.EvaluationInProgress, &se.SslGrade,
		&se.Logo, &se.Title, &se.IsDown)
	switch err {
	case sql.ErrNoRows:
		return errors.New("No rows were returned.")
	case nil:
		return nil
	default:
		return err
	}

}

func (se *ServerEvaluation) CreateInDB(dbc interface{}) error {
	sqlStatement := `INSERT INTO serverEvaluation (domain, EvaluationHour, EvaluationInProgress, sslGrade,
		logo, title, isDown) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;`
	row, err := QueryRow(dbc, sqlStatement, se.Domain, se.EvaluationHour,
		se.EvaluationInProgress, se.SslGrade, se.Logo, se.Title, se.IsDown)
	err = row.Scan(&se.Id)
	if err != nil {
		return err
	}
	if len(se.Servers) > 0 {
		for _, v := range se.Servers {
			if err = v.CreateInDB(dbc); err != nil {
				return err
			}
			if err = v.updateServerEvaluationInDB(se.Id, dbc); err != nil {
				return err
			}
		}
	}
	return err
}

func (se *ServerEvaluation) UpdateInDB(dbc interface{}) error {
	sqlStatement := `UPDATE serverEvaluation SET domain = $2, EvaluationHour = $3, EvaluationInProgress = $4,
	sslGrade = $5, logo = $6, title = $7, isDown = $8 WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, se.Id, se.Domain, se.EvaluationHour,
		se.EvaluationInProgress, se.SslGrade, se.Logo, se.Title, se.IsDown)
	if err != nil {
		return err
	}
	if len(se.Servers) > 0 {
		err = se.deleteAllServersInDB(dbc)
		if err != nil {
			return err
		}
		for _, v := range se.Servers {
			if err = v.CreateInDB(dbc); err != nil {
				return err
			}
			if err = v.updateServerEvaluationInDB(se.Id, dbc); err != nil {
				return err
			}
		}
	}
	return err
}

// logo = $6, title
func (se *ServerEvaluation) UpdateLogoInDb(dbc interface{}) error {
	sqlStatement := `UPDATE serverEvaluation SET logo = $2 WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, se.Id, se.Logo)
	return err
}
func (se *ServerEvaluation) UpdateTitleInDb(dbc interface{}) error {
	sqlStatement := `UPDATE serverEvaluation SET title = $2 WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, se.Id, se.Title)
	return err
}

func (se *ServerEvaluation) UpdateHourInDb(dbc interface{}) error {
	sqlStatement := `UPDATE serverEvaluation SET EvaluationHour = $2 WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, se.Id, se.EvaluationHour)
	return err
}

func (se *ServerEvaluation) deleteAllServersInDB(dbc interface{}) error {
	sqlStatement := `DELETE FROM server WHERE serverEvaluationId = $1;`
	_, err := Exec(dbc, sqlStatement, se.Id)
	return err
}

func (se *ServerEvaluation) DeleteInDB(dbc interface{}) error {
	sqlStatement := `DELETE FROM serverEvaluation WHERE id = $1;`
	_, err := Exec(dbc, sqlStatement, se.Id)
	if err == nil {
		return se.deleteAllServersInDB(dbc)
	}
	return err
}

func ListServerEvaluations(dbc interface{}) ([]ServerEvaluation, error) {
	var serverEvaluations []ServerEvaluation
	sqlStatement := `SELECT id, domain, EvaluationHour, EvaluationInProgress, sslGrade, logo, title, isDown FROM serverEvaluation;`
	rows, err := Query(dbc, sqlStatement)

	if err != nil {
		return serverEvaluations, err
	}

	for rows.Next() {
		var se ServerEvaluation
		if err = rows.Scan(&se.Id, &se.Domain, &se.EvaluationHour, &se.EvaluationInProgress,
			&se.SslGrade, &se.Logo, &se.Title, &se.IsDown); err != nil {
			return serverEvaluations, err
		}
		serverEvaluations = append(serverEvaluations, se)
	}

	if err = rows.Err(); err != nil {
		return serverEvaluations, err
	}

	return serverEvaluations, err
}

func ListRecentServerEvaluations(dbc interface{}) ([]ServerEvaluation, error) {
	sel, err := ListServerEvaluations(dbc)
	if err != nil {
		return nil, err
	}

	domains := make(map[string][]ServerEvaluation)
	for i, _ := range sel {
		if arr, ok := domains[sel[i].Domain]; ok {
			arr = append(arr, sel[i])
			domains[sel[i].Domain] = arr
		} else {
			domains[sel[i].Domain] = make([]ServerEvaluation, 0)
			arr = append(arr, sel[i])
			domains[sel[i].Domain] = arr
		}
	}

	recentEvaluations := make([]ServerEvaluation, 0)
	for _, selByDomain := range domains {

		if len(selByDomain) >= 1 {
			lowestBound := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
			highest := lowestBound
			highestSE := selByDomain[0]

			for i := 1; i < len(selByDomain); i++ {
				v := selByDomain[i]
				var d time.Time
				d, err = time.Parse(time.RFC3339, v.EvaluationHour)
				if err != nil {
					return nil, err
				}
				if d.After(highest) {
					highest = d
					highestSE = v
				}
			}
			recentEvaluations = append(recentEvaluations, highestSE)
		}

	}
	return recentEvaluations, err
}

func ListServersID(idServerEvaluation int, dbc interface{}) ([]Server, error) {
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

func (se *ServerEvaluation) listServers(dbc interface{}) error {
	servers, err := ListServersID(se.Id, dbc)
	se.Servers = servers
	return err
}

func (se *ServerEvaluation) SearchLastEvaluation(domainName string, EvaluationInProgress bool,
	upperBound time.Time, dbc interface{}) error {

	var serverEvaluations []ServerEvaluation
	sqlStatement := `SELECT id, domain, EvaluationHour, EvaluationInProgress, sslGrade, logo,
		title, isDown	FROM serverEvaluation WHERE domain = $1 AND EvaluationInProgress = $2;`
	rows, err := Query(dbc, sqlStatement, domainName, EvaluationInProgress)
	if err != nil {
		return err
	}

	for rows.Next() {
		var seTmp ServerEvaluation
		if err = rows.Scan(&seTmp.Id, &seTmp.Domain, &seTmp.EvaluationHour, &seTmp.EvaluationInProgress,
			&seTmp.SslGrade, &seTmp.Logo, &seTmp.Title, &seTmp.IsDown); err != nil {
			return err
		}
		serverEvaluations = append(serverEvaluations, seTmp)
	}

	if err = rows.Err(); err != nil {
		return err
	}

	if len(serverEvaluations) == 0 {
		return err
	}

	lowestBound := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	highest := lowestBound
	highestID := 0

	for _, v := range serverEvaluations {
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

	se.Id = highestID
	se.SelectInDB(dbc)
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

func (se *ServerEvaluation) HaveServersChanged(dbc interface{}) (int, error) {
	EvaluationHour, err := time.Parse(time.RFC3339, se.EvaluationHour)
	if err != nil {
		return SLStatus.NoPastEvaluation, err
	}
	EvaluationHourS1H := EvaluationHour.Add(time.Hour * -1)

	seTmp := ServerEvaluation{}
	err = seTmp.SearchLastEvaluation(se.Domain, false, EvaluationHourS1H, dbc)
	if err != nil {
		return SLStatus.NoPastEvaluation, err
	}
	if seTmp.Id == 0 {
		return SLStatus.NoPastEvaluation, nil
	}
	err = seTmp.listServers(dbc)
	if err != nil {
		return SLStatus.NoPastEvaluation, err
	}

	if !CompareServerList(seTmp.Servers, se.Servers) {
		return SLStatus.Changed, nil
	}

	return SLStatus.Unchanged, nil
}

func (se *ServerEvaluation) PreviousSSLgrade(dbc interface{}) (string, error) {
	EvaluationHour, err := time.Parse(time.RFC3339, se.EvaluationHour)
	if err != nil {
		return `NO EVALUATION`, err
	}
	EvaluationHourS1H := EvaluationHour.Add(time.Hour * -1)

	seTmp := ServerEvaluation{}
	err = seTmp.SearchLastEvaluation(se.Domain, false, EvaluationHourS1H, dbc)
	if err != nil {
		return `NO EVALUATION`, err
	}
	if seTmp.Id == 0 {
		return `NO EVALUATION`, nil
	}

	return seTmp.SslGrade, nil
}
