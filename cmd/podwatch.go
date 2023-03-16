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
	Use:   "podwatch",
	Short: "Watches Kubernetes events",
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

		// call podWatch function
		if err := podWatch(clientset, db); err != nil {
			return fmt.Errorf("failed to watch pods: %w", err)
		}

		stop := make(chan struct{})
		defer close(stop)

		//watch for pod and node events
		if err := watchEvents(clientset, db, stop); err != nil {
			return fmt.Errorf("failed to watch events: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func podWatch(clientset kubernetes.Interface, db *PodHistoryDB) error {

	//print startup message
	fmt.Println("Starting podwatch...")

	pods, err := clientset.CoreV1().Pods(v1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	for _, node := range nodes.Items {
		// add node history to database
		onAdd(&node, db)
	}
	for _, pod := range pods.Items {
		// add pod history to database
		onAdd(&pod, db)
	}

	return nil

}

// TODO: refactor the function to enable unit testing
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
		fmt.Printf("New Pod Added to Store: %s\n", pod.GetName())
		pod := obj.(*v1.Pod)
		// write podhistory to database
		if err := db.AddPodHistory(PodHistory{
			Pod: *pod,
		}); err != nil {
			fmt.Errorf("failed to write pod history to database: %w", err)
		}
	}
	if node, ok := obj.(*v1.Node); ok {
		fmt.Printf("New Node Added to Store: %s\n", node.GetName())
		if err := db.AddNodeHistory(NodeHistory{
			Node: *node,
		}); err != nil {
			fmt.Errorf("failed to write node history to database: %w", err)
		}
	}
}

func onDelete(obj interface{}, db *PodHistoryDB) {
	if pod, ok := obj.(*v1.Pod); ok {
		fmt.Printf("Pod Deleted from Store: %s\n", pod.GetName())
	}
	if node, ok := obj.(*v1.Node); ok {
		fmt.Printf("Node Deleted from Store: %s\n", node.GetName())
	}
}

func onUpdate(oldObj, newObj interface{}, db *PodHistoryDB) {
	if pod, ok := newObj.(*v1.Pod); ok {
		fmt.Printf("Pod Updated in Store: %s\n", pod.GetName())
	}
	if node, ok := newObj.(*v1.Node); ok {
		fmt.Printf("Node Updated in Store: %s\n", node.GetName())
	}

	// print the old object name
	if oldObj, ok := oldObj.(*v1.Node); ok {
		fmt.Printf("oldObj: %s", oldObj.GetName())
	}
	if oldObj, ok := oldObj.(*v1.Pod); ok {
		fmt.Printf("oldObj: %s", oldObj.GetName())
	}
}
