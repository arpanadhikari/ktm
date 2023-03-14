/*

Copywright: Arpan Adhikari

podhistorydb_test.go is a test file for podhistorydb.

*/

package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetPods tests the getPods function
func TestGetPods(t *testing.T) {

	podName := "test-pod-testgetpods"
	podNamespace := "test-namespace"

	db, err := OpenPodHistoryDB()

	if err != nil {
		t.Errorf("failed to open database: %v", err)
	}

	// add podhistory to db
	err = db.AddPodHistory(PodHistory{
		PodName:      podName,
		PodNamespace: podNamespace,
	})
	if err != nil {
		t.Errorf("failed to add pod to database: %v", err)
	}

	pod, err := db.GetPodHistory(podName)
	if err != nil {
		t.Errorf("failed to get pod from database: %v", err)
	}

	assert.Equal(t, pod.PodName, podName, "pod name is not the same as the one we created")

	//close db
	db.Close()

}
