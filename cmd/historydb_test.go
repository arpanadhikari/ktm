/*

Copywright: Arpan Adhikari

podhistorydb_test.go is a test file for podhistorydb.

*/

package cmd

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestGetPods tests the getPods function
func TestGetPodHistory(t *testing.T) {

	withTestPodHistoryDB(t, func(phdb *PodHistoryDB, t *testing.T) {

		podName := "test-pod"

		testData := []PodHistory{
			// add pods only with name and creationtimestamp
			{Pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "namespace1", CreationTimestamp: metav1.Time{Time: time.Now()}}}},
			{Pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "namespace2", CreationTimestamp: metav1.Time{Time: time.Now()}}}},
			{Pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "namespace3", CreationTimestamp: metav1.Time{Time: time.Now()}}}},
			{Pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: "namespace4", CreationTimestamp: metav1.Time{Time: time.Now()}}}},
		}

		// Store the test data in the database
		for _, ph := range testData {
			err := phdb.AddPodHistory(ph)
			assert.NoError(t, err)
		}

		// Get the pod history
		podHistory, err := phdb.GetPodHistory(podName, "30s")
		assert.NoError(t, err)

		// assert number of pods
		assert.Equal(t, len(testData), len(podHistory))

		// Check if the pod history is correct
		for _, ph := range podHistory {
			fmt.Printf("Pod: %s, %s, %s\n", ph.Pod.ObjectMeta.Name, ph.Pod.ObjectMeta.Namespace, ph.Pod.ObjectMeta.CreationTimestamp.Format(time.RFC3339Nano))
			assert.Equal(t, podName, ph.Pod.Name)
		}

		//close db
		phdb.Close()
	})

}

func TestGetPodsRelativeTime(t *testing.T) {

	withTestPodHistoryDB(t, func(phdb *PodHistoryDB, t *testing.T) {
		defer phdb.Close()

		// Create sample PodHistory data
		now := time.Now()
		testData := []PodHistory{
			// add pods only with name and creationtimestamp
			{Pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod1", CreationTimestamp: metav1.Time{Time: now.Add(-8 * time.Hour)}}}},
			{Pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod2", CreationTimestamp: metav1.Time{Time: now.Add(-6 * time.Hour)}}}},
			{Pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod3", CreationTimestamp: metav1.Time{Time: now.Add(-2 * time.Hour)}}}},
			{Pod: v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "pod4", CreationTimestamp: metav1.Time{Time: now.Add(-1 * time.Hour)}}}},
		}

		// Store the test data in the database
		for _, ph := range testData {
			err := phdb.AddPodHistory(ph)
			assert.NoError(t, err)
		}

		// Test GetPods function with different relative times
		testCases := []struct {
			relativeTime  string
			expectedCount int
		}{
			{"9h", 4},
			{"7h", 3},
			{"3h20m", 2},
			{"50m", 0},
		}

		for _, tc := range testCases {
			pods, err := phdb.GetPodsRelativeTime(tc.relativeTime)
			assert.NoError(t, err)
			assert.Len(t, pods, tc.expectedCount, "Unexpected number of pods for relative time %s", tc.relativeTime)
		}
	})
}
