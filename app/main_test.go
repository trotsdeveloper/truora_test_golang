package main

import (
	"database/sql"
	"fmt"
	"testing"
	"time"
	"github.com/google/go-cmp/cmp"
  "github.com/trotsdeveloper/truora_test/truora_test_golang/dao"
  "github.com/trotsdeveloper/truora_test/truora_test_golang/controller"
)

func TestMakeEvaluationInDomain(t *testing.T) {

	// DB TEST CONFIGURATION
	db, err := dao.InitDB()
	if err != nil {
		t.Error(fmt.Sprintf("Exception: %v", err))
	}
	dao.DropServerTable(db)
	dao.DropServerEvaluationTable(db)
	dao.InitServerEvaluationTable(db)
	dao.InitServerTable(db)
	dao.CleanDataInDB(db)

	domainName := `prueba1.com`
	currentHour1, _ := time.Parse(time.RFC3339, `2016-01-01T15:00:00+02:00`)
	expected1 := dao.ServerEvaluation{1, domainName, `2016-01-01T15:00:00+02:00`, true, make([]dao.Server, 0), ``, ``, ``, false}
	makeEvalCase1 := func(t time.Time, s string) (dao.ServerEvaluation, error) {
		return expected1, nil
	}
	t.Run("CASE 1: NO PENDING EVALUATION AND NO PAST EVALUATION HOUR ",
		testMakeEvaluationInDomainFunc(domainName, currentHour1, makeEvalCase1, db, expected1))

	domainName = `prueba1.com`
	currentHour2, _ := time.Parse(time.RFC3339, `2016-01-01T15:00:15+02:00`)
	expected2 := expected1
	makeEvalCase2 := func(t time.Time, s string) (dao.ServerEvaluation, error) {
		return dao.ServerEvaluation{2, domainName, `2016-01-01T15:00:15+02:00`, true, make([]dao.Server, 0), ``, ``, ``, false}, nil
	}

	t.Run("CASE 2: PENDING EVALUATION, CURRENT EVALUATION HOUR < PENDING EVALUATION HOUR + 20S",
		testMakeEvaluationInDomainFunc(domainName, currentHour2, makeEvalCase2, db, expected2))

	domainName = `prueba1.com`
	currentHour3, _ := time.Parse(time.RFC3339, `2016-01-01T15:00:25+02:00`)
	expected3 := dao.ServerEvaluation{3, domainName, `2016-01-01T15:00:25+02:00`, true, make([]dao.Server, 0), ``, ``, ``, false}
	makeEvalCase3 := func(t time.Time, s string) (dao.ServerEvaluation, error) {
		return expected3, nil
	}

	t.Run("CASE 3: PENDING EVALUATION, CURRENT HOUR > PENDING EVALUATION HOUR + 20 | CURRENT EVALUATION IN PROGRESS",
		testMakeEvaluationInDomainFunc(domainName, currentHour3, makeEvalCase3, db, expected3))

	domainName = `prueba1.com`
	currentHour4, _ := time.Parse(time.RFC3339, `2016-01-01T15:00:48+02:00`)
	expected4 := dao.ServerEvaluation{4, domainName, `2016-01-01T15:00:48+02:00`, false, make([]dao.Server, 0), `A+`, ``, ``,false}
	makeEvalCase4 := func(t time.Time, s string) (dao.ServerEvaluation, error) {
		return expected4, nil
	}
	t.Run("CASE 4: PENDING EVALUATION, CURRENT HOUR > PENDING EVALUATION HOUR + 20 | !CURRENT EVALUATION IN PROGRESS ",
		testMakeEvaluationInDomainFunc(domainName, currentHour4, makeEvalCase4, db, expected4))

	domainName = `prueba1.com`
	currentHour5, _ := time.Parse(time.RFC3339, `2016-01-01T15:01:01+02:00`)
	expected5 := expected4
	makeEvalCase5 := func(t time.Time, s string) (dao.ServerEvaluation, error) {
		return dao.ServerEvaluation{5, domainName, `2016-01-01T15:01:01+02:00`, false, make([]dao.Server, 0), `B+`, ``, ``,false}, nil
	}
	t.Run("CASE 5: PAST EVALUATION, CURRENT HOUR < PAST EVALUATION HOUR + 20",
		testMakeEvaluationInDomainFunc(domainName, currentHour5, makeEvalCase5, db, expected5))

	domainName = `prueba1.com`
	currentHour6, _ := time.Parse(time.RFC3339, `2016-01-01T15:01:18+02:00`)
	expected6 := dao.ServerEvaluation{6, domainName, `2016-01-01T15:01:18+02:00`, false, make([]dao.Server, 0), `B+`, ``, ``,false}
	makeEvalCase6 := func(t time.Time, s string) (dao.ServerEvaluation, error) {
		return dao.ServerEvaluation{6, domainName, `2016-01-01T15:01:18+02:00`, false, make([]dao.Server, 0), `B+`, ``, ``,false}, nil
	}
	t.Run("CASE 6: PAST EVALUATION, CURRENT HOUR > PAST EVALUATION HOUR + 20",
		testMakeEvaluationInDomainFunc(domainName, currentHour6, makeEvalCase6, db, expected6))

	dao.CleanDataInDB(db)
	db.Close()
}

func testMakeEvaluationInDomainFunc(domainName string, currentHour time.Time, evaluator func(time.Time, string) (dao.ServerEvaluation, error),
	db *sql.DB, expected dao.ServerEvaluation) func(*testing.T) {
	return func(t *testing.T) {
		actual, _, apiErr := controller.MakeEvaluationInDomain(domainName, currentHour, evaluator, db)
		if apiErr != controller.DefaultAPIError() {
			t.Error(fmt.Sprintf("Exception: %v", apiErr))
		}
		opt := cmp.Comparer(func(x, y dao.ServerEvaluation) bool {
			return dao.CompareServerEvaluation(x, y)
		})
		if !cmp.Equal(actual, expected, opt) {
			t.Error(fmt.Sprintf("Expected1: %v, Actual: %v", expected, actual))
		}
	}
}

// FUNCTION BLOCK
// HaveServersChanged, PreviousSSLgrade
func TestDBFunctions(t *testing.T) {
	// DB TEST CONFIGURATION
	db, err := dao.InitDB()
	if err != nil {
		t.Error(fmt.Sprintf("Exception: %v", err))
	}
	dao.CleanDataInDB(db)

	domainName := `prueba1.com`
	currentHour1, _ := time.Parse(time.RFC3339, `2016-01-01T15:00:00+02:00`)

	makeEvalCase1 := func(t time.Time, s string) (dao.ServerEvaluation, error) {
		servers1 := []dao.Server{dao.Server{Address: `128.30.20.10`}, dao.Server{Address: `128.28.20.10`}}
		return dao.ServerEvaluation{1, s, `2016-01-01T15:00:00+02:00`, false, servers1, `A+`, ``, ``, false}, nil
	}
	se1, _, _:= controller.MakeEvaluationInDomain(domainName, currentHour1, makeEvalCase1, db)
	t.Run("ServersChanged | CASE 1: NO PAST SERVER EVALUATIONS IN DATABASE",
		testHaveServersChangedFunc(se1, db, dao.SLStatus.NoPastEvaluation))
	t.Run("PreviousSSlGrade | CASE 1: NO PAST SERVER EVALUATIONS IN DATABASE",
		testPreviousSSLGradeFunc(se1, db, "NO EVALUATION"))

	domainName = `prueba1.com`
	currentHour2, _ := time.Parse(time.RFC3339, `2016-01-01T15:30:00+02:00`)
	makeEvalCase2 := func(t time.Time, s string) (dao.ServerEvaluation, error) {
		servers2 := []dao.Server{dao.Server{Address: `128.30.20.10`}, dao.Server{Address: `128.28.20.10`}}
		return dao.ServerEvaluation{1, s, `2016-01-01T15:30:00+02:00`, false, servers2, `A+`, ``, ``, false}, nil
	}
	se2, _, _:= controller.MakeEvaluationInDomain(domainName, currentHour2, makeEvalCase2, db)
	t.Run("ServersChanged | CASE 2: NO PAST SERVER EVALUATIONS ONE HOUR BEFORE",
		testHaveServersChangedFunc(se2, db, dao.SLStatus.NoPastEvaluation))
	t.Run("PreviousSSlGrade | CASE 2: NO PAST SERVER EVALUATIONS ONE HOUR BEFORE",
		testPreviousSSLGradeFunc(se2, db, "NO EVALUATION"))

	domainName = `prueba1.com`
	currentHour3, _ := time.Parse(time.RFC3339, `2016-01-01T16:20:00+02:00`)
	makeEvalCase3 := func(t time.Time, s string) (dao.ServerEvaluation, error) {
		servers3 := []dao.Server{dao.Server{Address: `128.30.20.10`}, dao.Server{Address: `128.28.20.10`}}
		return dao.ServerEvaluation{1, s, `2016-01-01T16:20:00+02:00`, false, servers3, `A+`, ``, ``, false}, nil
	}

	se3, _, _ := controller.MakeEvaluationInDomain(domainName, currentHour3, makeEvalCase3, db)
	t.Run("ServersChanged | CASE 3: PAST SERVER EVALUATION IN DATABASE | SERVER LIST UNCHANGED",
		testHaveServersChangedFunc(se3, db, dao.SLStatus.Unchanged))
	t.Run("PreviousSSlGrade | CASE 3: PAST SERVER EVALUATION IN DATABASE | PREVIOUS SSL GRADE UNCHANGED",
		testPreviousSSLGradeFunc(se3, db, "A+"))

	domainName = `prueba1.com`
	currentHour4, _ := time.Parse(time.RFC3339, `2016-01-01T16:25:00+02:00`)
	makeEvalCase4 := func(t time.Time, s string) (dao.ServerEvaluation, error) {
		servers4 := []dao.Server{dao.Server{Address: `128.30.28.10`}, dao.Server{Address: `128.28.20.10`}}
		return dao.ServerEvaluation{1, s, `2016-01-01T16:25:00+02:00`, false, servers4, `B+`, ``, ``, false}, nil
	}
	se4, _, _ := controller.MakeEvaluationInDomain(domainName, currentHour4, makeEvalCase4, db)
	t.Run("ServersChanged | CASE 4: PAST SERVER EVALUATION IN DATABASE | SERVER LIST CHANGED",
		testHaveServersChangedFunc(se4, db, dao.SLStatus.Changed))
	t.Run("PreviousSSlGrade | CASE 4: PAST SERVER EVALUATION IN DATABASE | PREVIOUS SSL GRADE CHANGED",
		testPreviousSSLGradeFunc(se4, db, "A+"))

	dao.CleanDataInDB(db)
	db.Close()
}

func testHaveServersChangedFunc(se dao.ServerEvaluation, db *sql.DB, expected int) func(*testing.T) {
	return func(t *testing.T) {
		actual, err := se.HaveServersChanged(db)
		if err != nil {
			t.Error(fmt.Sprintf("Exception: %v", err))
		}
		if !cmp.Equal(actual, expected) {
			t.Error(fmt.Sprintf("Expected: %v, Actual: %v", expected, actual))
		}
	}
}

func testPreviousSSLGradeFunc(se dao.ServerEvaluation, db *sql.DB, expected string) func(*testing.T) {
	return func(t *testing.T) {
		actual, err := se.PreviousSSLgrade(db)
		if err != nil {
			t.Error(fmt.Sprintf("Exception: %v", err))
		}
		if !cmp.Equal(actual, expected) {
			t.Error(fmt.Sprintf("Expected: %v, Actual: %v", expected, actual))
		}
	}
}
