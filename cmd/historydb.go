package cmd

// This package is a wrapper around the BoltDB key/value store. It provides
// a simple interface for storing and retrieving pod history data.

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	bolt "go.etcd.io/bbolt"
)

// PodHistoryDB is a wrapper around the BoltDB key/value store.
type PodHistoryDB struct {
	db *bolt.DB
}

// PodHistory is a struct that contains the history of a pod.
type PodHistory struct {
	Pod   v1.Pod
	Event v1.Event
}

// NodeHistory is a struct that contains the history of a node.
type NodeHistory struct {
	Node  v1.Node
	Event v1.Event
}

// ClusterSnapshot is a struct that contains the snapshot of a cluster.
type ClusterSnapshot struct {
	ClusterName       string
	Nodes             []v1.Node
	Pods              []v1.Pod
	snapshotTimestamp metav1.Time
}

// OpenPodHistoryDB opens the podhistorydb database.
func OpenPodHistoryDB() (*PodHistoryDB, error) {
	var db *bolt.DB
	db, err := bolt.Open("ktm_podhistorydb.db", 0600, nil)
	if err != nil {
		return nil, err
	}

	return &PodHistoryDB{
		db: db,
	}, nil

}

// OpenPodHistoryDBWithFile opens the podhistorydb database with a given file.
func OpenPodHistoryDBWithFile(filename string) (*PodHistoryDB, error) {
	var db *bolt.DB
	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &PodHistoryDB{
		db: db,
	}, nil

}

// CheckPodHistoryDB checks if the podhistorydb database exists.
func (phdb *PodHistoryDB) CheckPodHistoryDB() error {
	return phdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("podhistory"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}
		return nil
	})
}

// NewPodHistoryDB creates a new PodHistoryDB instance.
func NewPodHistoryDB(db *bolt.DB) *PodHistoryDB {
	return &PodHistoryDB{
		db: db,
	}
}

// Close closes the database.
func (phdb *PodHistoryDB) Close() error {
	return phdb.db.Close()
}

// getTimestampSuffix calculates the timestamp suffix for object history after #
func getTimestampSuffix(object interface{}) string {
	// check if the object is podhistory or nodehistory
	if ph, ok := (object).(*PodHistory); ok {
		// use podname#event timestamp as the key, use creation timestamp if event timestamp is not available
		if ph.Event.FirstTimestamp.IsZero() {
			return ph.Pod.ObjectMeta.CreationTimestamp.Format(time.RFC3339Nano)
		} else {
			return ph.Event.FirstTimestamp.Format(time.RFC3339Nano)
		}
	}
	if nh, ok := (object).(*NodeHistory); ok {
		// use nodename#event timestamp as the key, use creation timestamp if event timestamp is not available
		if nh.Event.FirstTimestamp.IsZero() {
			fmt.Printf("Object Creation timestamp for nodehistory=%s\n", nh.Node.ObjectMeta.CreationTimestamp.Format(time.RFC3339Nano))
			return nh.Node.ObjectMeta.CreationTimestamp.Format(time.RFC3339Nano)
		} else {
			// fmt.Printf("Event timestamp for nodehistory=%s\n", nh.Node.ObjectMeta.CreationTimestamp.Format(time.RFC3339Nano))
			return nh.Event.FirstTimestamp.Format(time.RFC3339Nano)
		}
	}

	// print the object type
	fmt.Printf("Object type=%T\n", object)
	// TODO: better return value than ""?
	return ""
}
