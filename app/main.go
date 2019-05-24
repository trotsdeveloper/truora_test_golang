package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/trotsdeveloper/truora_test/truora_test_golang/rest"
)

func main() {
	r := chi.NewRouter()

	// A good base middleware stack

	r.Get("/", rest.TestEndPoint)

	r.Route("/serverEvaluations", func(r chi.Router) {
		r.Get("/{domainName}", rest.EvaluateServerEndPoint)
		r.Get("/", rest.ViewPastEvaluationsEndPoint)
	})
	http.ListenAndServe(":3000", r)
}
