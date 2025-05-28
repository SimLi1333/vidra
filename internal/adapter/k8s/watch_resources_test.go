package k8s

import (
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

type callbackMock struct {
	mock.Mock
}

func (m *callbackMock) Callback(obj *unstructured.Unstructured, gvr schema.GroupVersionResource) {
	m.Called(obj, gvr)
}

func TestStartWatchingGVRs_AddEvent(t *testing.T) {
	// Setup
	scheme := runtime.NewScheme()
	gvr := schema.GroupVersionResource{Group: "test", Version: "v1", Resource: "foos"}

	// Register GVR to ListKind mapping to prevent panic
	client := fake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[schema.GroupVersionResource]string{
			gvr: "FooList",
		},
	)

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvr.GroupVersion().WithKind("Foo"))
	obj.SetName("foo1")
	obj.SetLabels(map[string]string{"managed-by": "vida"})

	// Prepare the fake Watch interface
	watcher := watch.NewFake()

	client.PrependWatchReactor("foos", func(action k8stesting.Action) (handled bool, ret watch.Interface, err error) {
		return true, watcher, nil
	})

	client.PrependReactor("list", "foos", func(action k8stesting.Action) (bool, runtime.Object, error) {
		list := &unstructured.UnstructuredList{
			Items: []unstructured.Unstructured{},
		}
		list.SetGroupVersionKind(gvr.GroupVersion().WithKind("FooList"))
		return true, list, nil
	})

	cb := new(callbackMock)
	cb.On("Callback", mock.AnythingOfType("*unstructured.Unstructured"), gvr).Once()

	factory := NewDynamicWatcherFactory()

	// Act
	go factory.StartWatchingGVRs(client, []schema.GroupVersionResource{gvr}, cb.Callback)

	// Trigger the event
	time.Sleep(100 * time.Millisecond) // Wait for informer to start
	watcher.Add(obj)

	// Assert
	time.Sleep(100 * time.Millisecond)
	cb.AssertExpectations(t)
}

func TestStartWatchingGVRs_SkipAlreadyStarted(t *testing.T) {
	scheme := runtime.NewScheme()
	gvr := schema.GroupVersionResource{Group: "test", Version: "v1", Resource: "foos"}

	client := fake.NewSimpleDynamicClientWithCustomListKinds(
		scheme,
		map[schema.GroupVersionResource]string{
			gvr: "FooList",
		},
	)

	cb := new(callbackMock)
	cb.On("Callback", mock.AnythingOfType("*unstructured.Unstructured"), gvr).Once()

	factory := NewDynamicWatcherFactory()

	// Start watching the same GVR twice
	go factory.StartWatchingGVRs(client, []schema.GroupVersionResource{gvr}, cb.Callback)
	go factory.StartWatchingGVRs(client, []schema.GroupVersionResource{gvr}, cb.Callback)

	// Trigger a watch event
	watcher := watch.NewFake()
	client.PrependWatchReactor("foos", func(action k8stesting.Action) (bool, watch.Interface, error) {
		return true, watcher, nil
	})

	// Wait for goroutines to run
	time.Sleep(100 * time.Millisecond)

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvr.GroupVersion().WithKind("Foo"))
	obj.SetName("test-foo")
	obj.SetLabels(map[string]string{"managed-by": "vida"})

	watcher.Add(obj)

	time.Sleep(100 * time.Millisecond)

	// Expect callback to have been called once
	cb.AssertExpectations(t)
}

func factoryStartedMap(factory *DynamicWatcherFactory) map[schema.GroupVersionResource]struct{} {
	factoryMu := getUnexportedField(factory, "mu").(*sync.Mutex)
	factoryMu.Lock()
	defer factoryMu.Unlock()
	return getUnexportedField(factory, "started").(map[schema.GroupVersionResource]struct{})
}

// getUnexportedField uses reflection to get a private field for testing
func getUnexportedField(obj interface{}, fieldName string) interface{} {
	rv := reflect.ValueOf(obj).Elem()
	f := rv.FieldByName(fieldName)
	return f.Interface()
}
