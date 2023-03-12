/*

Copywright: Arpan Adhikari

Podwatch_test.go is a test file for podwatch.

*/

package cmd

import (
	"context"
	"sync"
	"testing"
	"time"

	// import kubernetes pod and node packages

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	// assert package for testing
	"github.com/stretchr/testify/assert"
	// require package for testing
	// "github.com/stretchr/testify/require"
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
			Spec: v1.PodSpec{
				NodeName: nodeName,
				Containers: []v1.Container{
					{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceCPU:    *resource.NewMilliQuantity(100, resource.DecimalSI),
								v1.ResourceMemory: *resource.NewQuantity(100*1024*1024, resource.BinarySI),
							},
						},
					},
				},
			},
		},
		&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "test-node",
				CreationTimestamp: metav1.Time{Time: time.Now()},
			},
			Status: v1.NodeStatus{
				Allocatable: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(1000, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(1*1024*1024*1024, resource.BinarySI),
				},
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

	pod, err := db.GetPodHistory(podName)
	if err != nil {
		t.Errorf("failed to get pod from database: %v", err)
	}
	assert.Equal(t, pod.PodName, podName, "pod name is not the same as the one we created")

	node, err := db.GetNodeHistory(nodeName)
	if err != nil {
		t.Errorf("failed to get node from database: %v", err)
	}
	assert.Equal(t, node.NodeName, nodeName, "node name is not the same as the one we created")

	db.Close()

}

// TestWatchEvents tests the watchEvents function
func TestWatchEvents(t *testing.T) {

	clientset := fake.NewSimpleClientset()
	db, _ := OpenPodHistoryDB()

	stop := make(chan struct{})
	// close(stop)

	// use a wait group to wait for the goroutine to complete
	var wg sync.WaitGroup
	wg.Add(1)

	// run watchEvents
	go func() {
		defer wg.Done()
		watchEvents(clientset, db, stop)
	}()

	// create a new pod
	podName := "test-pod-testwatchevents"
	podNamespace := "test-namespace"

	// trigger a fake event
	// clientset.CoreV1().Events(podNamespace).Create(context.TODO(), &v1.Event{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:              "test-event",
	// 		Namespace:         podNamespace,
	// 		CreationTimestamp: metav1.Time{Time: time.Now()},
	// 	},
	// 	InvolvedObject: v1.ObjectReference{
	// 		Kind:      "Pod",
	// 		Name:      podName,
	// 		Namespace: podNamespace,
	// 	},
	// 	Reason:  "test-reason",
	// 	Message: "test-message",
	// }, metav1.CreateOptions{})

	// send the new pod to the fake clientset
	go clientset.CoreV1().Pods(podNamespace).Create(context.TODO(), &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              podName,
			Namespace:         podNamespace,
			CreationTimestamp: metav1.Time{Time: time.Now()},
			UID:               "test-uid",
		},
		Spec: v1.PodSpec{
			NodeName: "test-node",
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceCPU:    *resource.NewMilliQuantity(100, resource.DecimalSI),
							v1.ResourceMemory: *resource.NewQuantity(100*1024*1024, resource.BinarySI),
						},
					},
				},
			},
		},
	}, metav1.CreateOptions{})

	// wait for the goroutine to complete before checking the state of the database
	wg.Wait()

	// wait for 10 seconds
	// time.Sleep(10 * time.Second)

	// get the pod from the database
	pod, err := db.GetPodHistory(podName)
	if err != nil {
		t.Errorf("failed to get pod from database: %v", err)
	}
	db.Close()

	close(stop)

	// check if the pod name is the same as the one we created
	assert.Equal(t, pod.PodName, podName, "pod name is not the same as the one we created %v: ", pod)
}

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
