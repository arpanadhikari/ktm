package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
)

/*
	StartWebServer starts the web server,
   	By default it listens on port 8080 and handles the following routes:
		- /podhistory
		- /nodehistory
*/
func StartWebServer(stop chan struct{}) {
	fmt.Println("Starting web server")

	http.HandleFunc("/podhistory", handlePodHistory)
	http.HandleFunc("/nodehistory", handleNodeHistory)
	fmt.Println("Starting server on :8080")
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			panic(err)
		}
	}()

	// wait for stop signal
	<-stop

}

func handlePodHistory(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Pod History")
	// send http headers
	w.Header().Set("Content-Type", "application/json")
	// send response
	json.NewEncoder(w).Encode(map[string]string{"message": "Pod History"})
}

func handleNodeHistory(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Node History")
	// send http headers
	w.Header().Set("Content-Type", "application/json")
	// send response
	json.NewEncoder(w).Encode(map[string]string{"message": "Node History"})
}
