/*

Copywright: Arpan Adhikari

initWatch_test.go is a test file for initWatch.

*/

package cmd

import (
	"testing"
)

// TestWithTestPodHistoryDB is kinda stupid, but it's thought provoking. 🤔
func TestWithTestPodHistoryDB(t *testing.T) {
	withTestPodHistoryDB(t, func(phdb *PodHistoryDB, t *testing.T) {
		// do nothing
	})
}
