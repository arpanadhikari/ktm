/*

Copywright: Arpan Adhikari

podhistorydb_test.go is a test file for podhistorydb.

*/

package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		Pod: v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:              podName,
				Namespace:         podNamespace,
				CreationTimestamp: metav1.Time{Time: time.Now()},
			},
		},
	})
	if err != nil {
		t.Errorf("failed to add pod to database: %v", err)
	}

	podHistory, err := db.GetPodHistory(podName)
	if err != nil {
		t.Errorf("failed to get pod from database: %v", err)
	}

	assert.Equal(t, podHistory.Pod.ObjectMeta.Name, podName, "pod name is not the same as the one we created")

	//close db
	db.Close()

}
