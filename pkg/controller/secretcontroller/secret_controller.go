package secretcontroller

import (
	"context"
	"reflect"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreinformersv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"

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
	// Sync awsCred
	awsCred := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-aws-cred",
			Namespace: "kube-system",
		},
		Data: map[string][]byte{
			"aws_access_key_id":     {1, 2, 3, 4},
			"aws_secret_access_key": {1, 2, 3, 4},
		},
	}

	// Ensure awsCred always exists
	gotSec, err := c.secretInformer.Lister().Secrets("kube-system").Get(awsCred.Name)
	if err != nil {
		if apierrors.IsNotFound(err) {
			klog.Info("Creating kube-system/my-aws-cred secret")
			if _, err := c.kubeClient.CoreV1().Secrets("kube-system").Create(ctx, awsCred, metav1.CreateOptions{}); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		// If gotSec exists, compare its data with awsCred and update if necessary
		if !reflect.DeepEqual(gotSec.Data, awsCred.Data) {
			klog.Info("Updating kube-system/my-aws-cred secret")
			gotSec.Data = awsCred.Data
			_, err := c.kubeClient.CoreV1().Secrets("kube-system").Update(ctx, gotSec, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
