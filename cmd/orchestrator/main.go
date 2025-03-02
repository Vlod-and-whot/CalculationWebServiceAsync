package main

import (
	"log"
	"net/http"
	"os"

	"CalcWebServiceAsync/internal/application"
)

func main() {
	port := os.Getenv("ORCHESTRATOR_PORT")
	if port == "" {
		port = "8080"
	}

	app := application.NewOrchestrator()

	http.HandleFunc("/api/v1/calculate", app.HandleCalculate)
	http.HandleFunc("/api/v1/expressions", app.HandleListExpressions)
	http.HandleFunc("/api/v1/expressions/", app.HandleGetExpression)
	http.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			app.HandleGetTask(w, r)
		} else if r.Method == http.MethodPost {
			app.HandlePostTask(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Printf("Orchestrator listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
