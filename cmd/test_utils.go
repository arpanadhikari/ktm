/*

Copywright: Arpan Adhikari

initWatch_test.go is a test file for initWatch.

*/

package cmd

import (
	"io/ioutil"
	"os"
	"testing"
)

func withTestPodHistoryDB(t *testing.T, testFunc func(phdb *PodHistoryDB, t *testing.T)) {
	tempFile, err := ioutil.TempFile("", "test_podhistorydb")
	if err != nil {
		t.Fatal("Failed to create temporary file:", err)
	}
	t.Cleanup(func() {
		os.Remove(tempFile.Name())
	})

	phdb, err := OpenPodHistoryDBWithFile(tempFile.Name())
	if err != nil {
		t.Fatal("Failed to open test database:", err)
	}
	t.Cleanup(func() {
		phdb.Close()
	})

	testFunc(phdb, t)
}
