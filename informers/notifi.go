package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/informers"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
)

const (
	podDeletedEventType = "DELETED"
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
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Error creating clientset: %s", err.Error())
	}

	// Create a new shared informer factory
	sharedInformerFactory := informers.NewSharedInformerFactory(clientset, time.Second*30)

	// Get the pod informer from the shared informer factory
	podInformer := sharedInformerFactory.Core().V1().Pods().Informer()

	// Initialize the work queue
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "pods")

	// Handle events from the informer and put them into the work queue
	go handleEvents(podInformer, queue)

	// Start a worker to process items from the work queue
	go processItems(queue)

	// Start all informers
	sharedInformerFactory.Start(wait.NeverStop)

	// Wait forever
	select {}
}

func handleEvents(informer cache.SharedIndexInformer, queue workqueue.RateLimitingInterface) {
	// Add event handlers to the informer
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: func(obj interface{}) {
			// Assert the type of the deleted object
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err != nil {
				klog.Errorf("Error getting key: %v", err)
				return
			}
			queue.Add(key)
		},
	})

	// Run the informer until a stop signal is received
	stopCh := make(chan struct{})
	defer close(stopCh)
	go func() {
		defer runtime.HandleCrash()
		informer.Run(stopCh)
	}()

	// Wait until the informer is synced
	if !cache.WaitForCacheSync(stopCh, informer.HasSynced) {
		runtime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}
	klog.Info("Cache is synced")

	// Wait forever
	<-stopCh
}

func processItems(queue workqueue.RateLimitingInterface) {
	for {
		// Wait for an item from the queue
		key, quit := queue.Get()
		if quit {
			return
		}

		// Handle the item
		err := handleItem(key.(string))
		if err != nil {
			queue.AddRateLimited(key)
			klog.Errorf("Error handling item %q: %v", key, err)
		} else {
			queue.Forget(key)
		}

		// Mark the item as done
		queue.Done(key)
	}
}

func handleItem(key string) error {
	fmt.Printf("Pod %s deleted\n", key)
	return nil
}

