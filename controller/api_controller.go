package controller

import (
  "time"
  "github.com/trotsdeveloper/truora_test/truora_test_golang/scrapers"
  "github.com/trotsdeveloper/truora_test/truora_test_golang/dao"
  "github.com/cockroachdb/cockroach-go/crdb"
  "database/sql"
  "context"
)

var APIErrors = newAPIErrorsRegistry()

func newAPIErrorsRegistry() *apiErrorsRegistry {
	E601v := makeAPIError("601", "Error in database.")
	E602v := makeAPIError("602", "Error in SSLabs API.")
	E701v := makeAPIError("701", "Error getting Icon")
	E702v := makeAPIError("702", "Error getting HTML Title")
	E801v := makeAPIError("801", "Error getting country from WHOIS")
	E802v := makeAPIError("802", "Error getting owner from WHOIS")

	return &apiErrorsRegistry{
		E601: E601v,
		E602: E602v,
		E701: E701v,
		E702: E702v,
		E801: E801v,
		E802: E802v,
	}
}

func makeAPIError(code string, description string) func(error) (APIError) {
	return func(err error) (APIError) {
		return APIError{code, description, err.Error()}
	}
}

type apiErrorsRegistry struct {
	E601 func(error) (APIError) //
	E602 func(error) (APIError) //
	E701 func(error) (APIError) //
	E702 func(error) (APIError) //
	E801 func(error) (APIError) //
	E802 func(error) (APIError) //
}

type APIError struct {
	Code string `json:"code"`
	Description string `json:"description"`
	Err string `json:"error_message"`
}


func DefaultAPIError() APIError {
  return APIError{Code: "600"}
}

func (x *APIError) IsInArray(a []APIError) bool {
  for _, n := range a {
    if x.Code == n.Code {
      return true
    }
  }
  return false
}

func MakeEvaluationInDomain(domainName string, currentHour time.Time, evaluator func(time.Time, string) (dao.ServerEvaluation, error),
	db *sql.DB) (se dao.ServerEvaluation, appErr APIError) {
	// 1) In the database, is there a server Evaluation in process
	// with the same given domain?
  appErr = DefaultAPIError()
  var err error
	pendingEvaluation := dao.ServerEvaluation{}

	err = pendingEvaluation.SearchLastEvaluation(domainName, true, currentHour, db)
	if err != nil {
    appErr = APIErrors.E601(err)
		return
	}
	if pendingEvaluation.Id != 0 {
		// 1.1) YES: Is difference between the current hour
		// and the pending Evaluation lower than 20 seconds?
		var pendingEvaluationHour time.Time
		pendingEvaluationHour, err = time.Parse(time.RFC3339, pendingEvaluation.EvaluationHour)
		if err != nil {
      appErr = APIErrors.E601(err)
			return
		}
		pendingEvaluationHourA20 := pendingEvaluationHour.Add(time.Second * 20)
		if pendingEvaluationHourA20.After(currentHour) {
			// 1.1.1) YES: In the database, data will remain unchanged
			// return the pending Evaluation
			se = pendingEvaluation
			return
		} else {
			// 1.1.2) Update the hour of the pending Evaluation with the current hour
			pendingEvaluation.EvaluationHour = currentHour.Format(time.RFC3339)
			err = pendingEvaluation.UpdateHourInDb(db)
			if err != nil {
        appErr = APIErrors.E601(err)
				return
			}

			var currentEvaluation dao.ServerEvaluation
			currentEvaluation, err = evaluator(currentHour, domainName)
			if err != nil {
        appErr = APIErrors.E602(err)
				return
			}
			// 1.1.2) NO: In SSLabs, Is the server Evaluation in process?
			if currentEvaluation.EvaluationInProgress {
				// 1.1.2.1) YES: Return the pending Evaluation with the new hour
				se = pendingEvaluation
				return
			} else {
				// 1.1.2.2) NO: Update the pending Evaluation in the database, with
				// the information of the current Evaluation
				currentEvaluation.Id = pendingEvaluation.Id
        err = crdb.ExecuteTx(context.Background(), db, nil, func(tx *sql.Tx) error {
          return currentEvaluation.UpdateInDB(tx)
        })
        if err != nil {
          appErr = APIErrors.E601(err)
          return
        }
				se = currentEvaluation
				return
			}
		}
	} else {
		// 1.2) NO: Is there a past server Evaluation, ready, with the same given domain?
		pastEvaluation := dao.ServerEvaluation{}
		err = pastEvaluation.SearchLastEvaluation(domainName, false, currentHour, db)
    if err != nil {
      appErr = APIErrors.E601(err)
      return
    }
		if pastEvaluation.Id != 0 {
			// 1.2.1) YES: Is difference lower than 20 seconds?
			var pastEvaluationHour time.Time
			pastEvaluationHour, err = time.Parse(time.RFC3339, pastEvaluation.EvaluationHour)
			if err != nil {
        appErr = APIErrors.E601(err)
				return
			}
			pastEvaluationHourA20 := pastEvaluationHour.Add(time.Second * 20)
			if pastEvaluationHourA20.After(currentHour) {
				// 1.2.1.1) YES: In the database, data will remain unchanged,
				// return the past Evaluation
				se = pastEvaluation
				return
			} else {
				// 1.2.1.2) NO: Make a server Evaluation from SSLabs, save it in DB.
				var currentEvaluation dao.ServerEvaluation
				currentEvaluation, err = evaluator(currentHour, domainName)
				if err != nil {
          appErr = APIErrors.E602(err)
					return
				}
        err = crdb.ExecuteTx(context.Background(), db, nil, func(tx *sql.Tx) error {
          return currentEvaluation.CreateInDB(tx)
        })
        if err != nil {
          appErr =  APIErrors.E601(err)
          return
        }
				se = currentEvaluation
				return
			}
		} else {
			// 1.2.2) NO: Make a server Evaluation from SSLabs, save it in DB.

			var currentEvaluation dao.ServerEvaluation
			currentEvaluation, err = evaluator(currentHour, domainName)
			if err != nil {
        appErr = APIErrors.E602(err)
				return
			}
      err = crdb.ExecuteTx(context.Background(), db, nil, func(tx *sql.Tx) error {
        return currentEvaluation.CreateInDB(tx)
      })
      if err != nil {
        appErr = APIErrors.E601(err)
        return
      }
			se = currentEvaluation
			return
		}
	}
}

func ScraperTestComplete(domain string, currentHour time.Time, db *sql.DB) (sec dao.ServerEvaluationComplete, appErrs []APIError) {

	sec = dao.ServerEvaluationComplete{}
	appErrs = make([]APIError, 0)

	se, appErr := MakeEvaluationInDomain(domain, currentHour, scrapers.ScraperSSLabs, db)
  defaultCode := DefaultAPIError()
	if !(appErr.Code == defaultCode.Code) {
		appErrs = append(appErrs, appErr)
	}
  var err error
	sec.Copy(se)

	if !se.IsDown {
		sec.Logo, err = scrapers.ScraperLogo(domain)
		if err != nil {
			appErrs = append(appErrs, APIErrors.E701(err))
		}
		sec.Title, err = scrapers.ScraperTitle(domain)
		if err != nil {
			appErrs = append(appErrs, APIErrors.E702(err))
		}
	}

	if !se.EvaluationInProgress && !se.IsDown {
		for i := range sec.Servers {
			ip := sec.Servers[i].Address
			sec.Servers[i].Country, err = scrapers.ScraperCountry(ip)
			if err != nil {
				appErrs = append(appErrs, APIErrors.E801(err))
			}
			sec.Servers[i].Owner, err = scrapers.ScraperOwner(ip)
			if err != nil {
				appErrs = append(appErrs, APIErrors.E802(err))
			}
		}
		var serversChangedI int
		serversChangedI, err = se.HaveServersChanged(db)
		if err != nil {
			appErrs = append(appErrs, APIErrors.E601(err))
			return
		}
		sec.ServersChanged = (serversChangedI == dao.SLStatus.Changed)
		sec.PreviousSslGrade, err = se.PreviousSSLgrade(db)
		if err != nil {
			appErrs = append(appErrs, APIErrors.E601(err))
      return
		}
	}

	return
}
