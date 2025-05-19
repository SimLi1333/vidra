package mock

import (
	"context"
	"fmt"

	infrahubv1alpha1 "gitlab.ost.ch/ins-stud/sa-ba/ba-fs25-infrahub/infrahub-operator/api/v1alpha1"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/scale/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FailingUpdateClient struct {
	client.Client
	UpdateErrMsg  string
	RestMapper    meta.RESTMapper
	FailingMethod string // Holds the name of the function that needs to fail
}

func (f *FailingUpdateClient) shouldFail(methodName string) bool {
	return f.FailingMethod == methodName
}

// NewFakeFailingUpdateClient ensures that all fields are initialized properly
func NewFakeFailingUpdateClient(client client.Client, errorMsg string, restMapper MockRESTMapper) *FailingUpdateClient {
	// Return the struct with mockClient and mockRESTMapper initialized
	return &FailingUpdateClient{
		Client:        client,      // delegate most calls to the original client
		UpdateErrMsg:  errorMsg,    // inject the error message if provided
		RestMapper:    &restMapper, // Use mockRESTMapper here
		FailingMethod: "",          // Initialize FailingMethod to an empty string
	}
}

func (f *FailingUpdateClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if f.shouldFail("Update") {
		return fmt.Errorf("%s: Update - %v", f.UpdateErrMsg, obj)
	}
	return f.Client.Update(ctx, obj, opts...)
}

func (f *FailingUpdateClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if f.shouldFail("Create") {
		return fmt.Errorf("%s: Create - %v", f.UpdateErrMsg, obj)
	}
	return f.Client.Create(ctx, obj, opts...)
}

func (f *FailingUpdateClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if f.shouldFail("Get") {
		return fmt.Errorf("%s: Get - %v", f.UpdateErrMsg, obj)
	}
	return f.Client.Get(ctx, key, obj, opts...)
}

func (f *FailingUpdateClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	if f.shouldFail("Delete") {
		return fmt.Errorf("%s: Delete - %v", f.UpdateErrMsg, obj)
	}
	return f.Client.Delete(ctx, obj, opts...)
}

func (f *FailingUpdateClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	// For the purpose of the test, you can simulate a deletion failure or delegate it
	return f.Client.DeleteAllOf(ctx, obj, opts...)
}

func (f *FailingUpdateClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	// Delegate to the underlying client for the List operation
	if f.shouldFail("List") {
		return fmt.Errorf("%s: List - %v", f.UpdateErrMsg, opts)
	}
	return f.Client.List(ctx, list, opts...)
}

func (f *FailingUpdateClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	if f.shouldFail("Patch") {
		return fmt.Errorf("%s: Patch - %v", f.UpdateErrMsg, obj)
	}
	return f.Client.Patch(ctx, obj, patch, opts...)
}

func (f *FailingUpdateClient) Status() client.StatusWriter {
	// Return the status writer for the underlying client
	return f.Client.Status()
}

// Implement GroupVersionKindFor
func (f *FailingUpdateClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	// Return a default GVK for testing
	return schema.GroupVersionKind{
		Group:   "",   // Core group
		Version: "v1", // For core resources like ConfigMap
		Kind:    "ConfigMap",
	}, nil
}

func (f *FailingUpdateClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	// You can return true for namespaced resources like ConfigMap
	return true, nil
}

func (f *FailingUpdateClient) RESTMapper() meta.RESTMapper {
	return f.RestMapper
}

// Implement other methods as needed for the fake client

var testScheme = runtime.NewScheme()

func init() {
	// Add core Kubernetes types
	_ = scheme.AddToScheme(testScheme)

	// Add your CRDs
	_ = infrahubv1alpha1.AddToScheme(testScheme)
}
