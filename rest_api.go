package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
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
	fmt.Println("LIST")
}

func main() {
	r := chi.NewRouter()

	// A good base middleware stack

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		mapD := map[string]int{"apple": 5, "lettuce": 7}
		mapB, _ := json.Marshal(mapD)
		mapJSON := string(mapB[:])
		fmt.Println(mapJSON)
		w.Header().Set("Content-Type", "application/json; charset=utf-8") // normal header
		w.WriteHeader(http.StatusOK)
		w.Write(mapB[:])
	})

	r.Route("/serverEvaluations", func(r chi.Router) {
		r.Get("/{domainName}", evaluateServerEndPoint)
		r.Get("/", viewPastEvaluationsEndPoint)
	})
	http.ListenAndServe(":3000", r)
}
