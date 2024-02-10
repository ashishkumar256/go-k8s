package main

import (
	"context"
	"fmt"
	"log"

        "flag"
	"path/filepath"
	"k8s.io/client-go/util/homedir"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig string
)

func init() {
	home := homedir.HomeDir()
	flag.StringVar(&kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to the kubeconfig file")
	flag.Parse()
}

func main() {
	// Path to kubeconfig file.
	//kubeconfig := "/Users/ashishsingh/.kube/config"

	// Load kubeconfig file.
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
       
	if err != nil {
		fmt.Printf(err.Error())
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatal(err)
		}
	}

	// Create Kubernetes clientset.
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	// Get list of pods.
	pods, err := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err)
	}

	// Print pod names.
	fmt.Println("Pods:")
	for _, pod := range pods.Items {
		fmt.Printf(" - %s\n", pod.GetName())
	}
}
