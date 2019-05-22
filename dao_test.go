package main

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestMakeEvaluationInDomain(t *testing.T) {

	// DB TEST CONFIGURATION
	db, err := initDB()

	if err != nil {
		t.Error(fmt.Sprintf("Exception: %v", err))
	}
	CleanDataInDB(db)

	domainName := `prueba1.com`
	currentHour1, _ := time.Parse(time.RFC3339, `2016-01-01T15:00:00+02:00`)
	makeEvalCase1 := func(s string) (ServerEvaluation, error) {
		return ServerEvaluation{1, s, `2016-01-01T15:00:00+02:00`, true, make([]Server, 0), ``, false}, nil
	}
	expected1 := ServerEvaluation{1, domainName, `2016-01-01T15:00:00+02:00`, true, make([]Server, 0), ``, false}
	t.Run("CASE 1: NO PENDING EVALUATION AND NO PAST EVALUATION HOUR ",
		testMakeEvaluationInDomainFunc(domainName, currentHour1, makeEvalCase1, db, expected1))

	domainName = `prueba1.com`
	currentHour2, _ := time.Parse(time.RFC3339, `2016-01-01T15:00:15+02:00`)
	makeEvalCase2 := func(s string) (ServerEvaluation, error) {
		return ServerEvaluation{2, s, `2016-01-01T15:00:15+02:00`, true, make([]Server, 0), ``, false}, nil
	}
	expected2 := expected1
	t.Run("CASE 2: PENDING EVALUATION, CURRENT EVALUATION HOUR < PENDING EVALUATION HOUR + 20S",
		testMakeEvaluationInDomainFunc(domainName, currentHour2, makeEvalCase2, db, expected2))

	domainName = `prueba1.com`
	currentHour3, _ := time.Parse(time.RFC3339, `2016-01-01T15:00:25+02:00`)
	makeEvalCase3 := func(s string) (ServerEvaluation, error) {
		return ServerEvaluation{3, s, `2016-01-01T15:00:25+02:00`, true, make([]Server, 0), ``, false}, nil
	}
	expected3 := ServerEvaluation{3, domainName, `2016-01-01T15:00:25+02:00`, true, make([]Server, 0), ``, false}
	t.Run("CASE 3: PENDING EVALUATION, CURRENT HOUR > PENDING EVALUATION HOUR + 20 | CURRENT EVALUATION IN PROGRESS",
		testMakeEvaluationInDomainFunc(domainName, currentHour3, makeEvalCase3, db, expected3))

	domainName = `prueba1.com`
	currentHour4, _ := time.Parse(time.RFC3339, `2016-01-01T15:00:48+02:00`)
	makeEvalCase4 := func(s string) (ServerEvaluation, error) {
		return ServerEvaluation{4, s, `2016-01-01T15:00:48+02:00`, false, make([]Server, 0), `A+`, false}, nil
	}
	expected4 := ServerEvaluation{4, domainName, `2016-01-01T15:00:48+02:00`, false, make([]Server, 0), `A+`, false}
	t.Run("CASE 4: PENDING EVALUATION, CURRENT HOUR > PENDING EVALUATION HOUR + 20 | !CURRENT EVALUATION IN PROGRESS ",
		testMakeEvaluationInDomainFunc(domainName, currentHour4, makeEvalCase4, db, expected4))

	domainName = `prueba1.com`
	currentHour5, _ := time.Parse(time.RFC3339, `2016-01-01T15:01:01+02:00`)
	makeEvalCase5 := func(s string) (ServerEvaluation, error) {
		return ServerEvaluation{5, s, `2016-01-01T15:01:01+02:00`, false, make([]Server, 0), `B+`, false}, nil
	}
	expected5 := expected4
	t.Run("CASE 5: PAST EVALUATION, CURRENT HOUR < PAST EVALUATION HOUR + 20",
		testMakeEvaluationInDomainFunc(domainName, currentHour5, makeEvalCase5, db, expected5))

	domainName = `prueba1.com`
	currentHour6, _ := time.Parse(time.RFC3339, `2016-01-01T15:01:18+02:00`)
	makeEvalCase6 := func(s string) (ServerEvaluation, error) {
		return ServerEvaluation{6, s, `2016-01-01T15:01:18+02:00`, false, make([]Server, 0), `B+`, false}, nil
	}
	expected6 := ServerEvaluation{6, domainName, `2016-01-01T15:01:18+02:00`, false, make([]Server, 0), `B+`, false}
	t.Run("CASE 6: PAST EVALUATION, CURRENT HOUR > PAST EVALUATION HOUR + 20",
		testMakeEvaluationInDomainFunc(domainName, currentHour6, makeEvalCase6, db, expected6))

	CleanDataInDB(db)
	db.Close()
}

func testMakeEvaluationInDomainFunc(domainName string, currentHour time.Time, makeTest func(string) (ServerEvaluation, error),
	db *sql.DB, expected ServerEvaluation) func(*testing.T) {
	return func(t *testing.T) {
		actual, err := MakeEvaluationInDomain(domainName, currentHour, makeTest, db)
		if err != nil {
			t.Error(fmt.Sprintf("Exception: %v", err))
		}
		opt := cmp.Comparer(func(x, y ServerEvaluation) bool {
			return CompareServerEvaluation(x, y)
		})
		if !cmp.Equal(actual, expected, opt) {
			t.Error(fmt.Sprintf("Expected: %v, Actual: %v", expected, actual))
		}
	}
}
