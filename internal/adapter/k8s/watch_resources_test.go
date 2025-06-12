package k8s

import (
	"log"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

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

type mockWriter struct {
	writeFunc func(p []byte) (n int, err error)
}

func (w mockWriter) Write(p []byte) (n int, err error) {
	return w.writeFunc(p)
}

var _ = Describe("DynamicWatcher", func() {
	var (
		scheme runtime.Scheme
		gvr    schema.GroupVersionResource
	)

	BeforeEach(func() {
		scheme = *runtime.NewScheme()
		gvr = schema.GroupVersionResource{Group: "test", Version: "v1", Resource: "foos"}
	})

	Describe("StartWatchingGVRs", func() {
		It("should call callback on added event", func() {
			client := fake.NewSimpleDynamicClientWithCustomListKinds(&scheme, map[schema.GroupVersionResource]string{
				gvr: "FooList",
			})

			obj := &unstructured.Unstructured{}
			obj.SetGroupVersionKind(gvr.GroupVersion().WithKind("Foo"))
			obj.SetName("foo1")
			obj.SetLabels(map[string]string{"managed-by": "vidra"})

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

			Eventually(done, "1s").Should(BeClosed())
			cb.AssertExpectations(GinkgoT())
		})
	})

	Describe("EventHandler AddFunc", func() {
		It("should handle tombstone unstructured object", func() {
			cb := new(callbackMock)
			obj := &unstructured.Unstructured{}
			obj.SetGroupVersionKind(gvr.GroupVersion().WithKind("Foo"))
			obj.SetName("tombstoned-foo")
			obj.SetLabels(map[string]string{"managed-by": "vidra"})

			tombstone := cache.DeletedFinalStateUnknown{
				Key: "test/tombstoned-foo",
				Obj: obj,
			}

			cb.On("Callback", obj, gvr).Once()

			handler := getEventHandler(gvr, cb.Callback)
			handler.AddFunc(tombstone)

			cb.AssertExpectations(GinkgoT())
		})

		It("should log unexpected type on AddFunc", func() {
			cb := new(callbackMock)

			handler := getEventHandler(gvr, cb.Callback)

			var logOutput string
			log.SetFlags(0)
			log.SetOutput(mockWriter{func(p []byte) (n int, err error) {
				logOutput = string(p)
				return len(p), nil
			}})

			handler.AddFunc(123) // pass int to trigger log

			Expect(logOutput).To(ContainSubstring("AddFunc: unexpected type int"))
			cb.AssertExpectations(GinkgoT())
		})

		It("should handle tombstone with unexpected type silently", func() {
			cb := new(callbackMock)

			tombstone := cache.DeletedFinalStateUnknown{
				Key: "test/bad-type",
				Obj: "this-is-not-an-unstructured-object",
			}

			handler := getEventHandler(gvr, cb.Callback)
			handler.AddFunc(tombstone)

			cb.AssertExpectations(GinkgoT())
		})
	})

	Describe("EventHandler UpdateFunc", func() {
		It("should call callback on generation change", func() {
			cb := new(callbackMock)

			oldObj := &unstructured.Unstructured{}
			oldObj.SetGeneration(1)
			oldObj.SetName("gen-change-foo")
			oldObj.SetLabels(map[string]string{"managed-by": "vidra"})

			newObj := &unstructured.Unstructured{}
			newObj.SetGeneration(2)
			newObj.SetName("gen-change-foo")
			newObj.SetLabels(map[string]string{"managed-by": "vidra"})

			cb.On("Callback", newObj, gvr).Once()

			handler := getEventHandler(gvr, cb.Callback)
			handler.UpdateFunc(oldObj, newObj)

			cb.AssertExpectations(GinkgoT())
		})

		It("should NOT call callback if generation unchanged", func() {
			cb := new(callbackMock)

			obj := &unstructured.Unstructured{}
			obj.SetGeneration(1)
			obj.SetName("same-gen-foo")
			obj.SetLabels(map[string]string{"managed-by": "vidra"})

			handler := getEventHandler(gvr, cb.Callback)
			handler.UpdateFunc(obj, obj)

			cb.AssertExpectations(GinkgoT()) // no calls expected
		})

		It("should silently handle unexpected update types", func() {
			cb := new(callbackMock)

			oldObj := "not-unstructured"
			newObj := 42

			handler := getEventHandler(gvr, cb.Callback)
			handler.UpdateFunc(oldObj, newObj)

			cb.AssertExpectations(GinkgoT())
		})
	})

	Describe("EventHandler DeleteFunc", func() {
		It("should call callback on unstructured delete", func() {
			cb := new(callbackMock)

			deleted := &unstructured.Unstructured{}
			deleted.SetName("my-resource")
			deleted.SetLabels(map[string]string{"managed-by": "vidra"})

			cb.On("Callback", deleted, gvr).Once()

			handler := getEventHandler(gvr, cb.Callback)
			handler.DeleteFunc(deleted)

			cb.AssertExpectations(GinkgoT())
		})

		It("should call callback on tombstone delete", func() {
			cb := new(callbackMock)

			obj := &unstructured.Unstructured{}
			obj.SetName("tombstoned-foo")
			obj.SetLabels(map[string]string{"managed-by": "vidra"})

			tombstone := cache.DeletedFinalStateUnknown{
				Key: "test/tombstoned-foo",
				Obj: obj,
			}

			cb.On("Callback", obj, gvr).Once()

			handler := getEventHandler(gvr, cb.Callback)
			handler.DeleteFunc(tombstone)

			cb.AssertExpectations(GinkgoT())
		})

		It("should silently handle unexpected delete types", func() {
			cb := new(callbackMock)
			handler := getEventHandler(gvr, cb.Callback)

			invalidObj := "not-a-tombstone"
			handler.DeleteFunc(invalidObj)

			cb.AssertExpectations(GinkgoT())
		})

		It("should silently handle unexpected tombstone delete types", func() {
			cb := new(callbackMock)
			handler := getEventHandler(gvr, cb.Callback)

			tombstone := cache.DeletedFinalStateUnknown{
				Key: "foo",
				Obj: "not-unstructured",
			}

			handler.DeleteFunc(tombstone)

			cb.AssertExpectations(GinkgoT())
		})
	})
})
