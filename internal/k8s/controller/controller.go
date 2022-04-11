package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8sync/pkg/logger"

	"k8sync/internal/k8s/client"
	"k8sync/internal/k8s/handler"
	"k8sync/internal/k8s/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

var serverStartTime time.Time

// Event indicate the informerEvent
type Event struct {
	key          string
	eventType    string
	namespace    string
	resourceType string
	oldObj       interface{}
	//eventHandler handler.Handler
}

// Controller object
type Controller struct {
	//k8s          *client.K8s
	clientset    kubernetes.Interface
	queue        workqueue.RateLimitingInterface
	informer     cache.SharedIndexInformer
	eventHandler handler.Handler
}

// Start prepares watchers and run their controllers, then waits for process termination signals
func Start(ctx context.Context, k8s *client.K8s, eventHandler handler.Handler) {
	var kubeClient kubernetes.Interface
	var namespace string

	kubeClient = k8s.Clientset
	namespace = k8s.GetNamespace()

	informer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return kubeClient.CoreV1().Services(namespace).List(ctx, options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return kubeClient.CoreV1().Services(namespace).Watch(ctx, options)
			},
		},
		&corev1.Service{},
		0, // Skip resync
		cache.Indexers{},
	)

	rc := newResourceController(kubeClient, eventHandler, informer, "service")
	go rc.Run(ctx.Done())
}

func newResourceController(client kubernetes.Interface, eventHandler handler.Handler,
	informer cache.SharedIndexInformer, resourceType string) *Controller {
	var newEvent Event
	var err error

	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			newEvent.key, err = cache.MetaNamespaceKeyFunc(obj)
			newEvent.eventType = utils.EventTypeCreate
			newEvent.resourceType = resourceType
			newEvent.namespace = utils.GetObjectMetaData(obj).Namespace
			logger.Debugf("processing add to k8s %v: %s", resourceType, newEvent.key)
			if err == nil {
				queue.Add(newEvent)
			}
		},
		UpdateFunc: func(old, new interface{}) {
			newEvent.key, err = cache.MetaNamespaceKeyFunc(old)
			newEvent.eventType = utils.EventTypeUpdate
			newEvent.resourceType = resourceType
			newEvent.namespace = utils.GetObjectMetaData(old).Namespace
			newEvent.oldObj = old
			logger.Debugf("processing update to k8s %v: %s", resourceType, newEvent.key)
			if err == nil {
				queue.Add(newEvent)
			}
		},
		DeleteFunc: func(obj interface{}) {
			newEvent.key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			newEvent.eventType = utils.EventTypeDelete
			newEvent.resourceType = resourceType
			newEvent.namespace = utils.GetObjectMetaData(obj).Namespace
			newEvent.oldObj = obj
			logger.Debugf("processing delete to k8s %v: %s", resourceType, newEvent.key)
			if err == nil {
				queue.Add(newEvent)
			}
		},
	})

	return &Controller{
		clientset:    client,
		informer:     informer,
		queue:        queue,
		eventHandler: eventHandler,
	}
}

// Run starts the k8s controller
func (c *Controller) Run(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	logger.Info("Starting k8s controller")
	serverStartTime = time.Now().Local()

	go c.informer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	logger.Info("k8s controller synced and ready")

	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced is required for the cache.Controller interface.
func (c *Controller) HasSynced() bool {
	return c.informer.HasSynced()
}

// LastSyncResourceVersion is required for the cache.Controller interface.
func (c *Controller) LastSyncResourceVersion() string {
	return c.informer.LastSyncResourceVersion()
}

func (c *Controller) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *Controller) processNextItem() bool {
	newEvent, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(newEvent)

	err := c.processItem(newEvent.(Event))
	if err == nil {
		// No error, reset the ratelimit counters
		c.queue.Forget(newEvent)
	} else if c.queue.NumRequeues(newEvent) < utils.MaxRetries {
		logger.Errorf("processing %s failed (will retry): %v", newEvent.(Event).key, err)
		c.queue.AddRateLimited(newEvent)
	} else {
		// err != nil and too many retries
		logger.Errorf("processing %s over max %d retries (giving up): %v", newEvent.(Event).key, utils.MaxRetries, err)
		c.queue.Forget(newEvent)
		utilruntime.HandleError(err)
	}

	return true
}

/* TODOs
- Enhance event creation using client-side cacheing machanisms - pending
- Enhance the processItem to classify events - done
- Send alerts correspoding to events - done
*/

func (c *Controller) processItem(ctlEvent Event) error {
	// hold status type for default critical alerts
	var status string

	obj, _, err := c.informer.GetIndexer().GetByKey(ctlEvent.key)
	if err != nil {
		return fmt.Errorf("error fetching object with key %s from store: %w", ctlEvent.key, err)
	}
	// get object's metedata
	objectMeta := utils.GetObjectMetaData(obj)

	// namespace retrieval from event key incase namespace value is empty
	if ctlEvent.namespace == "" && strings.Contains(ctlEvent.key, "/") {
		substring := strings.Split(ctlEvent.key, "/")
		ctlEvent.namespace = substring[0]
		ctlEvent.key = substring[1]
	}

	// process events based on its type
	switch ctlEvent.eventType {
	case utils.EventTypeCreate:
		// compare CreationTimestamp and serverStartTime and alert only on latest events
		// Could be Replaced by using Delta or DeltaFIFO
		if objectMeta.CreationTimestamp.Sub(serverStartTime).Seconds() > 0 {
			switch ctlEvent.resourceType {
			case "NodeNotReady":
				status = utils.StatusDanger
			case "NodeReady":
				status = utils.StatusDanger
			case "NodeRebooted":
				status = utils.StatusDanger
			case "Backoff":
				status = utils.StatusDanger
			default:
				status = utils.StatusNormal
			}
			kbEvent := handler.New(obj, ctlEvent.namespace, ctlEvent.eventType, ctlEvent.resourceType, status)
			c.eventHandler.Handle(kbEvent)
			return nil
		}
	case utils.EventTypeUpdate:
		switch ctlEvent.resourceType {
		case "Backoff":
			status = utils.StatusDanger
		default:
			status = utils.StatusWarning
		}
		kbEvent := handler.New(obj, ctlEvent.namespace, ctlEvent.eventType, ctlEvent.resourceType, status)
		c.eventHandler.Handle(kbEvent)
		return nil
	case utils.EventTypeDelete:
		if obj == nil {
			obj = ctlEvent.oldObj
		}
		kbEvent := handler.New(obj, ctlEvent.namespace, ctlEvent.eventType, ctlEvent.resourceType, utils.StatusDanger)
		c.eventHandler.Handle(kbEvent)
		return nil
	}
	return nil
}
