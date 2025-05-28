// File: vidraresource_controller_test.go

package controller

import (
	"context"
	"fmt"
	"time"

	infrahubv1alpha1 "github.com/infrahub-operator/vidra/api/v1alpha1"
	"github.com/infrahub-operator/vidra/internal/domain"
	mock "github.com/infrahub-operator/vidra/internal/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("should reconcile correctly with different Destination.Server values", func() {
	servers := []string{
		"", // empty destination server
		"https://cldop-test-0.network.garden:6443",
	}

	for _, destinationServer := range servers {
		destinationServer := destinationServer // capture range variable

		Context(fmt.Sprintf("Destination.Server = '%s'", destinationServer), func() {
			var (
				mockCtrl                       *gomock.Controller
				mockRESTMapper                 *mock.MockRESTMapper
				mockDynamicMulticlusterFactory *mock.MockDynamicMulticlusterFactory
				mockWatcherFactory             *mock.MockDynamicWatcherFactory
				ctx                            context.Context
				namespacedName                 types.NamespacedName
				reconciler                     *VidraResourceReconciler
			)

			const (
				resourceName  = "test-resource"
				targetBranche = "main"
				targetDate    = "2025-01-01T00:00:00Z"
				namespace     = "default"
			)

			BeforeEach(func() {
				mockCtrl = gomock.NewController(GinkgoT())
				mockDynamicMulticlusterFactory = mock.NewMockDynamicMulticlusterFactory(mockCtrl)
				mockWatcherFactory = mock.NewMockDynamicWatcherFactory(mockCtrl)
				mockRESTMapper = mock.NewMockRESTMapper(mockCtrl)
				ctx = context.Background()
				namespacedName = types.NamespacedName{
					Name: resourceName,
				}
				reconciler = &VidraResourceReconciler{
					Client:                     k8sClient,
					Scheme:                     k8sClient.Scheme(),
					RESTMapper:                 mockRESTMapper,
					DynamicMulticlusterFactory: mockDynamicMulticlusterFactory,
				}
			})

			AfterEach(func() {
				By("deleting the custom resource for cleanup")
				mockCtrl.Finish()
			})

			Context("When reconciling a local resource", func() {
				BeforeEach(func() {
					deployK8sClient := setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)
					_ = deployK8sClient.Delete(ctx, &v1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "example-config",
							Namespace: "default",
						},
					})
					By("creating the custom resource for the Kind VidraResource if not exists")
					instance := &infrahubv1alpha1.VidraResource{}
					err := k8sClient.Get(ctx, namespacedName, instance)
					if err != nil && k8serrors.IsNotFound(err) {
						resource := &infrahubv1alpha1.VidraResource{
							ObjectMeta: metav1.ObjectMeta{
								Name:      resourceName,
								Namespace: namespace,
							},
							Spec: infrahubv1alpha1.VidraResourceSpec{
								Destination: infrahubv1alpha1.InfrahubSyncDestination{
									Server:    destinationServer,
									Namespace: namespace,
								},
							},
						}
						Expect(k8sClient.Create(ctx, resource)).To(Succeed())
					}
				})

				AfterEach(func() {
					By("deleting the custom resource for cleanup")
					instance := &infrahubv1alpha1.VidraResource{}
					err := k8sClient.Get(ctx, namespacedName, instance)
					if err == nil {
						// Wait for the finalizer to be removed before deleting the resource
						// instance.SetFinalizers(nil)
						deployK8sClient := setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)
						_ = deployK8sClient.Delete(ctx, &v1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "example-config",
								Namespace: "default",
							},
						})
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())
						Expect(k8sClient.Delete(ctx, instance)).To(Succeed())
						_ = setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)

						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())
					}
					if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
						failingClient.FailingMethod = ""
					}
				})
				Context("Once the resource exists", func() {
					It("should successfully reconcile the resource (json) and call InfrahubClient methods", func() {
						By("setting up the mock client to return a JSON ConfigMap response")
						instance := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: example
  namespace: ` + namespace + `
data:
  key: value
`
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						By("reconciling the resource on the destination server")
						deployK8sClient := setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)

						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						Eventually(func() error {
							cm := &v1.ConfigMap{}
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "example", Namespace: namespace}, cm)
						}).Should(Succeed())
					})

					It("should reconcile multiple resources from artifact", func() {
						By("setting up the mock client to return multiple resources")
						instance := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: config1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config2
`
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						By("reconciling the resource on the destination server")
						deployK8sClient := setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)

						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						created1 := &v1.ConfigMap{}
						err = deployK8sClient.Get(ctx, types.NamespacedName{Name: "config1", Namespace: namespace}, created1)
						Expect(err).NotTo(HaveOccurred())
						Expect(created1.Name).To(Equal("config1"))
						Expect(created1.Namespace).To(Equal(namespace))

						created2 := &v1.ConfigMap{}
						err = deployK8sClient.Get(ctx, types.NamespacedName{Name: "config2", Namespace: namespace}, created2)
						Expect(err).NotTo(HaveOccurred())
						Expect(created2.Name).To(Equal("config2"))
						Expect(created2.Namespace).To(Equal(namespace))
					})

					It("should update the deployed resource (yaml) if the manifest changes", func() {
						By("setting up the mock client to return a YAML webserver (1 replica) response")
						webserver := `
apiVersion: v1
kind: Namespace
metadata:
  name: test-namespace
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: test-namespace
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test-app
  template:
    metadata:
      labels:
        app: test-app
    spec:
      containers:
        - name: test-container
          image: nginx:latest
          ports:
            - containerPort: 80
`
						instance := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = webserver
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(schema.GroupKind{Group: "", Kind: "Namespace"}, "v1").
							Return(&meta.RESTMapping{
								Resource: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"},
								Scope:    meta.RESTScopeRoot,
							}, nil)
						mockRESTMapper.EXPECT().
							RESTMapping(schema.GroupKind{Group: "apps", Kind: "Deployment"}, "v1").
							Return(&meta.RESTMapping{
								Resource: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
								Scope:    meta.RESTScopeNamespace,
							}, nil)

						By("reconciling the resource on the destination server")
						deployK8sClient := setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)

						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						Eventually(func() error {
							ns := &v1.Namespace{}
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "test-namespace"}, ns)
						}).Should(Succeed())
						Eventually(func() error {
							deployment := &appsv1.Deployment{}
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "test-deployment", Namespace: "test-namespace"}, deployment)
						}).Should(Succeed())

						By("updating the artifact to have 2 replicas")
						err = k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						Expect(k8sClient.Status().Update(ctx, instance)).To(Succeed())

						yamlDataUpdated := `
apiVersion: v1
kind: Namespace
metadata:
  name: test-namespace
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
  namespace: test-namespace
spec:
  replicas: 2
  selector:
    matchLabels:
      app: test-app
  template:
    metadata:
      labels:
        app: test-app
    spec:
      containers:
        - name: test-container
          image: nginx:latest
          ports:
            - containerPort: 80
`
						err = k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = yamlDataUpdated
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())
						mockRESTMapper.EXPECT().
							RESTMapping(schema.GroupKind{Group: "", Kind: "Namespace"}, "v1").
							Return(&meta.RESTMapping{
								Resource: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"},
								Scope:    meta.RESTScopeRoot,
							}, nil)
						mockRESTMapper.EXPECT().
							RESTMapping(schema.GroupKind{Group: "apps", Kind: "Deployment"}, "v1").
							Return(&meta.RESTMapping{
								Resource: schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"},
								Scope:    meta.RESTScopeNamespace,
							}, nil)

						By("reconciling the resource on the destination server again")
						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						Eventually(func() error {
							ns := &v1.Namespace{}
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "test-namespace"}, ns)
						}).Should(Succeed())

						Eventually(func() error {
							deployment := &appsv1.Deployment{}
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "test-deployment", Namespace: "test-namespace"}, deployment)
						}).Should(Succeed())
						// Check if the replicas were updated
						Eventually(func() int32 {
							deployment := &appsv1.Deployment{}
							err := deployK8sClient.Get(ctx, types.NamespacedName{Name: "test-deployment", Namespace: "test-namespace"}, deployment)
							if err != nil {
								return 0
							}
							return *deployment.Spec.Replicas
						}).Should(Equal(int32(2)))
					})

					It("should start a webserver with 3 replicas and expose it via Service and Ingress", func() {
						yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: www
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: www
  namespace: www
  labels:
    app: www
spec:
  replicas: 3
  selector:
    matchLabels:
      app: www
  template:
    metadata:
      labels:
        app: www
    spec:
      containers:
      - name: www
        image: public.ecr.aws/pahudnet/nyancat-docker-image
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: www
  namespace: www
  labels:
    app: www
spec:
  ports:
  - port: 80
    name: www
  selector:
    app: www
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: www
  namespace: www
spec:
  rules:
  - host: demo.cldop-test-0.network.garden
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: www
            port:
              number: 80
`
						instance := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = yaml
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()
						deployK8sClient := setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)
						// Run reconciliation
						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						// Check Deployment exists and has 3 replicas
						deploy := &appsv1.Deployment{}
						Eventually(func(g Gomega) {
							err := deployK8sClient.Get(ctx, client.ObjectKey{Name: "www", Namespace: "www"}, deploy)
							g.Expect(err).NotTo(HaveOccurred())
							g.Expect(*deploy.Spec.Replicas).To(Equal(int32(3)))
						}).Should(Succeed())

						// Check Service exists
						svc := &v1.Service{}
						Expect(deployK8sClient.Get(ctx, client.ObjectKey{Name: "www", Namespace: "www"}, svc)).To(Succeed())
						Expect(svc.Spec.Ports).NotTo(BeEmpty())

						// Check Ingress exists
						ing := &networkingv1.Ingress{}
						Expect(deployK8sClient.Get(ctx, client.ObjectKey{Name: "www", Namespace: "www"}, ing)).To(Succeed())
						Expect(ing.Spec.Rules).NotTo(BeEmpty())
					})

					It("should reconcile the cached resource form the crd if the manifest does not change", func() {
						By("setting up the mock client to return a YAML with a resource")
						yamlData := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: resource
  namespace: ` + namespace + `
data:
  key: value`
						instance := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = yamlData
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(schema.GroupKind{Group: "", Kind: "ConfigMap"}, "v1").
							Return(&meta.RESTMapping{
								Resource: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"},
								Scope:    meta.RESTScopeNamespace,
							}, nil).AnyTimes()
						By("reconciling the resource on the destination server")
						deployK8sClient := setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)
						// First reconciliation - should create resource
						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())
						// resource should exist
						Eventually(func() error {
							res := &unstructured.Unstructured{}
							res.SetAPIVersion("v1")
							res.SetKind("ConfigMap")
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource", Namespace: namespace}, res)
						}).Should(Succeed())

						By("running reconciliation again with the same manifest")

						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())
						// Check if the resource still exists and has the same data
						Eventually(func() error {
							res := &unstructured.Unstructured{}
							res.SetAPIVersion("v1")
							res.SetKind("ConfigMap")
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource", Namespace: namespace}, res)
						}).Should(Succeed())
					})

					It("should overwrite the resource if it was manually changed", func() {
						By("setting up the mock client to return a YAML with a resource")
						yamlData := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: resource
  namespace: ` + namespace + `
data:
  key: value
`
						instance := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = yamlData
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(schema.GroupKind{Group: "", Kind: "ConfigMap"}, "v1").
							Return(&meta.RESTMapping{
								Resource: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"},
								Scope:    meta.RESTScopeNamespace,
							}, nil).AnyTimes()
						By("reconciling the resource on the destination server")
						deployK8sClient := setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)
						// First reconciliation - should create resource
						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())
						// resource should exist
						Eventually(func() error {
							res := &unstructured.Unstructured{}
							res.SetAPIVersion("v1")
							res.SetKind("ConfigMap")
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource", Namespace: namespace}, res)
						}).Should(Succeed())

						// Simulate manual change to the resource
						updatedResource := &unstructured.Unstructured{}
						updatedResource.SetAPIVersion("v1")
						updatedResource.SetKind("ConfigMap")
						updatedResource.SetName("resource")
						updatedResource.SetNamespace(namespace)
						updatedResource.Object["data"] = map[string]interface{}{
							"key": "new-value",
						}
						err = deployK8sClient.Update(ctx, updatedResource)
						Expect(err).NotTo(HaveOccurred())
						// Check if the resource was updated
						Eventually(func() error {
							res := &unstructured.Unstructured{}
							res.SetAPIVersion("v1")
							res.SetKind("ConfigMap")
							err := deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource", Namespace: namespace}, res)
							if err != nil {
								return err
							}
							if res.Object["data"].(map[string]interface{})["key"] != "new-value" {
								return fmt.Errorf("resource was not updated")
							}
							return nil
						}).Should(Succeed())
						// Now run reconciliation again
						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())
						// Check if the resource was overwritten
						Eventually(func() error {
							res := &unstructured.Unstructured{}
							res.SetAPIVersion("v1")
							res.SetKind("ConfigMap")
							err := deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource", Namespace: namespace}, res)
							if err != nil {
								return err
							}
							if res.Object["data"].(map[string]interface{})["key"] != "value" {
								return fmt.Errorf("resource was not overwritten")
							}
							return nil
						}).Should(Succeed())
					})

					It("should reconcile resources in to its namespace if a namespace is in the artifact", func() {
						By("setting up the mock client to return a YAML with a namespace and resources in it")

						yaml := `
apiVersion: v1
kind: Namespace
metadata:
  name: test2-namespace
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config1
  namespace: test2-namespace
---
apiVersion: v1
kind: Namespace
metadata:
  name: test2-namespace2
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config2
  namespace: test2-namespace2
`

						instance := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = yaml

						Expect(k8sClient.Update(ctx, instance)).To(Succeed())
						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						By("reconciling the resource on the destination server")
						deployK8sClient := setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)

						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						created1 := &v1.ConfigMap{}
						namespace := &v1.Namespace{}
						err = deployK8sClient.Get(ctx, types.NamespacedName{Name: "test2-namespace"}, namespace)
						Expect(err).NotTo(HaveOccurred())
						Expect(namespace.Name).To(Equal("test2-namespace"))
						err = deployK8sClient.Get(ctx, types.NamespacedName{Name: "config1", Namespace: "test2-namespace"}, created1)
						Expect(err).NotTo(HaveOccurred())
						Expect(created1.Name).To(Equal("config1"))
						Expect(created1.Namespace).To(Equal("test2-namespace"))

						created2 := &v1.ConfigMap{}
						err = deployK8sClient.Get(ctx, types.NamespacedName{Name: "config2", Namespace: "test2-namespace2"}, created2)
						Expect(err).NotTo(HaveOccurred())
						Expect(created2.Name).To(Equal("config2"))
						Expect(created2.Namespace).To(Equal("test2-namespace2"))
					})

					It("should remove old resources and create new ones if the name changes", func() {
						By("setting up the mock client to return a YAML with a resource")
						yamlData := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: resource-old
  namespace: ` + namespace + `
`
						instance := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = yamlData
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(schema.GroupKind{Group: "", Kind: "ConfigMap"}, "v1").
							Return(&meta.RESTMapping{
								Resource: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"},
								Scope:    meta.RESTScopeNamespace,
							}, nil).AnyTimes()

						By("reconciling the resource on the destination server")
						deployK8sClient := setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)

						// First reconciliation - should create resource-old
						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						// resource-old should exist
						Eventually(func() error {
							res := &unstructured.Unstructured{}
							res.SetAPIVersion("v1")
							res.SetKind("ConfigMap")
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-old", Namespace: namespace}, res)
						}).Should(Succeed())

						By("reconciling the resource on the destination server again with a new name")
						err = k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						Expect(k8sClient.Status().Update(ctx, instance)).To(Succeed())

						yamlDataNew := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: resource-new
  namespace: ` + namespace + `
`
						err = k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = yamlDataNew
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(schema.GroupKind{Group: "", Kind: "ConfigMap"}, "v1").
							Return(&meta.RESTMapping{
								Resource: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"},
								Scope:    meta.RESTScopeNamespace,
							}, nil).AnyTimes()

						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						// resource-new should exist
						Eventually(func() error {
							res := &unstructured.Unstructured{}
							res.SetAPIVersion("v1")
							res.SetKind("ConfigMap")
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-new", Namespace: namespace}, res)
						}).Should(Succeed())

						// resource-old should be deleted
						Eventually(func() error {
							res := &unstructured.Unstructured{}
							res.SetAPIVersion("v1")
							res.SetKind("ConfigMap")
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-old", Namespace: namespace}, res)
						}).Should(WithTransform(func(err error) string {
							if err == nil {
								return ""
							}
							return err.Error()
						}, ContainSubstring(`configmaps "resource-old" not found`)))
					})

					It("should update an existing resource if it is managed by the operator", func() {
						By("creating an existing ConfigMap")
						existingConfigMap := &v1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "example-config",
								Namespace: "default",
								Annotations: map[string]string{
									"managed-by":    vidraOperator,
									OwnerAnnotation: resourceName,
								},
							},
							Data: map[string]string{
								"key1": "value1",
							},
						}
						deployK8sClient := setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)
						Expect(deployK8sClient.Create(ctx, existingConfigMap)).To(Succeed())

						defer (func() {
							err := deployK8sClient.Delete(ctx, existingConfigMap)
							Expect(err).NotTo(HaveOccurred())
						})()

						By("setting up the mock client to return a YAML with an updated ConfigMap")
						yaml := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: example-config
data:
  key1: updated-value
`
						instance := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = yaml
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						By("reconciling the resource on the destination server")
						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						updatedConfigMap := &v1.ConfigMap{}
						err = deployK8sClient.Get(ctx, types.NamespacedName{Name: "example-config", Namespace: "default"}, updatedConfigMap)
						Expect(err).NotTo(HaveOccurred())

						// Ensure the resource was updated with new data
						Expect(updatedConfigMap.Data["key1"]).To(Equal("updated-value"))
					})

					It("should not update an existing resource if it is not managed by the operator", func() {
						By("creating an existing ConfigMap")
						existingConfigMap := &v1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "example-config",
								Namespace: "default",
							},
							Data: map[string]string{
								"key1": "value1",
							},
						}

						deployK8sClient := setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)

						Expect(deployK8sClient.Create(ctx, existingConfigMap)).To(Succeed())

						defer (func() {
							err := deployK8sClient.Delete(ctx, existingConfigMap)
							Expect(err).NotTo(HaveOccurred())
						})()

						By("setting up the mock client to return a YAML with an updated ConfigMap")
						yaml := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: example-config
data:
  key1: updated-value
`
						instance := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = yaml
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						By("reconciling the resource on the destination server")
						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("already exists but is not managed by this operator"))

						// Fetch the updated ConfigMap
						notUpdatedConfigMap := &v1.ConfigMap{}
						err = deployK8sClient.Get(ctx, types.NamespacedName{Name: "example-config", Namespace: "default"}, notUpdatedConfigMap)
						Expect(err).NotTo(HaveOccurred())

						// Ensure the resource was updated with new data
						Expect(notUpdatedConfigMap.Data["key1"]).To(Equal("value1"))
					})

				})

				Context("Once the managed resource is removed from infrahub", func() {
					It("should clean up stale resources during reconciliation", func() {
						By("setting up the mock client to return a YAML with multiple resources")
						yaml := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: resource-a
  namespace: ` + namespace + `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: resource-b
  namespace: ` + namespace + `
`
						instance := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = yaml
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(schema.GroupKind{Group: "", Kind: "ConfigMap"}, "v1").
							Return(&meta.RESTMapping{
								Resource: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"},
								Scope:    meta.RESTScopeNamespace,
							}, nil).AnyTimes()

						deployK8sClient := setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)

						By("reconciling the resource on the destination server")
						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						// resource-a should exist
						Eventually(func() error {
							a := &unstructured.Unstructured{}
							a.SetAPIVersion("v1")
							a.SetKind("ConfigMap")
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-a", Namespace: namespace}, a)
						}).Should(Succeed())

						// resource-b should exist
						Eventually(func() error {
							b := &unstructured.Unstructured{}
							b.SetAPIVersion("v1")
							b.SetKind("ConfigMap")
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-b", Namespace: namespace}, b)
						}).Should(Succeed())

						By("setting up the mock client to return a YAML with only resource-a")
						err = k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						Expect(k8sClient.Status().Update(ctx, instance)).To(Succeed())

						yamlDataNew := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: resource-a
  namespace: ` + namespace + `
`
						err = k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = yamlDataNew
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(schema.GroupKind{Group: "", Kind: "ConfigMap"}, "v1").
							Return(&meta.RESTMapping{
								Resource: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"},
								Scope:    meta.RESTScopeNamespace,
							}, nil).AnyTimes()

						// Run reconciliation
						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						// resource-a should exist
						Eventually(func() error {
							a := &unstructured.Unstructured{}
							a.SetAPIVersion("v1")
							a.SetKind("ConfigMap")
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-a", Namespace: namespace}, a)
						}).Should(Succeed())

						// resource-b should be deleted
						Eventually(func() error {
							b := &unstructured.Unstructured{}
							b.SetAPIVersion("v1")
							b.SetKind("ConfigMap")
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-b", Namespace: namespace}, b)
						}).Should(WithTransform(func(err error) string {
							if err == nil {
								return ""
							}
							return err.Error()
						}, ContainSubstring(`configmaps "resource-b" not found`)))

					})

					It("should skip if the managed resource is not found anymore during deletion", func() {
						By("setting up the mock client to return a YAML with resource-a")
						yamlData := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: resource-a
  namespace: ` + namespace + `
`
						instance := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = yamlData
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()
						By("reconciling the resource on the destination server")
						deployK8sClient := setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)

						_, err = reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						Eventually(func() error {
							cm := &v1.ConfigMap{}
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-a", Namespace: namespace}, cm)
						}).Should(Succeed())

						By("deleting resource-a")
						err = deployK8sClient.Delete(ctx, &v1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "resource-a",
								Namespace: namespace,
							},
						})
						Expect(err).NotTo(HaveOccurred())
						By("renaming the ConfigMap to resource-b and reconciling again")
						err = k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						Expect(k8sClient.Status().Update(ctx, instance)).To(Succeed())

						yaml := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: resource-b
  namespace: ` + namespace + `
data:
  key1: value1
`
						instance.Spec.Manifest = yaml
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						_, err = reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						// Check if the resource-b exists
						Eventually(func() error {
							cm := &v1.ConfigMap{}
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-b", Namespace: namespace}, cm)
						}).Should(Succeed())
						// Check if the resource-a does not exist anymore
						Eventually(func() error {
							cm := &v1.ConfigMap{}
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-a", Namespace: namespace}, cm)
						}).Should(WithTransform(func(err error) string {
							if err == nil {
								return ""
							}
							return err.Error()
						}, ContainSubstring(`configmaps "resource-a" not found`)))

					})

					It("should not delete the resource if it was overwritten before it is removed from managed resources", func() {
						By("setting up the mock client to return a YAML with resources a and b")
						yamlData := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: resource-a
  namespace: ` + namespace + `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: resource-b
  namespace: ` + namespace + `
`
						instance := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = yamlData
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						By("reconciling the resource on the destination server")
						deployK8sClient := setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)

						_, err = reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						Eventually(func() error {
							cm := &v1.ConfigMap{}
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-a", Namespace: namespace}, cm)
						}).Should(Succeed())

						By("overwriting resource-a with a new ConfigMap")
						newConfigMap := &v1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "resource-a",
								Namespace: namespace,
							},
							Data: map[string]string{"key1": "new-value"},
						}
						Expect(deployK8sClient.Update(ctx, newConfigMap)).To(Succeed())

						defer func() {
							// Clean up the new ConfigMap
							err := deployK8sClient.Delete(ctx, newConfigMap)
							Expect(err).NotTo(HaveOccurred())
						}()

						By("reconciling the resource on the destination server again with just resource-b")
						err = k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						Expect(k8sClient.Status().Update(ctx, instance)).To(Succeed())

						yaml := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: resource-b
  namespace: ` + namespace + `
data:
  key1: value1
`
						instance.Spec.Manifest = yaml
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						_, err = reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						Eventually(func() error {
							cm := &v1.ConfigMap{}
							err := deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-a", Namespace: namespace}, cm)
							return err
						}).Should(Succeed())

						Eventually(func() error {
							cm := &v1.ConfigMap{}
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-b", Namespace: namespace}, cm)
						}).Should(Succeed())
					})

					It("should handle it grasefully if the managed resource is already managed by another instance of inrfrahubResource (managed by two vidraResources)", func() {
						By("setting up the mock client to return a YAML with resources a in a namespace")
						infrahubRes := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, infrahubRes)
						Expect(err).NotTo(HaveOccurred())
						// Simulate downloading deployment resource
						yamlData := `
apiVersion: v1
kind: Namespace
metadata:
  name: test
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: resource-a
  namespace: test
`
						infrahubRes.Spec.Manifest = yamlData
						Expect(k8sClient.Update(ctx, infrahubRes)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(schema.GroupKind{Group: "", Kind: "Namespace"}, "v1").
							Return(&meta.RESTMapping{
								Resource: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"},
								Scope:    meta.RESTScopeRoot,
							}, nil).Times(2)
						mockRESTMapper.EXPECT().
							RESTMapping(schema.GroupKind{Group: "", Kind: "ConfigMap"}, "v1").
							Return(&meta.RESTMapping{
								Resource: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"},
								Scope:    meta.RESTScopeNamespace,
							}, nil).AnyTimes()

						By("reconciling the resource on the destination server")
						deployK8sClient := setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)

						_, err = reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						Eventually(func() error {
							cm := &v1.ConfigMap{}
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-a", Namespace: "test"}, cm)
						}).Should(Succeed())

						By("creating a new VidraResource that manages the same resource")
						newInfrahubRes := &infrahubv1alpha1.VidraResource{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "new-resource",
								Namespace: namespace,
							},
							Spec: infrahubv1alpha1.VidraResourceSpec{
								Destination: infrahubv1alpha1.InfrahubSyncDestination{
									Server:    destinationServer,
									Namespace: namespace,
								},
							},
						}

						Expect(k8sClient.Create(ctx, newInfrahubRes)).To(Succeed())

						By("reconciling the same namespace and an other resource in the second VidraResource")
						yamlData2 := `
apiVersion: v1
kind: Namespace
metadata:
  name: test
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: resource-b
  namespace: test
`
						infrahubRes2 := &infrahubv1alpha1.VidraResource{}
						err = k8sClient.Get(ctx, types.NamespacedName{Name: "new-resource", Namespace: namespace}, infrahubRes2)
						Expect(err).NotTo(HaveOccurred())
						infrahubRes2.Spec.Manifest = yamlData2
						Expect(k8sClient.Update(ctx, infrahubRes2)).To(Succeed())

						deployK8sClient = setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)

						_, err = reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: "new-resource", Namespace: namespace}})
						Expect(err).NotTo(HaveOccurred())
						// Check if the resource-a still exists
						Eventually(func() error {
							cm := &v1.ConfigMap{}
							err := deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-a", Namespace: "test"}, cm)
							return err
						}).Should(Succeed())
						// Check if the resource-b still exists
						Eventually(func() error {
							cm := &v1.ConfigMap{}
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-b", Namespace: "test"}, cm)
						}).Should(Succeed())
						// Check if the namespace exists and has the correct owner annotations
						testNamespace := &v1.Namespace{}
						err = deployK8sClient.Get(ctx, types.NamespacedName{Name: "test"}, testNamespace)
						Expect(err).NotTo(HaveOccurred())
						Expect(testNamespace.Annotations[OwnerAnnotation]).To(ContainSubstring(newInfrahubRes.Name))
						Expect(testNamespace.Annotations[OwnerAnnotation]).To(ContainSubstring(infrahubRes.Name))

						// Fetch the updated newInfrahubRes
						updatedNewInfrahubRes := &infrahubv1alpha1.VidraResource{}
						err = k8sClient.Get(ctx, types.NamespacedName{Name: "new-resource", Namespace: namespace}, updatedNewInfrahubRes)
						Expect(err).NotTo(HaveOccurred())

						// Expect the updated newInfrahubRes to have a warning in Status.LastError
						Expect(updatedNewInfrahubRes.Status.LastError).To(ContainSubstring(fmt.Sprintf(
							"Warning: resource is already managed by vidraResource: %s", infrahubRes.Name,
						)))

						By("Deleting the new vidraResources again / namespace should stay")
						instance := &infrahubv1alpha1.VidraResource{}
						err = k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())

						Expect(k8sClient.Status().Update(ctx, instance)).To(Succeed())

						err = k8sClient.Delete(ctx, newInfrahubRes)
						Expect(err).NotTo(HaveOccurred())
						_, err = reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: types.NamespacedName{Name: "new-resource", Namespace: namespace}})
						Expect(err).NotTo(HaveOccurred())
						// Check if the resource-a still exists
						Eventually(func() error {
							cm := &v1.ConfigMap{}
							err := deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-a", Namespace: "test"}, cm)
							return err
						}).Should(Succeed())
						// Check if the resource-b is deleted
						Eventually(func() error {
							cm := &v1.ConfigMap{}
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-b", Namespace: "test"}, cm)
						}).Should(WithTransform(func(err error) string {
							if err == nil {
								return ""
							}
							return err.Error()
						}, ContainSubstring(`configmaps "resource-b" not found`)))

						// Check if the namespace not deleted
						err = deployK8sClient.Get(ctx, types.NamespacedName{Name: "test"}, testNamespace)
						Expect(err).NotTo(HaveOccurred())
						Expect(testNamespace.Annotations[OwnerAnnotation]).To(ContainSubstring(infrahubRes.Name))

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						By("deleting the original vidraResource")
						err = k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						Expect(k8sClient.Delete(ctx, instance)).To(Succeed())

						_, err = reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())
						// Check if the resource-a id deleted
						Eventually(func() error {
							cm := &v1.ConfigMap{}
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-a", Namespace: "test"}, cm)
						}).Should(WithTransform(func(err error) string {
							if err == nil {
								return ""
							}
							return err.Error()
						}, ContainSubstring(`configmaps "resource-a" not found`)))

						// Check if the namespace is deleted
						Eventually(func() bool {
							ns := &v1.Namespace{}
							err := deployK8sClient.Get(ctx, types.NamespacedName{Name: "test"}, ns)
							if err != nil {
								return k8serrors.IsNotFound(err)
							}
							return ns.DeletionTimestamp != nil
						}, "30s", "1s").Should(BeTrue(), "expected namespace to be terminating or deleted")

					})

					It("should start watching and trigger reconciliation on event", func() {
						By("setting up the VidraResource with reconcileOnEvents=true")
						instance := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = `{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"example"},"data":{"key":"value"}}`
						instance.Spec.Destination.ReconcileOnEvents = true
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						cfg := ctrl.GetConfigOrDie()

						dynClient, err := dynamic.NewForConfig(cfg)
						ExpectWithOffset(1, err).NotTo(HaveOccurred(), "failed to create dynamic client")

						reconciler := &VidraResourceReconciler{
							Client:                     k8sClient,
							Scheme:                     k8sClient.Scheme(),
							RESTMapper:                 mockRESTMapper,
							DynamicMulticlusterFactory: mockDynamicMulticlusterFactory,
							DynamicWatcherFactory:      mockWatcherFactory,
							DynamicWatcherClient:       dynClient,
							EventBasedReconcile:        true,
						}

						deployK8sClient := setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)
						By("mocking the WatcherFactory to expect watching setup")
						mockWatcherFactory.EXPECT().
							StartWatchingGVRs(gomock.Any(), gomock.Any(), gomock.AssignableToTypeOf(func(*unstructured.Unstructured, schema.GroupVersionResource) {})).
							Do(func(_ dynamic.Interface, _ []schema.GroupVersionResource, cb domain.ResourceCallback) {
								By("simulating an external event")
								u := &unstructured.Unstructured{}
								u.SetAPIVersion("vidra.simli.dev/v1alpha1")
								u.SetKind("ConfigMap")
								u.SetNamespace(namespace)
								u.SetName("example")
								u.SetOwnerReferences([]metav1.OwnerReference{{
									APIVersion: infrahubv1alpha1.GroupVersion.String(),
									Kind:       "VidraResource",
									Name:       instance.Name,
								}})
								cb(u, schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"})
							})

						By("reconciling the resource")
						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						Eventually(func() error {
							cm := &v1.ConfigMap{}
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "example", Namespace: namespace}, cm)
						}).Should(Succeed())

						By("editing the ConfigMap to trigger another reconciliation")
						cm := &v1.ConfigMap{}
						err = deployK8sClient.Get(ctx, types.NamespacedName{Name: "example", Namespace: namespace}, cm)
						Expect(err).NotTo(HaveOccurred())
						cm.Data["key"] = "new-value"
						Expect(deployK8sClient.Update(ctx, cm)).To(Succeed())
						// Wait for the reconciliation to complete
						Eventually(func() error {
							updatedCM := &v1.ConfigMap{}
							err := deployK8sClient.Get(ctx, types.NamespacedName{Name: "example", Namespace: namespace}, updatedCM)
							if err != nil {
								return err
							}
							if updatedCM.Data["key"] != "new-value" {
								return fmt.Errorf("expected key to be 'new-value', got '%s'", updatedCM.Data["key"])
							}
							return nil
						}).Should(Succeed(), "expected ConfigMap to be updated with new value")

					})

				})
				Context("When the vidraresource is deleted", func() {
					It("should ignore not found error of a resource", func() {
						By("calling the reconcile function on a non-existent resource")
						result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: "nonexistent", Namespace: "default"}})

						// Expect no error and empty result (reconciliation ended)
						Expect(err).NotTo(HaveOccurred())
						Expect(result).To(Equal(reconcile.Result{}))
					})

					It("should not delete the managed resource if it was overwritten before finalizer handling", func() {
						By("setting up the mock client to return a YAML with resources a and b")
						res := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, res)
						Expect(err).NotTo(HaveOccurred())

						yamlData := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: resource-a
  namespace: ` + namespace + `
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: resource-b
  namespace: ` + namespace + `
`
						res.Spec.Manifest = yamlData
						Expect(k8sClient.Update(ctx, res)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						deployK8sClient := setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)

						By("reconciling the resource on the destination server")
						_, err = reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						Eventually(func() error {
							cm := &v1.ConfigMap{}
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-a", Namespace: namespace}, cm)
						}).Should(Succeed())

						By("overwriting resource-a with a new ConfigMap")
						newConfigMap := &v1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "resource-a",
								Namespace: namespace,
							},
							Data: map[string]string{"key1": "new-value"},
						}
						Expect(deployK8sClient.Delete(ctx, newConfigMap)).To(Succeed())
						Expect(deployK8sClient.Create(ctx, newConfigMap)).To(Succeed())

						By("deleting the VidraResource and reconciling again")
						Expect(k8sClient.Delete(ctx, res)).To(Succeed())

						_, err = reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						// res should be deleted
						Eventually(func() error {
							res := &infrahubv1alpha1.VidraResource{}
							err := k8sClient.Get(ctx, namespacedName, res)
							if k8serrors.IsNotFound(err) {
								return nil
							}
							if err != nil {
								return err
							}
							if len(res.Finalizers) > 0 {
								return fmt.Errorf("finalizer still present")
							}
							return nil
						}).Should(Succeed())

						// Check if the resource-a still exists
						Eventually(func() error {
							cm := &v1.ConfigMap{}
							err := deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-a", Namespace: namespace}, cm)
							return err
						}).Should(Succeed())
						// Check if the resource-b is deleted
						Eventually(func() error {
							cm := &v1.ConfigMap{}
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-b", Namespace: namespace}, cm)
						}).Should(WithTransform(func(err error) string {
							if err == nil {
								return ""
							}
							return err.Error()
						}, ContainSubstring(`configmaps "resource-b" not found`)))

					})
				})

				Context("Once a Error happens", func() {
					It("should return error if decoding fails", func() {
						By("setting up the mock client to return invalid YAML and reconcile")
						instance := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = "invalid: yaml: content"
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)

						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("decode artifact: error converting YAML to JSON: yaml"))
					})

					It("should return error if RESTMapping fails", func() {
						By("setting up the mock client to return a valid YAML but RESTMapping fails")
						instance := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = `{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"test"}}`
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(schema.GroupKind{Group: "", Kind: "ConfigMap"}, "v1").
							Return(nil, fmt.Errorf("REST mapping error"))

						setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)

						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("REST mapping error"))
					})

					It("should return an error when resource creation fails", func() {
						By("setting up the mock client to return a valid YAML and simulate a failure during resource creation")
						reconciler := &VidraResourceReconciler{
							Client:                     failingK8sClient,
							Scheme:                     k8sClient.Scheme(),
							RESTMapper:                 mockRESTMapper,
							DynamicMulticlusterFactory: mockDynamicMulticlusterFactory,
						}

						instance := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = `{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"example-config","namespace":"default"},"data":{"key1":"value1"}}`
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						// Act
						if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
							failingClient.FailingMethod = "Create"
						}

						setupDynamicMulticlusterFactoryMock(ctx, failingK8sClient, mockDynamicMulticlusterFactory, namespacedName, failingK8sClient)

						result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})

						// Assert: Check for the expected error
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("simulated failure: Create - &{map[apiVersion:v1 data:map[key1:value1]"))
						// Assert: Ensure the state of the VidraResource is marked as failed
						err = k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						Expect(instance.Status.DeployState).To(Equal(infrahubv1alpha1.StateFailed))
						// Assert: Ensure the returned result is empty (as the reconcile failed)
						Expect(result).To(Equal(reconcile.Result{}))
						if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
							failingClient.FailingMethod = ""
						}
					})

					It("should return an error when update an existing managed resource fails", func() {
						By("creating an existing ConfigMap and simulating a failure during update")
						existingConfigMap := &v1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "example-config",
								Namespace: "default",
								Annotations: map[string]string{
									"managed-by":    vidraOperator,
									OwnerAnnotation: resourceName,
								},
							},
							Data: map[string]string{
								"key1": "value1",
							},
						}

						deployK8sClient := setupDynamicMulticlusterFactoryMock(ctx, failingK8sClient, mockDynamicMulticlusterFactory, namespacedName, failingK8sClient)

						defer (func() {
							err := deployK8sClient.Delete(ctx, &v1.ConfigMap{
								ObjectMeta: metav1.ObjectMeta{
									Name:      "example-config",
									Namespace: "default",
								},
							})
							Expect(err).NotTo(HaveOccurred())
						})()

						Expect(deployK8sClient.Create(ctx, existingConfigMap)).To(Succeed())

						// Add a finalizer to the existing resource
						instance := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.SetFinalizers([]string{FinalizerName})
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						instance = &infrahubv1alpha1.VidraResource{}
						err = k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = `{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"example-config","namespace":"default"},"data":{"key1":"updated-value"}}`
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						reconciler := &VidraResourceReconciler{
							Client:                     failingK8sClient,
							Scheme:                     k8sClient.Scheme(),
							RESTMapper:                 mockRESTMapper,
							DynamicMulticlusterFactory: mockDynamicMulticlusterFactory,
						}
						if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
							failingClient.FailingMethod = "Update"
						}

						By("reconciling the resource on the destination server simulating a failure during update")
						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})

						// Assert: Check that the resource was updated successfully
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("simulated failure: Update - &{map[apiVersion:v1 data:map[key1:updated-value]"))

						// Fetch the updated ConfigMap
						notUpdatedConfigMap := &v1.ConfigMap{}
						err = deployK8sClient.Get(ctx, types.NamespacedName{Name: "example-config", Namespace: "default"}, notUpdatedConfigMap)
						Expect(err).NotTo(HaveOccurred())

						Expect(notUpdatedConfigMap.Data["key1"]).To(Equal("value1"))
						if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
							failingClient.FailingMethod = ""
						}
					})

					It("should return an error when patching the state of vidraResource with finalizer fails", func() {
						By("creating an existing ConfigMap and simulating a failure during Patch (State)")
						reconciler := &VidraResourceReconciler{
							Client:                     failingK8sClient,
							Scheme:                     k8sClient.Scheme(),
							RESTMapper:                 mockRESTMapper,
							DynamicMulticlusterFactory: mockDynamicMulticlusterFactory,
						}
						if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
							failingClient.FailingMethod = "Patch"
						}
						setupDynamicMulticlusterFactoryMock(ctx, failingK8sClient, mockDynamicMulticlusterFactory, namespacedName, failingK8sClient)

						_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("simulated failure: Patch - &{{ } {test-resource"))

						if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
							failingClient.FailingMethod = ""
						}
					})

					It("should return an error when deleting an existing managed resource fails and mark the resource stail", func() {
						By("creating an existing ConfigMap by reconcyling it and simulating a failure during update")
						instance := &infrahubv1alpha1.VidraResource{}
						err := k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.Spec.Manifest = `{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"example-config","namespace":"default"},"data":{"key1":"value1"}}`
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						// Set up the initial resource in the cluster
						reconciler := &VidraResourceReconciler{
							Client:                     failingK8sClient,
							Scheme:                     k8sClient.Scheme(),
							RESTMapper:                 mockRESTMapper,
							DynamicMulticlusterFactory: mockDynamicMulticlusterFactory,
						}
						deployK8sClient := setupDynamicMulticlusterFactoryMock(ctx, failingK8sClient, mockDynamicMulticlusterFactory, namespacedName, failingK8sClient)
						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						By("updating the existing ConfigMap and recycling it with Delete failure")
						err = k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						Expect(k8sClient.Status().Update(ctx, instance)).To(Succeed())

						instance.Spec.Manifest = `{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"example-config2","namespace":"default"},"data":{"key1":"updated-value"}}`
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
							failingClient.FailingMethod = "Delete"
						}

						_, err = reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("simulated failure: Delete - &{map[apiVersion:v1 data:map[key1:value1] kind:ConfigMap"))

						// Fetch the updated ConfigMap
						notUpdatedConfigMap := &v1.ConfigMap{}
						err = deployK8sClient.Get(ctx, types.NamespacedName{Name: "example-config", Namespace: "default"}, notUpdatedConfigMap)
						Expect(err).NotTo(HaveOccurred())

						Expect(notUpdatedConfigMap.Data["key1"]).To(Equal("value1"))
						if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
							failingClient.FailingMethod = ""
						}

						// Fetch the updated ConfigMap "example-config2"
						newConfigMap := &v1.ConfigMap{}
						err = deployK8sClient.Get(ctx, types.NamespacedName{Name: "example-config2", Namespace: "default"}, newConfigMap)
						Expect(err).NotTo(HaveOccurred())
						Expect(newConfigMap.Data["key1"]).To(Equal("updated-value"))
						// Delete the ConfigMap "example-config"
						err = deployK8sClient.Delete(ctx, &v1.ConfigMap{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "example-config",
								Namespace: "default",
							},
						})
						Expect(err).NotTo(HaveOccurred())

						// Verify that the DeployState is set to StateStale
						err = k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						Expect(instance.Status.DeployState).To(Equal(infrahubv1alpha1.StateStale))

					})

					It("should return error if Get fails with unexpected error", func() {
						By("setting up the mock client to return a valid YAML and simulate a failure during Get")
						reconciler := &VidraResourceReconciler{
							Client:                     failingK8sClient,
							Scheme:                     nil, // Not needed for this test
							RESTMapper:                 mockRESTMapper,
							DynamicMulticlusterFactory: mockDynamicMulticlusterFactory,
						}
						if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
							failingClient.FailingMethod = "Get"
						}
						setupDynamicMulticlusterFactoryMock(ctx, k8sClient, mockDynamicMulticlusterFactory, namespacedName, secondK8sClient)

						result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})

						Expect(err.Error()).To(ContainSubstring("simulated failure: Get"))
						Expect(result).To(Equal(reconcile.Result{}))
						if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
							failingClient.FailingMethod = ""
						}
					})
				})
			})
		})
	}
})

var _ = Describe("VidraResourceReconciler SetupWithManager", func() {
	var (
		mgr        manager.Manager
		reconciler *VidraResourceReconciler
		mockCtrl   *gomock.Controller
	)

	BeforeEach(func() {
		var err error
		mockCtrl = gomock.NewController(GinkgoT())

		mgr, err = ctrl.NewManager(cfg, ctrl.Options{
			Scheme: k8sClient.Scheme(),
		})
		Expect(err).NotTo(HaveOccurred())

		reconciler = &VidraResourceReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should create a ConfigMap for the controller and set RequeueAfter", func() {
		By("creating the ConfigMap for the controller")
		configMap := &v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "vidra-config",
				Namespace: "default",
				Labels: map[string]string{
					"app": "vidra",
				},
			},
			Data: map[string]string{
				"requeueRecourcesAfter": "12m",
				"eventBasedReconcile":   "true",
			},
		}
		err := k8sClient.Create(ctx, configMap)
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			By("deleting the ConfigMap for cleanup")
			err := k8sClient.Delete(ctx, configMap)
			Expect(err).NotTo(HaveOccurred())
		}()

		By("setting up the reconciler with the manager")
		err = reconciler.SetupWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())

		configMap = &v1.ConfigMap{}
		err = k8sClient.Get(ctx, types.NamespacedName{
			Name:      "vidra-config",
			Namespace: "default",
		}, configMap)
		Expect(err).NotTo(HaveOccurred())
		Expect(reconciler.RequeueAfter).To(Equal(12 * time.Minute))
		Expect(reconciler.EventBasedReconcile).To(BeTrue(), "EventBasedReconcile should be true")
	})
})
