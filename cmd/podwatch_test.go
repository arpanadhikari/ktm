/*

Copywright: Arpan Adhikari

Podwatch_test.go is a test file for podwatch.

*/

package cmd

import (
	"testing"
	"time"

	// import kubernetes pod and node packages

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
