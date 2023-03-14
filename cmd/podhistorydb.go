package cmd

// This package is a wrapper around the BoltDB key/value store. It provides
// a simple interface for storing and retrieving pod history data.

import (
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"

	bolt "go.etcd.io/bbolt"
)

// PodHistoryDB is a wrapper around the BoltDB key/value store.
type PodHistoryDB struct {
	db *bolt.DB
}

// OpenPodHistoryDB opens the podhistorydb database.
func OpenPodHistoryDB() (*PodHistoryDB, error) {
	db, err := bolt.Open("ktm_podhistorydb.db", 0600, nil)
	if err != nil {

		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// return an instance of PodHistoryDB
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
	Pod v1.Pod
}

// NodeHistory is a struct that contains the history of a node.
type NodeHistory struct {
	Node v1.Node
}

// AddPodHistory adds a pod history to the database.
func (phdb *PodHistoryDB) AddPodHistory(ph PodHistory) error {
	return phdb.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("pods"))
		if err != nil {
			return err
		}
		// searialize the pod history
		serialized, _ := json.Marshal(ph)

		return b.Put([]byte(ph.Pod.ObjectMeta.Name), []byte(serialized))
	})

}

// GetPodHistory returns the pod history for a given pod.
func (phdb *PodHistoryDB) GetPodHistory(podName string) (PodHistory, error) {
	var ph PodHistory

	// retrieve the pod history struct from the database
	err := phdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("pods"))
		if b == nil {
			return nil
		}
		//desearlize the pod history
		serialized := b.Get([]byte(podName))
		json.Unmarshal(serialized, &ph)
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
		serialized, err := json.Marshal(nh)
		if err != nil {
			return err
		}
		return b.Put([]byte(nh.Node.ObjectMeta.Name), serialized)
	})
}

// GetNodeHistory returns the node history for a given node.
func (phdb *PodHistoryDB) GetNodeHistory(nodeName string) (NodeHistory, error) {
	var nh NodeHistory
	err := phdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		if b == nil {
			return nil
		}
		// desearlize the node history
		serialized := b.Get([]byte(nodeName))
		json.Unmarshal(serialized, &nh)
		return nil
	})
	return nh, err
}

// Close closes the database.
func (phdb *PodHistoryDB) Close() error {
	return phdb.db.Close()
}
