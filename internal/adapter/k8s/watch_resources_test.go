package k8s

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
)

type callbackMock struct {
	mock.Mock
}

func (m *callbackMock) Callback(obj *unstructured.Unstructured, gvr schema.GroupVersionResource) {
	m.Called(obj, gvr)
}

func TestStartWatchingGVRs_AddEvent(t *testing.T) {
	scheme := runtime.NewScheme()
	gvr := schema.GroupVersionResource{Group: "test", Version: "v1", Resource: "foos"}

	client := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, map[schema.GroupVersionResource]string{
		gvr: "FooList",
	})

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvr.GroupVersion().WithKind("Foo"))
	obj.SetName("foo1")
	obj.SetLabels(map[string]string{"managed-by": "vida"})

	watcher := watch.NewFake()
	client.PrependWatchReactor("foos", func(action k8stesting.Action) (bool, watch.Interface, error) {
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
	done := make(chan struct{})
	cb.On("Callback", mock.AnythingOfType("*unstructured.Unstructured"), gvr).Run(func(args mock.Arguments) {
		close(done)
	}).Once()

	factory := NewDynamicWatcherFactory()
	go factory.StartWatchingGVRs(client, []schema.GroupVersionResource{gvr}, cb.Callback)

	time.Sleep(100 * time.Millisecond)
	watcher.Add(obj)

	select {
	case <-done:
		cb.AssertExpectations(t)
	case <-time.After(1 * time.Second):
		t.Fatal("Expected callback not called")
	}
}

func TestAddFunc_WithTombstone_Unstructured(t *testing.T) {
	cb := new(callbackMock)

	gvr := schema.GroupVersionResource{Group: "test", Version: "v1", Resource: "foos"}

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvr.GroupVersion().WithKind("Foo"))
	obj.SetName("tombstoned-foo")
	obj.SetLabels(map[string]string{"managed-by": "vida"})

	tombstone := cache.DeletedFinalStateUnknown{
		Key: "test/tombstoned-foo",
		Obj: obj,
	}

	cb.On("Callback", obj, gvr).Once()

	handler := getEventHandler(gvr, cb.Callback)
	handler.AddFunc(tombstone)

	cb.AssertExpectations(t)
}

func TestUpdateFunc_GenerationChange(t *testing.T) {
	cb := new(callbackMock)

	gvr := schema.GroupVersionResource{Group: "test", Version: "v1", Resource: "foos"}

	oldObj := &unstructured.Unstructured{}
	oldObj.SetGeneration(1)
	oldObj.SetName("gen-change-foo")
	oldObj.SetLabels(map[string]string{"managed-by": "vida"})

	newObj := &unstructured.Unstructured{}
	newObj.SetGeneration(2)
	newObj.SetName("gen-change-foo")
	newObj.SetLabels(map[string]string{"managed-by": "vida"})

	cb.On("Callback", newObj, gvr).Once()

	handler := getEventHandler(gvr, cb.Callback)
	handler.UpdateFunc(oldObj, newObj)

	cb.AssertExpectations(t)
}

func TestUpdateFunc_NoGenerationChange(t *testing.T) {
	cb := new(callbackMock)

	gvr := schema.GroupVersionResource{Group: "test", Version: "v1", Resource: "foos"}

	obj := &unstructured.Unstructured{}
	obj.SetGeneration(1)
	obj.SetName("same-gen-foo")
	obj.SetLabels(map[string]string{"managed-by": "vida"})

	// No expected call since generation didn't change
	handler := getEventHandler(gvr, cb.Callback)
	handler.UpdateFunc(obj, obj)

	cb.AssertExpectations(t) // Ensures no unexpected calls were made
}

func TestDeleteFunc_WithUnstructured(t *testing.T) {
	cb := new(callbackMock)

	gvr := schema.GroupVersionResource{Group: "test", Version: "v1", Resource: "foos"}

	deleted := &unstructured.Unstructured{}
	deleted.SetName("my-resource")
	deleted.SetLabels(map[string]string{"managed-by": "vida"}) // <-- add this label

	cb.On("Callback", deleted, gvr).Once()

	handler := getEventHandler(gvr, cb.Callback)
	handler.DeleteFunc(deleted)

	cb.AssertExpectations(t)
}

func TestDeleteFunc_WithTombstone(t *testing.T) {
	cb := new(callbackMock)

	gvr := schema.GroupVersionResource{Group: "test", Version: "v1", Resource: "foos"}

	obj := &unstructured.Unstructured{}
	obj.SetName("tombstoned-foo")
	obj.SetLabels(map[string]string{"managed-by": "vida"})

	tombstone := cache.DeletedFinalStateUnknown{
		Key: "test/tombstoned-foo",
		Obj: obj,
	}

	cb.On("Callback", obj, gvr).Once()

	handler := getEventHandler(gvr, cb.Callback)
	handler.DeleteFunc(tombstone)

	cb.AssertExpectations(t)
}

type mockWriter struct {
	writeFunc func(p []byte) (n int, err error)
}

func (w mockWriter) Write(p []byte) (n int, err error) {
	return w.writeFunc(p)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}

func TestAddFunc_UnexpectedType(t *testing.T) {
	cb := new(callbackMock)

	gvr := schema.GroupVersionResource{Group: "test", Version: "v1", Resource: "foos"}

	handler := getEventHandler(gvr, cb.Callback)

	// Capture log output
	var logOutput string
	log.SetFlags(0)
	log.SetOutput(mockWriter{func(p []byte) (n int, err error) {
		logOutput = string(p)
		return len(p), nil
	}})

	handler.AddFunc(123) // Passing an int to trigger the unexpected type log

	if logOutput == "" || !contains(logOutput, "AddFunc: unexpected type int") {
		t.Errorf("expected log to contain 'AddFunc: unexpected type int', got: %q", logOutput)
	}

	cb.AssertExpectations(t)
}

func TestAddFunc_WithTombstone_UnexpectedType(t *testing.T) {
	cb := new(callbackMock)

	gvr := schema.GroupVersionResource{Group: "test", Version: "v1", Resource: "foos"}

	// Tombstone object with wrong type (e.g., string instead of *unstructured.Unstructured)
	tombstone := cache.DeletedFinalStateUnknown{
		Key: "test/bad-type",
		Obj: "this-is-not-an-unstructured-object",
	}

	handler := getEventHandler(gvr, cb.Callback)

	// No callback is expected
	handler.AddFunc(tombstone)

	cb.AssertExpectations(t)
}

func TestUpdateFunc_UnexpectedTypes(t *testing.T) {
	cb := new(callbackMock)

	gvr := schema.GroupVersionResource{Group: "test", Version: "v1", Resource: "foos"}

	// Use non-*unstructured.Unstructured types to trigger the log line
	oldObj := "this-is-not-unstructured"
	newObj := 42 // also not unstructured

	handler := getEventHandler(gvr, cb.Callback)

	// No callback is expected
	handler.UpdateFunc(oldObj, newObj)

	cb.AssertExpectations(t)
}

func TestDeleteFunc_UnexpectedType(t *testing.T) {
	cb := new(callbackMock)

	gvr := schema.GroupVersionResource{Group: "test", Version: "v1", Resource: "foos"}

	handler := getEventHandler(gvr, cb.Callback)

	invalidObj := "not-a-tombstone"

	handler.DeleteFunc(invalidObj)

	cb.AssertExpectations(t)
}

func TestDeleteFunc_UnexpectedTombstoneType(t *testing.T) {
	cb := new(callbackMock)

	gvr := schema.GroupVersionResource{Group: "test", Version: "v1", Resource: "foos"}

	handler := getEventHandler(gvr, cb.Callback)

	tombstone := cache.DeletedFinalStateUnknown{
		Key: "foo",
		Obj: "not-unstructured", // invalid tombstone.Obj type
	}

	handler.DeleteFunc(tombstone)

	cb.AssertExpectations(t)
}
