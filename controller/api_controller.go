package controller

import (
  "time"
  "github.com/trotsdeveloper/truora_test/truora_test_golang/scrapers"
  "github.com/trotsdeveloper/truora_test/truora_test_golang/dao"
)

var APIErrors = newAPIErrorsRegistry()

func newAPIErrorsRegistry() *apiErrorsRegistry {
	e601v := makeAPIError("601", "Error in database.")
	e602v := makeAPIError("602", "Error in SSLabs API.")
	e701v := makeAPIError("701", "Error getting Icon")
	e702v := makeAPIError("702", "Error getting HTML Title")
	e801v := makeAPIError("801", "Error getting country from WHOIS")
	e802v := makeAPIError("802", "Error getting owner from WHOIS")

	return &apiErrorsRegistry{
		e601: e601v,
		e602: e602v,
		e701: e701v,
		e702: e702v,
		e801: e801v,
		e802: e802v,
	}
}

func makeAPIError(code string, description string) func(error) (APIError) {
	return func(err error) (APIError) {
		return APIError{code, description, err.Error()}
	}
}

type apiErrorsRegistry struct {
	e601 func(error) (APIError) //
	e602 func(error) (APIError) //
	e701 func(error) (APIError) //
	e702 func(error) (APIError) //
	e801 func(error) (APIError) //
	e802 func(error) (APIError) //
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
	dbc interface{}) (se dao.ServerEvaluation, appErr APIError) {
	// 1) In the database, is there a server Evaluation in process
	// with the same given domain?
  appErr = DefaultAPIError()
  var err error
	pendingEvaluation := dao.ServerEvaluation{}
	err = pendingEvaluation.SearchLastEvaluation(domainName, true, currentHour, dbc)
	if err != nil {
    appErr = APIErrors.e601(err)
		return
	}
	if pendingEvaluation.Id != 0 {
		// 1.1) YES: Is difference between the current hour
		// and the pending Evaluation lower than 20 seconds?
		var pendingEvaluationHour time.Time
		pendingEvaluationHour, err = time.Parse(time.RFC3339, pendingEvaluation.EvaluationHour)
		if err != nil {
      appErr = APIErrors.e601(err)
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
			err = pendingEvaluation.UpdateHourInDb(dbc)
			if err != nil {
        appErr = APIErrors.e601(err)
				return
			}

			var currentEvaluation dao.ServerEvaluation
			currentEvaluation, err = evaluator(currentHour, domainName)
			if err != nil {
        appErr = APIErrors.e602(err)
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
				err = currentEvaluation.UpdateInDB(dbc)
        if err != nil {
          appErr = APIErrors.e601(err)
          return
        }
				se = currentEvaluation
				return
			}
		}
	} else {
		// 1.2) NO: Is there a past server Evaluation, ready, with the same given domain?
		pastEvaluation := dao.ServerEvaluation{}
		err = pastEvaluation.SearchLastEvaluation(domainName, false, currentHour, dbc)
    if err != nil {
      appErr = APIErrors.e601(err)
      return
    }
		if pastEvaluation.Id != 0 {
			// 1.2.1) YES: Is difference lower than 20 seconds?
			var pastEvaluationHour time.Time
			pastEvaluationHour, err = time.Parse(time.RFC3339, pastEvaluation.EvaluationHour)
			if err != nil {
        appErr = APIErrors.e601(err)
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
          appErr = APIErrors.e602(err)
					return
				}
				err = currentEvaluation.CreateInDB(dbc)
        if err != nil {
          appErr =  APIErrors.e601(err)
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
        appErr = APIErrors.e602(err)
				return
			}
			err = currentEvaluation.CreateInDB(dbc)
      if err != nil {
        appErr = APIErrors.e601(err)
        return
      }
			se = currentEvaluation
			return
		}
	}
}

func ScraperTestComplete(domain string, currentHour time.Time, dbc interface{}) (sec dao.ServerEvaluationComplete, appErrs []APIError) {

	sec = dao.ServerEvaluationComplete{}
	appErrs = make([]APIError, 0)

	se, appErr := MakeEvaluationInDomain(domain, currentHour, scrapers.ScraperSSLabs, dbc)
  defaultCode := DefaultAPIError()
	if !(appErr.Code == defaultCode.Code) {
		appErrs = append(appErrs, appErr)
	}
  var err error
	sec.Copy(se)

	if !se.IsDown {
		sec.Logo, err = scrapers.ScraperLogo(domain)
		if err != nil {
			appErrs = append(appErrs, APIErrors.e701(err))
		}
		sec.Title, err = scrapers.ScraperTitle(domain)
		if err != nil {
			appErrs = append(appErrs, APIErrors.e702(err))
		}
	}

	if !se.EvaluationInProgress && !se.IsDown {
		for i := range sec.Servers {
			ip := sec.Servers[i].Address
			sec.Servers[i].Country, err = scrapers.ScraperCountry(ip)
			if err != nil {
				appErrs = append(appErrs, APIErrors.e801(err))
			}
			sec.Servers[i].Owner, err = scrapers.ScraperOwner(ip)
			if err != nil {
				appErrs = append(appErrs, APIErrors.e802(err))
			}
		}
		var serversChangedI int
		serversChangedI, err = se.HaveServersChanged(dbc)
		if err != nil {
			appErrs = append(appErrs, APIErrors.e601(err))
			return
		}
		sec.ServersChanged = (serversChangedI == dao.SLStatus.Changed)
		sec.PreviousSslGrade, err = se.PreviousSSLgrade(dbc)
		if err != nil {
			appErrs = append(appErrs, APIErrors.e601(err))
      return
		}
	}

	return
}
