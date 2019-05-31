package rest

import (
	"encoding/json"
	"net/http"
	"time"
	"github.com/go-chi/chi"
	"github.com/trotsdeveloper/truora_test/truora_test_golang/dao"
	"github.com/trotsdeveloper/truora_test/truora_test_golang/controller"
)

type EvaluationResponse struct {
	Evaluation dao.ServerEvaluationComplete `json:"evaluation"`
	APIErrors []controller.APIError	`json:"errors"`
}

type PastEvaluationsResponse struct {
	Evaluations []dao.ServerEvaluation `json:"evaluations"`
	APIErrors []controller.APIError `json:"errors"`
}


func EvaluateServerEndPoint(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domainName")
	currentHour := time.Now()
	sec, apiErrs := controller.ScraperTestComplete(domain, currentHour, dao.DBConf)
	response := EvaluationResponse{Evaluation:sec, APIErrors:apiErrs}
	respB, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8") // normal header
	w.WriteHeader(http.StatusOK)
	w.Write(respB[:])
}

func ViewPastEvaluationsEndPoint(w http.ResponseWriter, r *http.Request) {
	sel, err := dao.ListRecentServerEvaluations(dao.DBConf)
	apiErrs := make([]controller.APIError,0)
	if err != nil {
		apiErrs = append(apiErrs, controller.APIErrors.E601(err))
	}
	response := PastEvaluationsResponse{Evaluations: sel, APIErrors:apiErrs}
	respB, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8") // normal header
	w.WriteHeader(http.StatusOK)
	w.Write(respB[:])
}
