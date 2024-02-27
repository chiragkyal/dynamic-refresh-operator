package operator

import (
	"context"

	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"github.com/chiragkyal/dynamic-refresh-operator/pkg/controller"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
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

	controller := controller.NewController(kubeClient, defaultNamespace, queue)

	klog.Info("Starting controller")
	go controller.Run(ctx.Done(), 1)

	<-ctx.Done()

	return nil
}
