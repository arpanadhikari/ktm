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
		// timestamp suffix for the podhistory object key
		var timestamp string = getTimestampSuffix(&ph)
		key := ph.Pod.ObjectMeta.Name + "#" + timestamp
		// print the key
		fmt.Printf("adding key=%s\n", key)
		return b.Put([]byte(key), serialized)
	})
}

func (phdb *PodHistoryDB) GetPodsRelativeTime(relativeTime string) ([]PodHistory, error) {
	var ph []PodHistory
	seenPodNames := make(map[string]bool)

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

		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			// Find the separator position in the key
			sepPos := bytes.IndexByte(k, '#')
			if sepPos == -1 {
				continue
			}

			// Extract the timestamp from the key
			timestampBytes := k[sepPos+1:]

			// parse the key before #
			podName := string(k[:sepPos])
			if seenPodNames[podName] {
				continue
			}

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
				// Mark the pod-name as seen
				seenPodNames[podName] = true

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
		// timestamp suffix for the nodehistory object key
		var timestamp string = getTimestampSuffix(&nh)
		key := nh.Node.ObjectMeta.Name + "#" + timestamp
		fmt.Printf("adding key=%s\n", key)
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

// GetNodesRelativeTime returns the node history for a given node.
func (phdb *PodHistoryDB) GetNodesRelativeTime(relativeTime string) ([]NodeHistory, error) {
	var nh []NodeHistory
	seenNodeNames := make(map[string]bool)

	// parse the relative time string into a duration
	duration, err := time.ParseDuration(relativeTime)
	if err != nil {
		return nil, err
	}

	// calculate the earliest timestamp
	earliestTime := time.Now().Add(-duration)

	err = phdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		if b == nil {
			return nil
		}
		c := b.Cursor()

		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			// find the separator position in the key
			sepPos := bytes.IndexByte(k, '#')
			if sepPos == -1 {
				continue
			}

			// extract the timestamp from the key
			timestampBytes := k[sepPos+1:]

			// parse the key before #
			nodeName := string(k[:sepPos])
			if seenNodeNames[nodeName] {
				continue
			}

			// parse the timestamp string into a time.Time value
			timestampStr := string(timestampBytes)
			timestamp, err := time.Parse(time.RFC3339, timestampStr)
			if err != nil {
				return err
			}

			// check if the timestamp of the event is after earliestTime
			if timestamp.After(earliestTime) {
				var n NodeHistory
				json.Unmarshal(v, &n)
				nh = append(nh, n)
				// mark the node-name as seen
				seenNodeNames[nodeName] = true
			}
		}
		return nil
	})

	return nh, err
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
			fmt.Printf("Event timestamp for nodehistory=%s\n", nh.Node.ObjectMeta.CreationTimestamp.Format(time.RFC3339Nano))
			return nh.Event.FirstTimestamp.Format(time.RFC3339Nano)
		}
	}

	// print the object type
	fmt.Printf("Object type=%T\n", object)
	// TODO: better return value than ""?
	return ""
}

// Close closes the database.
func (phdb *PodHistoryDB) Close() error {
	return phdb.db.Close()
}
