package cmd

// This package is a wrapper around the BoltDB key/value store. It provides
// a simple interface for storing and retrieving pod history data.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"

	bolt "go.etcd.io/bbolt"
)

// PodHistoryDB is a wrapper around the BoltDB key/value store.
type PodHistoryDB struct {
	db *bolt.DB
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

// AddPodHistory adds a pod history to the database.
func (phdb *PodHistoryDB) AddPodHistory(ph PodHistory) error {
	return phdb.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("pods"))
		if err != nil {
			return err
		}
		// Serialize the pod history
		serialized, _ := json.Marshal(ph)
		// Use podname#event timestamp as the key, use creation timestamp if event timestamp is not available
		var timestamp string
		if ph.Event.FirstTimestamp.IsZero() {
			timestamp = ph.Pod.ObjectMeta.CreationTimestamp.Format(time.RFC3339Nano)
		} else {
			timestamp = ph.Event.FirstTimestamp.Format(time.RFC3339Nano)
		}
		key := ph.Pod.ObjectMeta.Name + "#" + timestamp
		// print the key
		fmt.Printf("adding key=%s\n", key)
		return b.Put([]byte(key), serialized)
	})
}

func (phdb *PodHistoryDB) GetPodsRelativeTime(relativeTime string) ([]PodHistory, error) {
	var ph []PodHistory

	// Parse the relative time string into a duration
	duration, err := time.ParseDuration(relativeTime)
	if err != nil {
		return nil, err
	}

	// Calculate the earliest timestamp
	earliestTime := time.Now().Add(-duration)

	err = phdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("pods"))
		if b == nil {
			return nil
		}
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			// Find the separator position in the key
			sepPos := bytes.IndexByte(k, '#')
			if sepPos == -1 {
				continue
			}

			// Extract the timestamp from the key
			timestampBytes := k[sepPos+1:]

			// Parse the timestamp string into a time.Time value
			timestampStr := string(timestampBytes)
			timestamp, err := time.Parse(time.RFC3339, timestampStr)
			if err != nil {
				return err
			}

			// Check if the timestamp of the event is after earliestTime
			if timestamp.After(earliestTime) {
				var p PodHistory
				json.Unmarshal(v, &p)
				ph = append(ph, p)
			}
		}
		return nil
	})

	return ph, err
}

// GetPodHistory returns the pod history for a given pod.
func (phdb *PodHistoryDB) GetPodHistory(podName string) ([]PodHistory, error) {
	var ph []PodHistory

	// Retrieve all the pod history that starts with the pod name
	err := phdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("pods"))
		if b == nil {
			return nil
		}
		c := b.Cursor()

		prefix := []byte(podName + "#")
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var p PodHistory
			json.Unmarshal(v, &p)
			ph = append(ph, p)
			// print the key and value
			fmt.Printf("key=%s, value=%s\n", k, v)
		}
		return nil
	})

	return ph, err
}

// AddNodeHistory add a nodes history to the database.
func (phdb *PodHistoryDB) AddNodeHistory(nh NodeHistory) error {
	return phdb.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("nodes"))
		if err != nil {
			return err
		}
		// serialize the node history
		serialized, _ := json.Marshal(nh)
		// use nodename#event timestamp as the key, use creation timestamp if event timestamp is not available
		var timestamp string
		if nh.Event.FirstTimestamp.IsZero() {
			timestamp = nh.Node.ObjectMeta.CreationTimestamp.Format(time.RFC3339Nano)
		} else {
			timestamp = nh.Event.FirstTimestamp.Format(time.RFC3339Nano)
		}
		key := nh.Node.ObjectMeta.Name + "#" + timestamp
		return b.Put([]byte(key), serialized)
	})
}

// GetNodeHistory returns the node history for a given node.
func (phdb *PodHistoryDB) GetNodeHistory(nodeName string) ([]NodeHistory, error) {
	// retrieve all the node history that starts with the node name
	var nh []NodeHistory
	err := phdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		if b == nil {
			return nil
		}
		c := b.Cursor()

		prefix := []byte(nodeName)
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			var n NodeHistory
			json.Unmarshal(v, &n)
			nh = append(nh, n)
		}
		return nil
	})
	return nh, err
}

// Close closes the database.
func (phdb *PodHistoryDB) Close() error {
	return phdb.db.Close()
}
