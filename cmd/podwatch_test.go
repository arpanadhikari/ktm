/*

Copywright: Arpan Adhikari

podwatch_test.go is a test file for podwatch.

*/

package cmd

import (
	"context"
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

func TestWatchEvents(t *testing.T) {
	db, _ := OpenPodHistoryDB()
	clientset := fake.NewSimpleClientset()
	stop := make(chan struct{})

	errCh := make(chan error, 1)
	go func() {
		errCh <- watchEvents(clientset, db, stop)
	}()

	// wait for informers to start
	time.Sleep(1 * time.Second)

	// create a new pod
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod-testwatchevents",
			Namespace: "default",
		},
	}

	// create a new node
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node-testwatchevents",
		},
	}
	_, err := clientset.CoreV1().Pods(pod.Namespace).Create(context.Background(), pod, metav1.CreateOptions{})
	assert.NoError(t, err)

	_, err = clientset.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})
	assert.NoError(t, err)

	// wait for the event to be processed
	time.Sleep(1 * time.Second)

	// check if the pod was added to the database
	podHistory, err := db.GetPodHistory(pod.Name)
	assert.NoError(t, err)
	assert.Equal(t, pod.Name, podHistory.Pod.Name)

	// check if the node was added to the database
	nodeHistory, err := db.GetNodeHistory(node.Name)
	assert.NoError(t, err)
	assert.Equal(t, node.Name, nodeHistory.Node.Name)

	// stop the watcher
	close(stop)

	// wait for the watcher to exit
	err = <-errCh
	assert.NoError(t, err)
}
