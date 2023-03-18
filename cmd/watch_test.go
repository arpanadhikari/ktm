/*

Copywright: Arpan Adhikari

initWatch_test.go is a test file for initWatch.

*/

package cmd

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// TestInitWatch tests the initWatch function
func TestInitWatch(t *testing.T) {

	withTestPodHistoryDB(t, func(db *PodHistoryDB, t *testing.T) {

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

		err := initWatch(clientset, db)
		if err != nil {
			t.Errorf("failed to execute initWatch: %v", err)
		}

		podHistory, err := db.GetPodHistory(podName)
		if err != nil {
			t.Errorf("failed to get pod from database: %v", err)
		}
		assert.Equal(t, podHistory[0].Pod.ObjectMeta.Name, podName, "pod name is not the same as the one we created")

		nodeHistory, err := db.GetNodeHistory(nodeName)
		if err != nil {
			t.Errorf("failed to get node from database: %v", err)
		}
		assert.Equal(t, nodeHistory[0].Node.ObjectMeta.Name, nodeName, "node name is not the same as the one we created")

		db.Close()

	})

}

func TestWatchEvents(t *testing.T) {

	withTestPodHistoryDB(t, func(db *PodHistoryDB, t *testing.T) {

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
				Name:              "test-pod-testwatchevents",
				Namespace:         "default",
				CreationTimestamp: metav1.Time{Time: time.Now()},
			},
		}

		// create a new node
		node := &v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "test-node-testwatchevents",
				CreationTimestamp: metav1.Time{Time: time.Now()},
			},
		}

		/*
			Test scenario:
			1. Create a pod and a node
			2. Update the pod and the node with new labels, mutate the pod with an init container+new host
			3. Delete the pod and the node

			Expected result:
			1. The pod and node should be added to the database
			2. The updated pod and node should be added to the database, count of pods should be 3, nodes should be 2
			3. The pod and node deleted history event should be "Deleted", count of pods should be 4, nodes should be 3

		*/

		// Test 1: Create a pod and a node
		_, err := clientset.CoreV1().Pods(pod.Namespace).Create(context.Background(), pod, metav1.CreateOptions{})
		assert.NoError(t, err)
		_, err = clientset.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})
		assert.NoError(t, err)
		time.Sleep(1 * time.Second)
		podHistory, err := db.GetPodHistory(pod.Name)
		assert.NoError(t, err)
		assert.Equal(t, pod.Name, podHistory[0].Pod.Name)
		assert.Equal(t, 1, len(podHistory))
		nodeHistory, err := db.GetNodeHistory(node.Name)
		assert.NoError(t, err)
		assert.Equal(t, node.Name, nodeHistory[0].Node.Name)
		assert.Equal(t, 1, len(nodeHistory))

		// Test 2: Update the pod and the node with new labels, mutate the pod with an init container+new host
		pod.Labels = map[string]string{"test": "test"}
		pod.Spec = v1.PodSpec{
			InitContainers: []v1.Container{{
				Name:  "test-init-container",
				Image: "test-image"},
			}, NodeName: "test-node"}
		_, err = clientset.CoreV1().Pods(pod.Namespace).Update(context.Background(), pod, metav1.UpdateOptions{})
		assert.NoError(t, err)
		node.Labels = map[string]string{"test": "test"}
		_, err = clientset.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
		assert.NoError(t, err)
		time.Sleep(1 * time.Second)
		podHistory, err = db.GetPodHistory(pod.Name)

		fmt.Printf("podHistory: %v", podHistory)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(podHistory))
		assert.Equal(t, pod.Name, podHistory[1].Pod.Name)
		assert.Equal(t, pod.Spec.NodeName, podHistory[1].Pod.Spec.NodeName)
		nodeHistory, err = db.GetNodeHistory(node.Name)
		assert.Equal(t, 2, len(nodeHistory))
		assert.NoError(t, err)
		assert.Equal(t, node.Name, nodeHistory[1].Node.Name)

		// Test 3: Delete the pod and the node
		err = clientset.CoreV1().Pods(pod.Namespace).Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
		assert.NoError(t, err)
		err = clientset.CoreV1().Nodes().Delete(context.Background(), node.Name, metav1.DeleteOptions{})
		assert.NoError(t, err)
		time.Sleep(1 * time.Second)
		podHistory, err = db.GetPodHistory(pod.Name)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(podHistory))
		assert.Equal(t, pod.Name, podHistory[2].Pod.Name)
		assert.Equal(t, podHistory[2].Event.Type, "Deleted")
		nodeHistory, err = db.GetNodeHistory(node.Name)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(nodeHistory))
		assert.Equal(t, node.Name, nodeHistory[2].Node.Name)
		assert.Equal(t, nodeHistory[2].Event.Type, "Deleted")

		// stop the watcher
		close(stop)

		// wait for the watcher to exit
		err = <-errCh
		assert.NoError(t, err)

	})
}
