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
		   	- /clustersnapshot
			- /podhistory/{podname}
			- /nodehistory/{nodename}
*/
func StartWebServer(db *PodHistoryDB, stop chan struct{}) {
	fmt.Println("Starting web server")

	handler := &Handler{DB: db}

	// register handling of / to serve cmd/html/index.html
	http.Handle("/", http.FileServer(http.Dir("./cmd/ktmweb/build")))
	// http.Handle("/", http.FileServer(http.Dir("./cmd/html")))

	// handle routes
	// http.HandleFunc("/podhistory", handler.handlePodHistory)
	// handle /podhistory/{podname}
	http.HandleFunc("/clustersnapshot", handler.handleClusterSnapshot)
	http.HandleFunc("/podhistory/all/", handler.handleGetPods)
	http.HandleFunc("/podhistory/", handler.handlePodHistory)
	http.HandleFunc("/nodehistory/", handler.handleNodeHistory)

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

func (h *Handler) handleClusterSnapshot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received %s to %s\n", r.Method, r.URL.Path)
	// send http headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// get relative time from query
	relativeTime := r.URL.Query().Get("relativeTime")

	// get cluster snapshot from db
	clusterSnapshot, err := h.DB.GetReconciledClusterSnapshot("127.0.0.1:6443", relativeTime)
	if err != nil {
		fmt.Println(err)
	}

	data := formatToD3Tree(&clusterSnapshot)

	// Send json encoded response
	json.NewEncoder(w).Encode(data)

}

func (h *Handler) handleGetPods(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("Received %s to %s\n", r.Method, r.URL.Path)
	// send http headers
	w.Header().Set("Content-Type", "application/json")

	// get relative time from query
	relativeTime := r.URL.Query().Get("relativeTime")

	// get pod history from db
	podHistory, err := h.DB.GetPodsRelativeTime(relativeTime)

	if err != nil {
		fmt.Println(err)
	}

	pods := map[string]interface{}{
		"name":     "pods",
		"children": []interface{}{},
	}

	for _, pod := range podHistory {
		podMap := map[string]interface{}{
			"name": pod.Pod.ObjectMeta.Name,
			// "children": []interface{}{},
			"event": pod.Event.Type,
			"event-time": map[string]interface{}{
				"first": pod.Event.FirstTimestamp.Time,
			},
		}

		pods["children"] = append(pods["children"].([]interface{}), podMap)
	}

	// Send json encoded response
	json.NewEncoder(w).Encode(pods)

}

func (h *Handler) handlePodHistory(w http.ResponseWriter, r *http.Request) {

	fmt.Printf("Received %s to %s\n", r.Method, r.URL.Path)
	// send http headers
	w.Header().Set("Content-Type", "application/json")

	// extract podname from /podhistory/{podname}
	podname := r.URL.Path[len("/podhistory/"):]

	// extract relative time from query
	relativeTime := r.URL.Query().Get("relativeTime")

	// get pod history from db
	podHistory, err := h.DB.GetPodHistory(podname, relativeTime)

	if err != nil {
		fmt.Println(err)
	}

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

	// extract nodename from /nodehistory/{nodename}
	nodename := r.URL.Path[len("/nodehistory/"):]

	// extract relative time from query
	relativeTime := r.URL.Query().Get("relativeTime")

	nodes := map[string]interface{}{
		"name":     "nodes",
		"children": []interface{}{},
	}

	// get node history from db
	nodeHistory, err := h.DB.GetNodeHistory(nodename, relativeTime)
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

// d3 formatting functions

// formatToD3Tree formats the data to a format that d3.js can use
func formatToD3Tree(cs *ClusterSnapshot) map[string]interface{} {

	// set root name to pods and nodes
	data := []interface{}{
		map[string]interface{}{
			"name":     "pods",
			"children": []interface{}{},
		},
		map[string]interface{}{
			"name":     "nodes",
			"children": []interface{}{},
		},
	}

	// add pods to data
	for _, pod := range cs.Pods {
		podMap := map[string]interface{}{
			"name":     pod.Name,
			"nodename": pod.Spec.NodeName,
			"size": map[string]interface{}{
				// get cpu value in m omits the m
				"cpu":    pod.Spec.Containers[0].Resources.Requests.Cpu().Value(),
				"memory": pod.Spec.Containers[0].Resources.Requests.Memory().Value() / 1024 / 1024 / 1024,
			},
		}
		data[0].(map[string]interface{})["children"] = append(data[0].(map[string]interface{})["children"].([]interface{}), podMap)
	}

	// add nodes to data
	for _, node := range cs.Nodes {
		nodeMap := map[string]interface{}{
			"name": node.Name,
			"size": map[string]interface{}{
				"cpu": node.Status.Capacity.Cpu().Value(),
				// set memory in Gi
				"memory": node.Status.Capacity.Memory().Value() / 1024 / 1024 / 1024,
			},
		}
		data[1].(map[string]interface{})["children"] = append(data[1].(map[string]interface{})["children"].([]interface{}), nodeMap)
	}

	return map[string]interface{}{
		"data": data,
	}
}
