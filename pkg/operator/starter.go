package operator

import (
	"context"
	"time"

	"k8s.io/client-go/informers"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"github.com/chiragkyal/dynamic-refresh-operator/pkg/controller"
	"github.com/chiragkyal/dynamic-refresh-operator/pkg/controller/secretcontroller"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	"github.com/openshift/library-go/pkg/controller/factory"
)

const (
	// Operand and operator run in the same namespace
	// defaultNamespace = "openshift-dynamic-refresh-operator"
	defaultNamespace = "default"
	operatorName     = "dynamic-refresh-operator"
)

func RunOperator(ctx context.Context, controllerConfig *controllercmd.ControllerContext) error {
	// Create core clientset and informers
	kubeClient := kubeclient.NewForConfigOrDie(rest.AddUserAgent(controllerConfig.KubeConfig, operatorName))
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())

	// controller 1
	controller := controller.NewController(kubeClient, defaultNamespace, queue)

	secretInformers := informers.NewSharedInformerFactory(kubeClient, 1*time.Minute).Core().V1().Secrets()

	// controller 2
	secretcontroller := secretcontroller.NewSecretController(
		"secret-controller",
		controllerConfig.EventRecorder,
		kubeClient,
		secretInformers,
		[]factory.Informer{},
	)

	klog.Info("Starting the informers")
	go secretInformers.Informer().Run(ctx.Done())

	klog.Info("Starting controller")
	go controller.Run(ctx.Done(), 1)
	go secretcontroller.Run(ctx, 1)

	<-ctx.Done()

	return nil
}
