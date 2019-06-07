// Package for the declaration of the api controller.
// The package contains the functions and the global variables
// neccesary for handling the behaviour of the server during a http request.
package controller

import (
  "time"
  "github.com/trotsdeveloper/truora_test/truora_test_golang/scrapers"
  "github.com/trotsdeveloper/truora_test/truora_test_golang/dao"
  "github.com/cockroachdb/cockroach-go/crdb"
  "database/sql"
  "context"
)

// Var representing the time to wait between different domain evaluations.
var DomainEvaluationTW time.Duration = time.Second * 20

// Var representing the time to wait for getting the list of recent evaluations.
var RecentEvaluationsTW time.Duration = time.Second * 20
// Var representing the last hour in which the list of recent evaluations was requested.
var RecentEvaluationsLQ time.Time

// Var for storing the recent evaluations each time the list of recent evaluations
// is requested. The var is useful for returning the last consult when the difference
// between the current hour and the RecentEvaluationsLQ is > RecentEvaluationsTW.
var RecentEvaluations = make([]dao.DomainEvaluation, 0)

// Var simulating a "enum" in other languages. The var handle the different APIErrors
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

// Auxiliar function for the APIErrors var.
func makeAPIError(code string, description string) func(error) (APIError) {
	return func(err error) (APIError) {
		return APIError{code, description, err.Error()}
	}
}

// Main struct for the APIErrors var
type apiErrorsRegistry struct {
	E601 func(error) (APIError) //
	E602 func(error) (APIError) //
	E701 func(error) (APIError) //
	E702 func(error) (APIError) //
	E801 func(error) (APIError) //
	E802 func(error) (APIError) //
}

// APIError - Struct for handling different API Errors.
// Used for showing application errors into a json instead of handling it with
// panics or terminal impressions.
type APIError struct {
	Code string `json:"code"`
	Description string `json:"description"`
	Err string `json:"error_message"`
}

// Default constructor for the APIError struct.
func DefaultAPIError() APIError {
  return APIError{Code: "600"}
}

// Method for detecting if the given APIError is in the given array of APIError.
// The method only compares the code.
func (x *APIError) IsInArray(a []APIError) bool {
  for _, n := range a {
    if x.Code == n.Code {
      return true
    }
  }
  return false
}

// One of the main functions for the api controller.
// The current function receives:
// -waitTime, indicating the time to wait
//- domainName, name of the domain to evaluate
//- currentHour, current hour of the evaluation
//- evaluator, function which receives a time representing the current hour
// a string, representing the domain to evaluate, and returns a domainevaluation.
// normally, evaluator is a scraper
//- db, a pointer representing the database controller to use for db queries.

// and returns:
// -de, domain evaluation representing the result of the function.
// -changed, boolean var representing if the domain evaluation have changed in the database.
// it's useful for knowing if it's neccesary  to scrape the web searching for the information
// about the domain with the given domainName.
// - appErr, if there aren't errors, appErr is the result of calling DefaultAPIError,
// if there are errors, appErr is going to be one of the possible values APIErrors

// The function implements the next algorithm:
// 1) In the database, is there a Domain Evaluation in process
// with the same given domain?
// 1.1) YES: Is difference between the current hour
// and the pending evaluation hour lower than waitTime.
// 1.1.1) YES: In the database, data will remain unchanged
// return the pending evaluation
// 1.1.2) NO: Update the hour of the pending Evaluation with the current hour
// In SSLabs, Is the Domain Evaluation in process?
// 1.1.2.1) YES: Return the pending evaluation with the new hour
// 1.1.2.2) NO: Update the pending evaluation in the database, with
// the information of the current evaluation. Changed var is now true.
// 1.2) NO: Is there a past Domain Evaluation, ready, with the same given domain?
// 1.2.1) YES: Is difference between the current hour
// and the pending evaluation is lower than waitTime.
// 1.2.1.1) YES: In the database, data will remain unchanged,
// return the past evaluation
// 1.2.1.2) NO: Make a Domain Evaluation using the SSLabs API, save it in DB.
// Changed is now true.
// 1.2.2) NO: Make a Domain Evaluation using the SSLabs API, save it in DB.
// Changed is now true.


func EvaluateDomainTW(waitTime time.Duration, domainName string, currentHour time.Time,
  evaluator func(time.Time, string) (dao.DomainEvaluation, error), db *sql.DB) (de dao.DomainEvaluation, changed bool, appErr APIError) {

  de.Servers = make([]dao.Server, 0)
  // 1) In the database, is there a Domain Evaluation in process
	// with the same given domain?
  changed = false
  appErr = DefaultAPIError()

  var err error
	pendingEvaluation := dao.DomainEvaluation{}

	err = pendingEvaluation.SearchLastEvaluation(domainName, true, currentHour, db)
	if err != nil {
    appErr = APIErrors.E601(err)
		return
	}

	if pendingEvaluation.Id != 0 {
		// 1.1) YES: Is difference between the current hour
		// and the pending evaluation is lower than waitTime.
		var pendingEvaluationHour time.Time
		pendingEvaluationHour, err = time.Parse(time.RFC3339, pendingEvaluation.EvaluationHour)
		if err != nil {
      appErr = APIErrors.E601(err)
			return
		}
		pendingEvaluationHourA20 := pendingEvaluationHour.Add(waitTime)
		if pendingEvaluationHourA20.After(currentHour) {
			// 1.1.1) YES: In the database, data will remain unchanged
			// return the pending evaluation
			de = pendingEvaluation
			return
		} else {
			// 1.1.2) NO: Update the hour of the pending Evaluation with the current hour
			pendingEvaluation.EvaluationHour = currentHour.Format(time.RFC3339)
			err = pendingEvaluation.UpdateHourInDb(db)
			if err != nil {
        appErr = APIErrors.E601(err)
				return
			}

			var currentEvaluation dao.DomainEvaluation
			currentEvaluation, err = evaluator(currentHour, domainName)
			if err != nil {
        appErr = APIErrors.E602(err)
				return
			}
			// 1.1.2) In SSLabs, Is the Domain Evaluation in process?
			if currentEvaluation.EvaluationInProgress {
				// 1.1.2.1) YES: Return the pending evaluation with the new hour
				de = pendingEvaluation
				return
			} else {
				// 1.1.2.2) NO: Update the pending evaluation in the database, with
				// the information of the current evaluation. Changed var is now true.
				currentEvaluation.Id = pendingEvaluation.Id
        err = crdb.ExecuteTx(context.Background(), db, nil, func(tx *sql.Tx) error {
          return currentEvaluation.UpdateInDB(tx)
        })
        if err != nil {
          appErr = APIErrors.E601(err)
          return
        }
				de = currentEvaluation
        changed = true
				return
			}
		}
	} else {
		// 1.2) NO: Is there a past Domain Evaluation, ready, with the same given domain?
		pastEvaluation := dao.DomainEvaluation{}
		err = pastEvaluation.SearchLastEvaluation(domainName, false, currentHour, db)
    if err != nil {
      appErr = APIErrors.E601(err)
      return
    }
		if pastEvaluation.Id != 0 {
			// 1.2.1) YES: Is difference between the current hour
  		// and the pending evaluation is lower than waitTime.
			var pastEvaluationHour time.Time
			pastEvaluationHour, err = time.Parse(time.RFC3339, pastEvaluation.EvaluationHour)
			if err != nil {
        appErr = APIErrors.E601(err)
				return
			}
			pastEvaluationHourA20 := pastEvaluationHour.Add(waitTime)
			if pastEvaluationHourA20.After(currentHour) {
				// 1.2.1.1) YES: In the database, data will remain unchanged,
				// return the past evaluation
				de = pastEvaluation
				return
			} else {
				// 1.2.1.2) NO: Make a Domain Evaluation using the SSLabs API, save it in DB.
        // Changed is now true.
				var currentEvaluation dao.DomainEvaluation
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
				de = currentEvaluation
        changed = true
				return
			}
		} else {
			// 1.2.2) NO: Make a Domain Evaluation using the SSLabs API, save it in DB.
      // Changed is now true.
			var currentEvaluation dao.DomainEvaluation
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
			de = currentEvaluation
      changed = true
			return
		}
	}
}

// Main function for evaluating domains, the function use the EvaluateDomainTW,
// passing it the global var DomainEvaluationTW containing the waiting time between
// evaluation of domains.
func EvaluateDomain(domainName string, currentHour time.Time, evaluator func(time.Time, string) (dao.DomainEvaluation, error),
	db *sql.DB) (de dao.DomainEvaluation, changed bool, appErr APIError) {
    return EvaluateDomainTW(DomainEvaluationTW, domainName, currentHour, evaluator, db)
}

// Main function for evaluating domains and scrapping the info about domains.
// The ScraperTestComplete function receives the domainName, the currentHour, and the
// global database controller db.

// and returns
// dec, a structure representing DomainEvaluationComplete
// appErrs, multiple errs scraping the web or getting info from the database.

// This function is used for displaying all the information about a specific domain
// including the information from the different scrapers. And servers_changed,
// previous_ssl_grade. Internally, the function uses the EvaluateDomain function for
// getting a specific DomainEvaluation structure, after that, if the EvaluateDomain function
// returns true, then ScraperTestComplete updates all info about the domain using scrapers.
func ScraperTestComplete(domain string, currentHour time.Time, db *sql.DB) (dec dao.DomainEvaluationComplete, appErrs []APIError) {
	dec = dao.DomainEvaluationComplete{}
  dec.Servers = make([]dao.Server, 0)

	appErrs = make([]APIError, 0)

	de, changed, appErr := EvaluateDomain(domain, currentHour, scrapers.ScraperSSLabs, db)
  defaultCode := DefaultAPIError()
	if !(appErr.Code == defaultCode.Code) {
		appErrs = append(appErrs, appErr)
	}
  var err error
	dec.Copy(de)

  if changed {
    if !de.IsDown {
      dec.Logo, err = scrapers.ScraperLogo(domain)
      if err != nil {
        appErrs = append(appErrs, APIErrors.E701(err))
      }
      de.Logo = dec.Logo
      err = de.UpdateLogoInDb(db)
      if err != nil {
        appErrs = append(appErrs, APIErrors.E601(err))
      }

      dec.Title, err = scrapers.ScraperTitle(domain)
      if err != nil {
        appErrs = append(appErrs, APIErrors.E702(err))
      }
      de.Logo = dec.Title
      err = de.UpdateTitleInDb(db)
      if err != nil {
        appErrs = append(appErrs, APIErrors.E601(err))
      }
    }
    if !de.EvaluationInProgress && !de.IsDown {
      for i := range dec.Servers {
        ip := dec.Servers[i].Address
        dec.Servers[i].Country, err = scrapers.ScraperCountry(ip)
        if err != nil {
          appErrs = append(appErrs, APIErrors.E801(err))
        }
        dec.Servers[i].Owner, err = scrapers.ScraperOwner(ip)
        if err != nil {
          appErrs = append(appErrs, APIErrors.E802(err))
        }
        err = dec.Servers[i].UpdateInDB(db)
        if err != nil {
    			appErrs = append(appErrs, APIErrors.E601(err))
    			return
    		}
      }
      var serversChangedI int
      serversChangedI, err = de.HaveServersChanged(db)
      if err != nil {
        appErrs = append(appErrs, APIErrors.E601(err))
        return
      }
      dec.ServersChanged = (serversChangedI == dao.SLStatus.Changed)
      dec.PreviousSslGrade, err = de.PreviousSSLgrade(db)
      if err != nil {
        appErrs = append(appErrs, APIErrors.E601(err))
        return
      }
    }
  }

	return
}

// Main function for listing recent evaluations.
// The function gets the recent domain evaluations, but using a time limiter to
// avoid overloading the server.
// The time limiter is present in the global vars RecentEvaluations,
// RecentEvaluationsLQ

func ListRecentEvaluations(currentHour time.Time, db *sql.DB) (apiErrs []APIError) {

  if len(RecentEvaluations) == 0 {
    var err error
    RecentEvaluations, err = dao.ListRecentDomainEvaluations(dao.DBConf)
  	apiErrs = make([]APIError,0)
  	if err != nil {
  		apiErrs = append(apiErrs, APIErrors.E601(err))
  	}
    RecentEvaluationsLQ = currentHour
    return
  }

  if RecentEvaluationsLQ.Add(RecentEvaluationsTW).Before(currentHour) {
    var err error
    RecentEvaluations, err = dao.ListRecentDomainEvaluations(dao.DBConf)
  	apiErrs = make([]APIError,0)
  	if err != nil {
  		apiErrs = append(apiErrs, APIErrors.E601(err))
  	}
    RecentEvaluationsLQ = currentHour
    return
  }

  apiErrs = make([]APIError, 0)
  return
}
