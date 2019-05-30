package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/trotsdeveloper/truora_test/truora_test_golang/dao"
	"github.com/trotsdeveloper/truora_test/truora_test_golang/rest"

)

func main() {
	var err error
	dao.DBConf, err = dao.InitDB()
	fmt.Println("Detect changes")
	if err != nil {
		fmt.Println("Error initing DB")
	} else {
		r := chi.NewRouter()

		// A good base middleware stack

		r.Get("/", rest.TestEndPoint)

		r.Route("/serverEvaluations", func(r chi.Router) {
			r.Get("/{domainName}", rest.EvaluateServerEndPoint)
			r.Get("/", rest.ViewPastEvaluationsEndPoint)
		})
		http.ListenAndServe(":3000", r)
	}
}
