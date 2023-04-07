package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Handler struct {
	DB *PodHistoryDB
}

/*
		StartWebServer starts the web server,
	   	By default it listens on port 8080 and handles the following routes:
			- /podhistory
			- /nodehistory
*/
func StartWebServer(db *PodHistoryDB, stop chan struct{}) {
	fmt.Println("Starting web server")

	handler := &Handler{DB: db}

	// register handling of / to serve cmd/html/index.html
	http.Handle("/", http.FileServer(http.Dir("./cmd/html")))

	// handle routes
	http.HandleFunc("/podhistory", handler.handlePodHistory)
	http.HandleFunc("/nodehistory", handler.handleNodeHistory)
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

func (h *Handler) handlePodHistory(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("Received %s to %s\n", r.Method, r.URL.Path)
	// send http headers
	w.Header().Set("Content-Type", "application/json")

	// get pod history from db
	podHistory, err := h.DB.GetPodsRelativeTime("10h")
	if err != nil {
		fmt.Println(err)
	}

	// parse array of pods in podHistory to format required for d3.js
	// {
	// name: "pods",
	// children: [
	//     {
	//     name: "pod11",
	//     size: {
	//         cpu: 1,
	//         memory: 2
	//     	}
	//   },]
	// }
	pods := map[string]interface{}{
		"name":     "pods",
		"children": []interface{}{},
	}

	for _, pod := range podHistory {
		podMap := map[string]interface{}{
			"name":     pod.Pod.Name,
			"nodename": pod.Pod.Spec.NodeName,
			"size": map[string]interface{}{
				// get cpu value in m omits the m
				"cpu":    pod.Pod.Spec.Containers[0].Resources.Requests.Cpu().Value(),
				"memory": pod.Pod.Spec.Containers[0].Resources.Requests.Memory().Value() / 1024 / 1024 / 1024,
			},
			"timestamp": getTimestampSuffix(&pod),
		}
		pods["children"] = append(pods["children"].([]interface{}), podMap)
	}

	// send json encoded response
	json.NewEncoder(w).Encode(pods)

	// send response with the header as application/json
	// json.NewEncoder(w).Encode(map[string]string{"message": "Pod History"})
}

func (h *Handler) handleNodeHistory(w http.ResponseWriter, r *http.Request) {

	// print a log
	fmt.Printf("Received %s to %s\n", r.Method, r.URL.Path)

	// send http headers
	w.Header().Set("Content-Type", "application/json")

	nodes := map[string]interface{}{
		"name":     "nodes",
		"children": []interface{}{},
	}

	// get node history from db
	nodeHistory, err := h.DB.GetNodesRelativeTime("10m")
	if err != nil {
		fmt.Println(err)
	}

	for _, node := range nodeHistory {
		nodeMap := map[string]interface{}{
			"name": node.Node.Name,
			"size": map[string]interface{}{
				"cpu": node.Node.Status.Capacity.Cpu().Value(),
				// set memory in Gi
				"memory": node.Node.Status.Capacity.Memory().Value() / 1024 / 1024 / 1024,
			},
			"timestamp": getTimestampSuffix(&node),
		}
		nodes["children"] = append(nodes["children"].([]interface{}), nodeMap)
	}
	// send response
	json.NewEncoder(w).Encode(nodes)

	// send response
	// json.NewEncoder(w).Encode(map[string]string{"message": "Node Historyx"})
}
