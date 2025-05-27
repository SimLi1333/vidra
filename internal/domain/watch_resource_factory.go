package domain

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

type ResourceCallback func(obj *unstructured.Unstructured, gvr schema.GroupVersionResource)

type DynamicWatcherFactory interface {
	StartWatchingGVRs(dynamicClient dynamic.Interface, gvrs []schema.GroupVersionResource, onEvent ResourceCallback)
}
