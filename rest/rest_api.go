// Package for the declaration of the REST API
// The package contains the response structures and endpoints of the rest apis
package rest

import (
	"encoding/json"
	"net/http"
	"time"
	"github.com/go-chi/chi"
	"github.com/trotsdeveloper/truora_test/truora_test_golang/dao"
	"github.com/trotsdeveloper/truora_test/truora_test_golang/controller"
)

// Structure representing a response in the EvaluateDomainEndpoint
type EvaluationResponse struct {
	Evaluation dao.DomainEvaluationComplete `json:"evaluation"`
	APIErrors []controller.APIError	`json:"errors"`
}

// Structure representing a response in the ViewPastEvaluationsEndPoint
type PastEvaluationsResponse struct {
	Evaluations []dao.DomainEvaluation `json:"evaluations"`
	APIErrors []controller.APIError `json:"errors"`
}

func EvaluateDomainEndPoint(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domainName")
	currentHour := time.Now()
	sec, apiErrs := controller.ScraperTestComplete(domain, currentHour, dao.DBConf)
	response := EvaluationResponse{Evaluation:sec, APIErrors:apiErrs}
	respB, _ := json.Marshal(response)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8") // normal header
	w.WriteHeader(http.StatusOK)
	w.Write(respB[:])
}

func ViewPastEvaluationsEndPoint(w http.ResponseWriter, r *http.Request) {

	currentHour := time.Now()
	apiErrs := controller.ListRecentEvaluations(currentHour, dao.DBConf)
	response := PastEvaluationsResponse{Evaluations: controller.RecentEvaluations, APIErrors:apiErrs}
	respB, _ := json.Marshal(response)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8") // normal header
	w.WriteHeader(http.StatusOK)
	w.Write(respB[:])
}
