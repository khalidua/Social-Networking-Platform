package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)
// This can help you test how the API Gateway handles slow responses from the users-service.
// follow instruction in file:///.//mock-slow-users-service.md to run this mock service. 

func main() {
	http.HandleFunc("/api/v1/users/test", func(w http.ResponseWriter, r *http.Request) {
		log.Println("received request, simulating slow users-service...")
		time.Sleep(5 * time.Second)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"service": "slow-users-service",
			"path":    r.URL.Path,
		})
	})

	log.Println("slow users mock running on :19082")
	log.Fatal(http.ListenAndServe(":19082", nil))
}