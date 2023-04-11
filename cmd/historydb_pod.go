package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	bolt "go.etcd.io/bbolt"
)

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

// GetPodHistoryTimestamp returns the pod history for the given pod name after the given timestamp.
func (phdb *PodHistoryDB) GetPodHistoryTimestamp(timestamp string, podName string) (PodHistory, error) {
	var ph PodHistory
	err := phdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("pods"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}
		// Iterate over the pod history keys in reverse order
		c := b.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			// Find the separator position in the key
			sep := strings.Index(string(k), "#")
			// Parse the key as a timestamp
			t, err := time.Parse(time.RFC3339Nano, string(k[sep+1:]))
			if err != nil {
				return err
			}
			// return the first pod history found after the timestamp
			if t.After(time.Time{}) {
				err = json.Unmarshal(v, &ph)
				if err != nil {
					return err
				}
				return nil
			}
		}
		return nil
	})
	return ph, err
}

// GetPodsRelativeTime returns the pod history before the given relative timeframe.
func (phdb *PodHistoryDB) GetPodsRelativeTime(relativeTime string) ([]PodHistory, error) {
	var ph []PodHistory
	// seenPodNames := make(map[string]bool)

	// Parse the relative time string into a duration
	duration, err := time.ParseDuration(relativeTime)
	if err != nil {
		return nil, err
	}

	// Calculate the earliest timestamp if relativeTime is a relative duration
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
			// podName := string(k[:sepPos])
			// if seenPodNames[podName] {
			// 	continue
			// }

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
				// seenPodNames[podName] = true

			}
		}
		return nil
	})

	return ph, err
}

// GetPodHistory returns the pod history for a given pod during a given relative time.
func (phdb *PodHistoryDB) GetPodHistory(podName string, relativeTime string) ([]PodHistory, error) {
	var ph []PodHistory

	// Parse the relative time string into a duration
	duration, err := time.ParseDuration(relativeTime)
	if err != nil {
		return nil, err
	}

	// Calculate the earliest timestamp
	earliestTime := time.Now().Add(-duration)

	// Retrieve all the pod history that starts with the pod name and are after the earliestTime
	err = phdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("pods"))
		if b == nil {
			return nil
		}
		c := b.Cursor()

		prefix := []byte(podName + "#")

		// TODO: Search from the end of the bucket if relative time is specified
		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			// Parse the key as a timestamp
			timestamp, err := time.Parse(time.RFC3339, string(k[len(prefix):]))
			if err != nil {
				return err
			}
			if timestamp.After(earliestTime) {
				var p PodHistory
				json.Unmarshal(v, &p)
				ph = append(ph, p)
				// print the key and value
				fmt.Printf("key=%s", k)
			}
		}
		return nil
	})

	return ph, err
}
