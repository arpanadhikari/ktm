package cmd

// This package is a wrapper around the BoltDB key/value store. It provides
// a simple interface for storing and retrieving pod history data.

import (
	"time"

	bolt "go.etcd.io/bbolt"
)

// PodHistoryDB is a wrapper around the BoltDB key/value store.
type PodHistoryDB struct {
	db *bolt.DB
}

// OpenPodHistoryDB opens the podhistorydb database.
func OpenPodHistoryDB() (*PodHistoryDB, error) {
	db, err := bolt.Open("podhistorydb", 0600, nil)

	if err != nil {
		return nil, err
	}
	return NewPodHistoryDB(db), nil
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
	PodName      string
	PodNamespace string
	NodeName     string
	PodUID       string
	StartTime    time.Time
	EndTime      time.Time
	Resources    struct {
		CPU    string
		Memory string
	}
}

// NodeHistory is a struct that contains the history of a node.
type NodeHistory struct {
	NodeName  string
	StartTime time.Time
	EndTime   time.Time
	Resources struct {
		CPU    string
		Memory string
	}
}

// AddPodHistory adds a pod history to the database.
func (phdb *PodHistoryDB) AddPodHistory(ph PodHistory) error {
	return phdb.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(ph.PodNamespace))
		if err != nil {
			return err
		}
		return b.Put([]byte(ph.PodName), []byte(ph.PodUID))
	})
}

// GetPodHistory returns the pod history for a given pod.
func (phdb *PodHistoryDB) GetPodHistory(podName, podNamespace string) (PodHistory, error) {
	var ph PodHistory
	err := phdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(podNamespace))
		if b == nil {
			return nil
		}
		ph.PodUID = string(b.Get([]byte(podName)))
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
		return b.Put([]byte(nh.NodeName), []byte(nh.NodeName))
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
		nh.NodeName = string(b.Get([]byte(nodeName)))
		return nil
	})
	return nh, err
}

// Close closes the database.
func (phdb *PodHistoryDB) Close() error {
	return phdb.db.Close()
}
