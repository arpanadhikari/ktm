package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

// AddClusterSnapshot adds a cluster snapshot to the database.
func (phdb *PodHistoryDB) AddClusterSnapshot(cs ClusterSnapshot) error {

	// Print the cluster snapshot
	fmt.Println("Cluster snapshot:")
	fmt.Println("Nodes Count", len(cs.Nodes))
	fmt.Println("Pods Count", len(cs.Pods))

	return phdb.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("clustersnapshots"))
		if err != nil {
			return err
		}
		// Create  a sub-bucket for the cluster snapshot
		clusterBucket, err := b.CreateBucketIfNotExists([]byte(cs.ClusterName))
		if err != nil {
			return err
		}
		// Add the cluster snapshot to the sub-bucket with timestamp as key
		serialized, _ := json.Marshal(cs)
		key := cs.snapshotTimestamp.Format(time.RFC3339Nano)
		fmt.Println("New Cluster Snapshot: ", key)
		return clusterBucket.Put([]byte(key), serialized)
	})
}

// GetClusterSnapshot returns the last cluster snapshot before the given relative timeframe.
func (phdb *PodHistoryDB) GetClusterSnapshot(clusterName string, relativeTime string) (ClusterSnapshot, error) {
	var cs ClusterSnapshot
	err := phdb.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("clustersnapshots"))
		if b == nil {
			return bolt.ErrBucketNotFound
		}
		// Get the sub-bucket for the cluster
		clusterBucket := b.Bucket([]byte(clusterName))
		if clusterBucket == nil {
			return bolt.ErrBucketNotFound
		}
		// Use relativeTime to calculate last possible timestamp
		duration, err := time.ParseDuration(relativeTime)
		if err != nil {
			return err
		}

		earliestTime := time.Now().Add(-duration)
		fmt.Println("earliestTime", earliestTime)
		// Iterate over the cluster snapshot keys in reverse order
		c := clusterBucket.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			// Parse the key as a timestamp
			timestamp, err := time.Parse(time.RFC3339Nano, string(k))
			if err != nil {
				return err
			}

			// print timestamp
			fmt.Println("timestamp", timestamp)
			// fmt.Println("earliestTime", earliestTime)
			// return the first cluster snapshot found after the earliestTime
			if timestamp.After(earliestTime) {
				err = json.Unmarshal(v, &cs)
				if err != nil {
					return err
				}
				return nil
			}
		}
		return nil
	})
	return cs, err
}

// GetReconciledClusterSnapshot returns the last cluster snapshot
// before the given relative timeframe and infers the current state of the cluster through the pod history.
func (phdb *PodHistoryDB) GetReconciledClusterSnapshot(clusterName string, relativeTime string) (ClusterSnapshot, error) {
	var cs ClusterSnapshot
	var err error
	// Get the last cluster snapshot before the given relative timeframe
	cs, err = phdb.GetClusterSnapshot(clusterName, relativeTime)
	if err != nil {
		return cs, err
	}
	fmt.Println("retrieved Cluster snapshot")

	fmt.Println("Nodes Count", len(cs.Nodes))
	fmt.Println("Pods Count", len(cs.Pods))

	fmt.Println("Snapshot Timestamp: ", cs.snapshotTimestamp.Time.Format(time.RFC3339Nano))

	// Get the pod history for the last cluster snapshot
	ph, err := phdb.GetPodsRelativeTime(relativeTime)
	if err != nil {
		return cs, err
	}

	// Reconcile the cluster snapshot with the pod history in reverse order

	for i := len(ph) - 1; i >= 0; i-- {
		pod := ph[i]
		if pod.Event.Type == "Added" {
			// if pod already exist just replace it
			found := false
			for i, p := range cs.Pods {
				if p.ObjectMeta.Name == pod.Pod.ObjectMeta.Name {
					found = true
					cs.Pods[i] = pod.Pod
				}
			}
			if !found {
				cs.Pods = append(cs.Pods, pod.Pod)
			}
		}
		if pod.Event.Type == "Deleted" {
			for i, p := range cs.Pods {
				if p.ObjectMeta.Name == pod.Pod.ObjectMeta.Name {
					cs.Pods = append(cs.Pods[:i], cs.Pods[i+1:]...)
				}
			}
		}
		if pod.Event.Type == "Updated" {
			for i, p := range cs.Pods {
				if p.ObjectMeta.Name == pod.Pod.ObjectMeta.Name {
					cs.Pods[i] = pod.Pod
				}
			}
		}
	}

	fmt.Println("reconciled Cluster snapshot")
	fmt.Println("Nodes Count", len(cs.Nodes))
	fmt.Println("Pods Count", len(cs.Pods))

	return cs, nil
}
