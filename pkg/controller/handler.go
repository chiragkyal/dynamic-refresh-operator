package controller

import (
	"context"

	secret "github.com/chiragkyal/dynamic-refresh-operator/pkg/controller/monitor"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

type Handler interface {
	Handle() error
}

type Event struct {
	eventType     watch.EventType
	resource      *v1.ConfigMap
	secretManager *secret.Manager
}

func NewEvent(eventType watch.EventType, resource *v1.ConfigMap, secretManager *secret.Manager) Handler {
	return &Event{
		eventType:     eventType,
		resource:      resource,
		secretManager: secretManager,
	}
}

func (re *Event) Handle() error {
	resource := re.resource
	eventType := re.eventType
	switch eventType {

	case watch.Added:
		err := localRegister(re.secretManager, resource)
		if err != nil {
			klog.Error("failed to register resource")
			return err
		}

	case watch.Modified:
		// TODO: Restart deployment
		s, err := re.secretManager.GetSecret(resource.Namespace, resource.Name)
		if err != nil {
			klog.Error(err)
			return err
		}
		klog.Info("fetching secret data ", s.Data)

	case watch.Deleted:
		err := re.secretManager.UnregisterRoute(resource.Namespace, resource.Name)
		if err != nil {
			klog.Error(err)
			return err
		}
	default:
		klog.Error("invalid eventType", eventType)
	}
	return nil
}

func localRegister(secretManager *secret.Manager, resource *v1.ConfigMap) error {
	secreth := generateSecretHandler(secretManager, resource)
	secretManager.WithSecretHandler(secreth)
	return secretManager.RegisterRoute(context.Background(), resource.Namespace, resource.Name, getReferenceSecret(resource))

}

func generateSecretHandler(secretManager *secret.Manager, resource *v1.ConfigMap) cache.ResourceEventHandlerFuncs {
	// secret handler
	secreth := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			secret := obj.(*v1.Secret)
			klog.Info("Secret added ", "obj ", secret.ResourceVersion, " key ", secret.Name, " For ", resource.Namespace+"/"+resource.Name)
			secretManager.Queue().Add(NewEvent(watch.Modified, resource, secretManager))
		},
		UpdateFunc: func(old interface{}, new interface{}) {
			secretOld := old.(*v1.Secret)
			secretNew := new.(*v1.Secret)
			klog.Info("Secret updated ", "old ", secretOld.ResourceVersion, " new ", secretNew.ResourceVersion, " key ", secretNew.Name, " For ", resource.Namespace+"/"+resource.Name)
			secretManager.Queue().Add(NewEvent(watch.Modified, resource, secretManager))
		},
		DeleteFunc: func(obj interface{}) {
			secret := obj.(*v1.Secret)
			klog.Info("Secret deleted ", " obj ", secret.ResourceVersion, " key ", secret.Name, " For ", resource.Namespace+"/"+resource.Name)
			secretManager.Queue().Add(NewEvent(watch.Modified, resource, secretManager))
		},
	}
	return secreth
}
