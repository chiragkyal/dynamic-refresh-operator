package controller

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"

	secret "github.com/chiragkyal/dynamic-refresh-operator/pkg/controller/monitor"
	"k8s.io/apimachinery/pkg/fields"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	queue    workqueue.RateLimitingInterface
	informer cache.Controller
}

// NewController creates a new Controller.
func NewController(clientset *kubernetes.Clientset, namespace string, queue workqueue.RateLimitingInterface) *Controller {
	secretManager := secret.NewManager(clientset, queue)
	lw := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "configmaps", namespace, fields.Everything())
	_, informer := cache.NewIndexerInformer(lw, &v1.ConfigMap{}, 0, getHandler(queue, secretManager), cache.Indexers{})

	return &Controller{
		informer: informer,
		queue:    queue,
	}
}

func (c *Controller) processNextItem() bool {
	// Wait until there is a new item in the working queue
	item, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(item)

	if _, ok := item.(Handler); ok {
		klog.Infof("successfully got item %v", item)
	} else {
		klog.Infof("did not got item %v", item)
	}

	// Invoke business logic
	err := c.logic(item.(Handler))
	klog.Error(err)

	// TODO:Handle error
	return err == nil
}

func (c *Controller) logic(handler Handler) error {
	return handler.Handle()
}

// Run begins watching and syncing.
func (c *Controller) Run(stopCh <-chan struct{}, workers int) {
	defer utilruntime.HandleCrash()

	// Let the workers stop when we are done
	defer c.queue.ShutDown()
	klog.Infof("Starting controller")

	// Starting the informer
	go c.informer.Run(stopCh)

	// Wait for all involved caches to be synced, before processing items from the queue is started
	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	<-stopCh
	klog.Infof("Stopping controller")
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
	}
}

func getHandler(queue workqueue.RateLimitingInterface, secretManager *secret.Manager) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			resource := getResource(obj)
			klog.Info("Add Event ", resource.Name)
			queue.Add(NewEvent(watch.Added, resource, secretManager))
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			oldResource := getResource(old)
			newResource := getResource(new)

			if getReferenceSecret(oldResource) != getReferenceSecret(newResource) {
				klog.Info("Update event ", "old ", oldResource.ResourceVersion, " new ", newResource.ResourceVersion, " newkey ", newResource.Name, " oldKey ", oldResource.Name)
				// remove old watch
				queue.Add(NewEvent(watch.Deleted, oldResource, secretManager))
				// create new watch
				queue.Add(NewEvent(watch.Added, newResource, secretManager))
			}
		},
		DeleteFunc: func(obj interface{}) {
			resource := getResource(obj)
			klog.Info("Delete event ", " obj ", resource.ResourceVersion, " key ", resource.Name)
			// remove associated secret watcher
			queue.Add(NewEvent(watch.Deleted, resource, secretManager))
		},
	}
}
