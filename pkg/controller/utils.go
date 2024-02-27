package controller

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

func getResource(obj interface{}) *v1.ConfigMap {
	return obj.(*v1.ConfigMap)
}

func getReferenceSecret(resource *v1.ConfigMap) string {
	secretName := resource.Data["secretName"]
	klog.Info("Referenced secretName: ", secretName)
	return secretName
}
