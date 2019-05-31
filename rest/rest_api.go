package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	//"github.com/trotsdeveloper/truora_test/truora_test_golang/dao"
	"github.com/go-chi/chi"
	"github.com/trotsdeveloper/truora_test/truora_test_golang/dao"
	"github.com/trotsdeveloper/truora_test/truora_test_golang/controller"
	//"github.com/trotsdeveloper/truora_test/truora_test_golang/dao"
)

type Response struct {
	Evaluation dao.ServerEvaluationComplete `json:"evaluation"`
	APIErrors []controller.APIError	`json:"errors"`
}


func EvaluateServerEndPoint(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domainName")
	currentHour := time.Now()
	sec, apiErrs := controller.ScraperTestComplete(domain, currentHour, dao.DBConf)
	response := Response{}
	response.Evaluation = sec
	response.APIErrors = apiErrs
	respB, _ := json.Marshal(response)
	w.Header().Set("Content-Type", "application/json; charset=utf-8") // normal header
	w.WriteHeader(http.StatusOK)
	w.Write(respB[:])
}

func ViewPastEvaluationsEndPoint(w http.ResponseWriter, r *http.Request) {

	serverEvaluationList, err := dao.ListRecentServerEvaluations(dao.DBConf)
	if err != nil {
		fmt.Printf("Exception: %v", err)
	}
	evalListBytes, err := json.Marshal(serverEvaluationList)
	if err != nil {
		fmt.Printf("Exception: %v", err)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8") // normal header
	w.WriteHeader(http.StatusOK)
	w.Write(evalListBytes[:])
}
