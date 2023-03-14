/*

Copywright: Arpan Adhikari

podwatch_test.go is a test file for podwatch.

*/

package cmd

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// TestPodWatch tests the podwatch function
func TestPodWatch(t *testing.T) {

	podName := "test-pod"
	podNamespace := "test-namespace"

	nodeName := "test-node"

	clientset := fake.NewSimpleClientset(
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:              podName,
				Namespace:         podNamespace,
				CreationTimestamp: metav1.Time{Time: time.Now()},
				UID:               "test-uid",
			},
		},
		&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "test-node",
				CreationTimestamp: metav1.Time{Time: time.Now()},
			},
		},
	)

	db, err := OpenPodHistoryDB()
	if err != nil {
		t.Errorf("failed to open database: %v", err)
	}

	err = podWatch(clientset, db)
	if err != nil {
		t.Errorf("failed to execute podwatch: %v", err)
	}

	podHistory, err := db.GetPodHistory(podName)
	if err != nil {
		t.Errorf("failed to get pod from database: %v", err)
	}
	assert.Equal(t, podHistory.Pod.ObjectMeta.Name, podName, "pod name is not the same as the one we created")

	nodeHistory, err := db.GetNodeHistory(nodeName)
	if err != nil {
		t.Errorf("failed to get node from database: %v", err)
	}
	assert.Equal(t, nodeHistory.Node.ObjectMeta.Name, nodeName, "node name is not the same as the one we created")

	db.Close()

}

// TODO: Add test for podwatch function.
