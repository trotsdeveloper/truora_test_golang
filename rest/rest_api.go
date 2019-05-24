package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/trotsdeveloper/truora_test/truora_test_golang/dao"
	"github.com/trotsdeveloper/truora_test/truora_test_golang/scrapers"
	//"github.com/trotsdeveloper/truora_test/truora_test_golang/dao"
)

func TestEndPoint(w http.ResponseWriter, r *http.Request) {
	//articleID := chi.URLParam(r, "articleID")
	re := regexp.MustCompile(`a(x*)b`)
	fmt.Printf("%q\n", re.FindAllStringSubmatch("-ab-", -1))
	fmt.Printf("%q\n", re.FindAllStringSubmatch("-axxb-", -1))
	fmt.Printf("%q\n", re.FindAllStringSubmatch("-ab-axb-", -1))
	fmt.Printf("%q\n", re.FindAllStringSubmatch("-axxb-ab-", -1))

	domain := "54.239.132.139"
	country, err := scrapers.ScraperCountry(domain)
	serverOwner, err := scrapers.ScraperOwner(domain)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("country:%v, server_owner:%v", country, serverOwner)
	}

	//result2, err := whoisparser.Parse(result)
	//fmt.Println(fmt.Sprintf("Name: %v", result2.Registrar.DomainName))

	// Between START AND END // SEARCH ORG NAME AND COUNTRY
	//

	mapD := map[string]int{"apple": 5, "lettuce": 7}
	mapB, _ := json.Marshal(mapD)
	mapJSON := string(mapB[:])
	fmt.Println(mapJSON)
	w.Header().Set("Content-Type", "application/json; charset=utf-8") // normal header
	w.WriteHeader(http.StatusOK)
	w.Write(mapB[:])
}

func EvaluateServerEndPoint(w http.ResponseWriter, r *http.Request) {

	//http.Get(svr.URL)
	mapJson := map[string]int{"TEST": 1}
	mapB, _ := json.Marshal(mapJson)
	mapS := string(mapB[:])
	fmt.Println(mapS)
	w.Header().Set("Content-Type", "application/json; charset=utf-8") // normal header
	w.WriteHeader(http.StatusOK)
	w.Write(mapB[:])
}

func ViewPastEvaluationsEndPoint(w http.ResponseWriter, r *http.Request) {
	db, err := dao.InitDB()
	if err != nil {
		fmt.Printf("Exception: %v", err)
	}
	defer db.Close()

	serverEvaluationList, err := dao.ListRecentServerEvaluations(db)
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
