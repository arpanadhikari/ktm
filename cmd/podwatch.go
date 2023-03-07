/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
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
		return podWatch()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func podWatch() error {
	// Load Kubernetes configuration from file or environment variable
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		return fmt.Errorf("failed to load Kubernetes config: %w", err)
	}

	// Create a Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	//create a Kubernetes client while running in cluster
	// config, err := rest.InClusterConfig()
	// if err != nil {
	// 	return fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	// }

	// get a list of pods running in the cluster

	// initialize a new bolt database
	db, err := OpenPodHistoryDB()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	pods, err := clientset.CoreV1().Pods(v1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	// get a list of nodes with their cpu and memory capacity
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	for _, node := range nodes.Items {

		// fmt.Printf("%s\t%s\t%s", node.Name, node.Status.Allocatable.Cpu(), node.Status.Allocatable.Memory())

		// add node history to database
		if err := db.AddNodeHistory(NodeHistory{
			NodeName:  node.Name,
			StartTime: node.CreationTimestamp.Time,
			// EndTime: node.DeletionTimestamp.Time,
			Resources: struct {
				CPU    string
				Memory string
			}{
				CPU:    node.Status.Allocatable.Cpu().String(),
				Memory: node.Status.Allocatable.Memory().String(),
			},
		}); err != nil {
			return fmt.Errorf("failed to write node history to database: %w", err)
		}

	}

	// print list of pod names and their hosts
	for _, pod := range pods.Items {

		// add pod history to database
		if err := db.AddPodHistory(PodHistory{
			PodName:      pod.Name,
			PodNamespace: pod.Namespace,
			NodeName:     pod.Spec.NodeName,
			StartTime:    pod.CreationTimestamp.Time,
			// EndTime:      pod.DeletionTimestamp.Time,
			Resources: struct {
				CPU    string
				Memory string
			}{
				CPU:    pod.Spec.Containers[0].Resources.Requests.Cpu().String(),
				Memory: pod.Spec.Containers[0].Resources.Requests.Memory().String(),
			},
		}); err != nil {
			return fmt.Errorf("failed to write pod history to database: %w", err)
		}

	}

	// return nil

	// watch for new pod events
	podWatchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", v1.NamespaceAll, fields.Everything())
	_, controller_pod := cache.NewInformer(
		podWatchlist,
		&v1.Pod{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				pod := obj.(*v1.Pod)
				fmt.Printf("New Pod Added to Store: %s", pod.GetName())
				// write podhistory to database
				if err := db.AddPodHistory(PodHistory{
					PodName:      pod.Name,
					PodNamespace: pod.Namespace,
					NodeName:     pod.Spec.NodeName,
					StartTime:    pod.CreationTimestamp.Time,
					// EndTime:      pod.DeletionTimestamp.Time,
					Resources: struct {
						CPU    string
						Memory string
					}{
						CPU:    pod.Spec.Containers[0].Resources.Requests.Cpu().String(),
						Memory: pod.Spec.Containers[0].Resources.Requests.Memory().String(),
					},
				}); err != nil {
					fmt.Errorf("failed to write pod history to database: %w", err)
				}
			},
			DeleteFunc: func(obj interface{}) {
				pod := obj.(*v1.Pod)
				fmt.Printf("Pod Deleted from Store: %s", pod.GetName())
			},
		},
	)

	// watch for node events
	nodeWatchlist := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "nodes", v1.NamespaceAll, fields.Everything())
	_, controller_node := cache.NewInformer(
		nodeWatchlist,
		&v1.Node{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				node := obj.(*v1.Node)
				fmt.Printf("New Node Added to Store: %s", node.GetName())
				// write nodehistory to database
				if err := db.AddNodeHistory(NodeHistory{
					NodeName:  node.Name,
					StartTime: node.CreationTimestamp.Time,
					// EndTime:   node.DeletionTimestamp.Time,
					Resources: struct {
						CPU    string
						Memory string
					}{
						CPU:    node.Status.Allocatable.Cpu().String(),
						Memory: node.Status.Allocatable.Memory().String(),
					},
				}); err != nil {
					fmt.Errorf("failed to write node history to database: %w", err)
				}
			},
			DeleteFunc: func(obj interface{}) {
				node := obj.(*v1.Node)
				fmt.Printf("Node Deleted from Store: %s", node.GetName())
			},
		},
	)

	// multiple controllers can be run in parallel
	stop := make(chan struct{})
	defer close(stop)

	// Create channel to receive termination signals
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	// run the controllers
	go controller_pod.Run(stop)
	go controller_node.Run(stop)

	select {
	case <-term:
		fmt.Println("Received SIGTERM, exiting gracefully...")
		db.Close()
	case <-stop:
		fmt.Println("Received stop signal, exiting gracefully...")
		db.Close()
	}

	// Wait for controllers to finish
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		controller_pod.Run(stop)
		wg.Done()
	}()
	go func() {
		controller_node.Run(stop)
		wg.Done()
	}()
	wg.Wait()

	return nil

}
