package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

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
func (phdb *PodHistoryDB) GetNodeHistory(nodeName string, relativeTime string) ([]NodeHistory, error) {
	// retrieve all the node history that starts with the node name
	var nh []NodeHistory

	// Parse the relative time string into a duration
	duration, err := time.ParseDuration(relativeTime)
	if err != nil {
		return nil, err
	}

	// Calculate the earliest timestamp
	earliestTime := time.Now().Add(-duration)

	err = phdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("nodes"))
		if b == nil {
			return nil
		}
		c := b.Cursor()

		prefix := []byte(nodeName + "#")

		// TODO: Search from the end of the bucket if relative time is specified
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			// parse the key as a timestamp
			timestamp, err := time.Parse(time.RFC3339, string(k[len(prefix):]))
			if err != nil {
				return err
			}
			if timestamp.After(earliestTime) {
				var n NodeHistory
				json.Unmarshal(v, &n)
				nh = append(nh, n)
			}
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
