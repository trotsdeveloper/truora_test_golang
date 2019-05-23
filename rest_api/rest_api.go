package rest_api

import (
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/trotsdeveloper/truora_test/truora_test_golang/dao"
)

func evaluateServerEndPoint(w http.ResponseWriter, r *http.Request) {
	mapJson := map[string]int{"TEST": 1}
	mapB, _ := json.Marshal(mapJson)
	mapS := string(mapB[:])
	fmt.Println(mapS)
	w.Header().Set("Content-Type", "application/json; charset=utf-8") // normal header
	w.WriteHeader(http.StatusOK)
	w.Write(mapB[:])
}

func viewPastEvaluationsEndPoint(w http.ResponseWriter, r *http.Request) {
	db, err := dao.InitDB()
	if err != nil {
		fmt.Printf("Exception: %v", err)
	}
	defer db.Close()

	serverEvaluationList, err := dao.ServerEvaluationListFactory(db)
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
