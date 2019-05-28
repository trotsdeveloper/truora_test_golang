package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	//"github.com/trotsdeveloper/truora_test/truora_test_golang/dao"
	"github.com/go-chi/chi"
	"github.com/trotsdeveloper/truora_test/truora_test_golang/dao"
	"github.com/trotsdeveloper/truora_test/truora_test_golang/scrapers"
	//"github.com/trotsdeveloper/truora_test/truora_test_golang/dao"
)

func TestEndPoint(w http.ResponseWriter, r *http.Request) {
	err := dao.CleanDataInDB(dao.DBConf)
	if err != nil {
		fmt.Println(fmt.Sprintf("Err: %v, err", err))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Clean BD"))
}

func EvaluateServerEndPoint(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.URL.String())

	domain := chi.URLParam(r, "domainName")
	fmt.Println(domain)
	currentHour := time.Now()
	fmt.Println(fmt.Sprintf("SSLabs hour: %v", currentHour.Format(time.RFC3339)))
	sec, err := scrapers.ScraperTestComplete(domain, currentHour, dao.DBConf)
	if err != nil {
		fmt.Println(fmt.Sprintf("Err: %v", err))
		return
	}

	secB, _ := json.Marshal(sec)
	secS := string(secB[:])
	fmt.Println(secS)
	w.Header().Set("Content-Type", "application/json; charset=utf-8") // normal header
	w.WriteHeader(http.StatusOK)
	w.Write(secB[:])
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
