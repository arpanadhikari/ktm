/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watches Kubernetes events for pods and nodes",
	RunE: func(cmd *cobra.Command, args []string) error {

		//Load Kubernetes configuration from file or environment variable
		config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
		if err != nil {
			return fmt.Errorf("failed to load Kubernetes config: %w", err)
		}

		db, _ := OpenPodHistoryDB()
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return fmt.Errorf("failed to create Kubernetes clientset: %w", err)
		}

		// run snapshotCluster as a goroutine every 30 seconds
		go func() {
			for {
				if err := snapshotCluster(clientset, db); err != nil {
					fmt.Printf("failed to watch pods: %v", err)
				}
				time.Sleep(30 * time.Second)
			}
		}()

		stop := make(chan struct{})

		var wg sync.WaitGroup
		wg.Add(2)

		// run watchEvents as a goroutine
		go func() {
			defer wg.Done()
			if err := watchEvents(clientset, db, stop); err != nil {
				fmt.Printf("failed to watch events: %v", err)
			}
		}()

		// start web server as a goroutine
		fmt.Println("Starting web server...")
		go func() {
			defer wg.Done()
			StartWebServer(db, stop)
		}()

		// wait for both goroutines to complete
		wg.Wait()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func snapshotCluster(clientset kubernetes.Interface, db *PodHistoryDB) error {

	//print startup message
	fmt.Println("Taking cluster pod/node snapshot...")

	pods, err := clientset.CoreV1().Pods(v1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	// TODO: use a proper cluster unique identifier, currently hardcoded for testing
	// print cluster host
	// fmt.Println("Cluster host: ", clientset.CoreV1().RESTClient().Get().URL().Host)
	clusterName := "127.0.0.1:6443"

	db.AddClusterSnapshot(ClusterSnapshot{
		// get clustername via clientset
		ClusterName: clusterName,
		Nodes:       nodes.Items,
		Pods:        pods.Items,
		snapshotTimestamp: metav1.Time{
			Time: time.Now(),
		},
	},
	)

	return nil

}

// watchEvents watches for pod and node events and writes them to the database
func watchEvents(clientset kubernetes.Interface, db *PodHistoryDB, stop chan struct{}) error {

	informerFactory := informers.NewSharedInformerFactory(clientset, time.Second*0)

	podInformer := informerFactory.Core().V1().Pods().Informer()
	nodeInformer := informerFactory.Core().V1().Nodes().Informer()

	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			onAdd(obj, db)
		},
		DeleteFunc: func(obj interface{}) {
			onDelete(obj, db)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			onUpdate(oldObj, newObj, db)
		},
	})

	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			onAdd(obj, db)
		},
		DeleteFunc: func(obj interface{}) {
			onDelete(obj, db)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			onUpdate(oldObj, newObj, db)
		},
	})

	// check if informers are valid
	if nodeInformer == nil {
		return fmt.Errorf("failed to create node controller")
	}
	if podInformer == nil {
		return fmt.Errorf("failed to create pod controller")
	}

	fmt.Println("Starting controllers...")
	start := time.Now()
	fmt.Println("watchEvents: waiting for stop signal...")

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		podInformer.Run(stop)
	}()
	go func() {
		defer wg.Done()
		nodeInformer.Run(stop)
	}()
	wg.Wait()

	// wait for stop signal or informers to finish
	<-stop

	end := time.Now()
	fmt.Printf("Controllers stopped after %v", end.Sub(start))

	return nil

}

func onAdd(obj interface{}, db *PodHistoryDB) {
	if pod, ok := obj.(*v1.Pod); ok {
		fmt.Println("Pod add event added to store ")
		// write podhistory to database
		if err := db.AddPodHistory(PodHistory{
			Pod: *pod,
			Event: v1.Event{
				Type: "Added",
				FirstTimestamp: metav1.Time{
					Time: time.Now(),
				},
			},
		}); err != nil {
			fmt.Errorf("failed to write pod history to database: %w", err)
		}
	}
	if node, ok := obj.(*v1.Node); ok {
		fmt.Println("Node add event added to store ")
		if err := db.AddNodeHistory(NodeHistory{
			Node: *node,
			Event: v1.Event{
				Type: "Added",
				FirstTimestamp: metav1.Time{
					Time: time.Now(),
				},
			},
		}); err != nil {
			fmt.Errorf("failed to write node history to database: %w", err)
		}
	}
}

func onDelete(obj interface{}, db *PodHistoryDB) {
	if pod, ok := obj.(*v1.Pod); ok {
		fmt.Println("Pod delete event added to store ")
		// write podhistory to database
		if err := db.AddPodHistory(PodHistory{
			Pod: *pod,
			Event: v1.Event{
				Type: "Deleted",
				FirstTimestamp: metav1.Time{
					Time: time.Now(),
				},
			},
		}); err != nil {
			fmt.Errorf("failed to write pod history to database: %w", err)
		}
	}
	if node, ok := obj.(*v1.Node); ok {
		fmt.Println("Node delete event added to store ")
		if err := db.AddNodeHistory(NodeHistory{
			Node: *node,
			Event: v1.Event{
				Type: "Deleted",
				FirstTimestamp: metav1.Time{
					Time: time.Now(),
				},
			},
		}); err != nil {
			fmt.Errorf("failed to write node history to database: %w", err)
		}
	}
}

func onUpdate(oldObj, newObj interface{}, db *PodHistoryDB) {

	// fmt.Printf("Diff: %v", cmp.Diff(oldObj, newObj))
	if pod, ok := newObj.(*v1.Pod); ok {
		fmt.Printf("Pod update event added to store: %s\n", pod.GetName())
		// write podhistory to database
		if err := db.AddPodHistory(PodHistory{
			Pod: *pod,
			Event: v1.Event{
				Type: "Updated",
				FirstTimestamp: metav1.Time{
					Time: time.Now(),
				},
			},
		}); err != nil {
			fmt.Errorf("failed to write pod history to database: %w", err)
		}
	}
	if node, ok := newObj.(*v1.Node); ok {
		fmt.Printf("Node update event added to store: %s\n", node.GetName())
		if err := db.AddNodeHistory(NodeHistory{
			Node: *node,
			Event: v1.Event{
				Type: "Updated",
				FirstTimestamp: metav1.Time{
					Time: time.Now(),
				},
			},
		}); err != nil {
			fmt.Errorf("failed to write node history to database: %w", err)
		}
	}
}
