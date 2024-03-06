package secretcontroller

import (
	"context"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreinformersv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/openshift/library-go/pkg/controller/factory"
	"github.com/openshift/library-go/pkg/operator/events"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// SecretController is a generic controller that manages a deployment.
type SecretController struct {
	name       string
	kubeClient kubernetes.Interface
	// operatorClient v1helpers.OperatorClientWithFinalizers
	secretInformer coreinformersv1.SecretInformer
}

func NewSecretController(
	name string,
	recorder events.Recorder,
	// operatorClient v1helpers.OperatorClientWithFinalizers,
	kubeClient kubernetes.Interface,
	secretInformer coreinformersv1.SecretInformer,
	optionalInformers []factory.Informer,
) factory.Controller {
	c := &SecretController{
		name: name,
		// operatorClient: operatorClient,
		kubeClient:     kubeClient,
		secretInformer: secretInformer,
	}

	informers := append(
		optionalInformers,
		//	operatorClient.Informer(),
		secretInformer.Informer(),
	)

	return factory.New().WithInformers(
		informers...,
	).WithSync(
		c.sync,
	).ResyncEvery(
		time.Minute,
	// ).WithSyncDegradedOnError(
	// 	operatorClient,
	).ToController(
		c.name,
		recorder.WithComponentSuffix(strings.ToLower(name)+"-controller-"),
	)
}

func (c *SecretController) Name() string {
	return c.name
}

func (c *SecretController) sync(ctx context.Context, syncContext factory.SyncContext) error {
	// sync awsCred
	awsCred := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "awsCred",
			Namespace: "kube-system",
		},
		Data: map[string][]byte{
			"aws_access_key_id":     {1, 2, 3, 4},
			"aws_secret_access_key": {1, 2, 3, 4},
		},
	}

	// ensure awsCred always exists
	if _, err := c.kubeClient.CoreV1().Secrets("kube-system").Get(ctx, awsCred.Name, metav1.GetOptions{}); err != nil {
		if apierrors.IsNotFound(err) {
			if _, err := c.kubeClient.CoreV1().Secrets("kube-system").Create(ctx, awsCred, metav1.CreateOptions{}); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
