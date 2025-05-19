// File: infrahubresource_controller_test.go

package controller

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	infrahubv1alpha1 "github.com/simli1333/vidra/api/v1alpha1"
	mock "github.com/simli1333/vidra/internal/mocks"
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
				mockCtrl          *gomock.Controller
				mockClient        *mock.MockInfrahubClient
				mockRESTMapper    *mock.MockRESTMapper
				mockClientFactory *mock.MockClientFactory
				ctx               context.Context
				namespacedName    types.NamespacedName
				reconciler        *InfrahubResourceReconciler
			)

			const (
				resourceName  = "test-resource"
				apiURL        = "https://example.com"
				targetBranche = "main"
				targetDate    = "2025-01-01T00:00:00Z"
				artifactName  = "test-artifact"
				artifactID    = "artifact-12345"
				checksum      = "checksum-12345"
				storageID     = "storage-12345"
				namespace     = "default"
			)

			BeforeEach(func() {
				mockCtrl = gomock.NewController(GinkgoT())
				mockClient = mock.NewMockInfrahubClient(mockCtrl)
				mockClientFactory = mock.NewMockClientFactory(mockCtrl)
				mockRESTMapper = mock.NewMockRESTMapper(mockCtrl)
				ctx = context.Background()
				namespacedName = types.NamespacedName{
					Name: resourceName,
				}
				reconciler = &InfrahubResourceReconciler{
					Client:         k8sClient,
					Scheme:         k8sClient.Scheme(),
					InfrahubClient: mockClient,
					RESTMapper:     mockRESTMapper,
					ClientFactory:  mockClientFactory,
				}
			})

			AfterEach(func() {
				By("deleting the custom resource for cleanup")
				mockCtrl.Finish()
			})

			Context("When reconciling a local resource", func() {
				BeforeEach(func() {
					By("creating the custom resource for the Kind InfrahubResource if not exists")
					instance := &infrahubv1alpha1.InfrahubResource{}
					err := k8sClient.Get(ctx, namespacedName, instance)
					if err != nil && k8serrors.IsNotFound(err) {
						resource := &infrahubv1alpha1.InfrahubResource{
							ObjectMeta: metav1.ObjectMeta{
								Name:      resourceName,
								Namespace: namespace,
							},
							Spec: infrahubv1alpha1.InfrahubResourceSpec{
								Source: infrahubv1alpha1.InfrahubSyncSource{
									InfrahubAPIURL: apiURL,
									TargetBranch:   targetBranche,
									TargetDate:     targetDate,
									ArtifactName:   artifactName,
								},
								Destination: infrahubv1alpha1.InfrahubSyncDestination{
									Server:    destinationServer,
									Namespace: namespace,
								},
								IDs: infrahubv1alpha1.InfrahubResourceIDs{
									ArtifactID: artifactID,
									Checksum:   checksum,
									StorageID:  storageID,
								},
							},
						}
						Expect(k8sClient.Create(ctx, resource)).To(Succeed())
					}
				})

				AfterEach(func() {
					By("deleting the custom resource for cleanup")
					instance := &infrahubv1alpha1.InfrahubResource{}
					err := k8sClient.Get(ctx, namespacedName, instance)
					if err == nil {
						// Wait for the finalizer to be removed before deleting the resource
						instance.SetFinalizers(nil)
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())
						Expect(k8sClient.Delete(ctx, instance)).To(Succeed())
					}
					if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
						failingClient.FailingMethod = ""
					}
				})
				Context("Once the resource exists", func() {
					It("should successfully reconcile the resource (json) and call InfrahubClient methods", func() {
						By("setting up the mock client to return a JSON ConfigMap response")
						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(bytes.NewReader([]byte(`{
							"apiVersion": "v1", 
							"kind": "ConfigMap", 
							"metadata": {
								"name": "example"
							}
						}`)), nil)

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						By("reconciling the resource on the destination server")
						deployK8sClient := setupClientFactoryMock(ctx, k8sClient, mockClientFactory, namespacedName, secondK8sClient)

						_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						Eventually(func() error {
							cm := &v1.ConfigMap{}
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "example", Namespace: namespace}, cm)
						}).Should(Succeed())
					})

					It("should reconcile multiple resources from artifact", func() {
						By("setting up the mock client to return multiple resources")
						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(bytes.NewReader([]byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: config1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config2
`)), nil)

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						By("reconciling the resource on the destination server")
						deployK8sClient := setupClientFactoryMock(ctx, k8sClient, mockClientFactory, namespacedName, secondK8sClient)

						_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
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

					It("should update the deployed resource (yaml) if the checksum changes", func() {
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
						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(io.NopCloser(strings.NewReader(webserver)), nil)
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
						deployK8sClient := setupClientFactoryMock(ctx, k8sClient, mockClientFactory, namespacedName, secondK8sClient)

						_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
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
						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(io.NopCloser(strings.NewReader(yamlDataUpdated)), nil)
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

						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(io.NopCloser(strings.NewReader(yaml)), nil)

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()
						deployK8sClient := setupClientFactoryMock(ctx, k8sClient, mockClientFactory, namespacedName, secondK8sClient)
						// Run reconciliation
						_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
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

					It("should reconcile resources in to its namespace if a namespace is in the artifact", func() {
						By("setting up the mock client to return a YAML with a namespace and resources in it")
						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(bytes.NewReader([]byte(`
apiVersion: v1
kind: Namespace
metadata:
  name: test-namespace
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config1
  namespace: test-namespace
---
apiVersion: v1
kind: Namespace
metadata:
  name: test-namespace2
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: config2
  namespace: test-namespace2
`)), nil)

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						By("reconciling the resource on the destination server")
						deployK8sClient := setupClientFactoryMock(ctx, k8sClient, mockClientFactory, namespacedName, secondK8sClient)

						_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						created1 := &v1.ConfigMap{}
						namespace := &v1.Namespace{}
						err = deployK8sClient.Get(ctx, types.NamespacedName{Name: "test-namespace"}, namespace)
						Expect(err).NotTo(HaveOccurred())
						Expect(namespace.Name).To(Equal("test-namespace"))
						err = deployK8sClient.Get(ctx, types.NamespacedName{Name: "config1", Namespace: "test-namespace"}, created1)
						Expect(err).NotTo(HaveOccurred())
						Expect(created1.Name).To(Equal("config1"))
						Expect(created1.Namespace).To(Equal("test-namespace"))

						created2 := &v1.ConfigMap{}
						err = deployK8sClient.Get(ctx, types.NamespacedName{Name: "config2", Namespace: "test-namespace2"}, created2)
						Expect(err).NotTo(HaveOccurred())
						Expect(created2.Name).To(Equal("config2"))
						Expect(created2.Namespace).To(Equal("test-namespace2"))
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
						mockClient.EXPECT().
							DownloadArtifact(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
							Return(io.NopCloser(strings.NewReader(yamlData)), nil)

						mockRESTMapper.EXPECT().
							RESTMapping(schema.GroupKind{Group: "", Kind: "ConfigMap"}, "v1").
							Return(&meta.RESTMapping{
								Resource: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"},
								Scope:    meta.RESTScopeNamespace,
							}, nil).AnyTimes()

						By("reconciling the resource on the destination server")
						deployK8sClient := setupClientFactoryMock(ctx, k8sClient, mockClientFactory, namespacedName, secondK8sClient)

						// First reconciliation - should create resource-old
						_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						// resource-old should exist
						Eventually(func() error {
							res := &unstructured.Unstructured{}
							res.SetAPIVersion("v1")
							res.SetKind("ConfigMap")
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-old", Namespace: namespace}, res)
						}).Should(Succeed())

						By("reconciling the resource on the destination server again with a new name")
						yamlDataNew := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: resource-new
  namespace: ` + namespace + `
`
						mockClient.EXPECT().
							DownloadArtifact(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
							Return(io.NopCloser(strings.NewReader(yamlDataNew)), nil)

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
									"managed-by":    infrahubOperator,
									OwnerAnnotation: resourceName,
								},
							},
							Data: map[string]string{
								"key1": "value1",
							},
						}
						deployK8sClient := setupClientFactoryMock(ctx, k8sClient, mockClientFactory, namespacedName, secondK8sClient)
						Expect(deployK8sClient.Create(ctx, existingConfigMap)).To(Succeed())

						defer (func() {
							err := deployK8sClient.Delete(ctx, existingConfigMap)
							Expect(err).NotTo(HaveOccurred())
						})()

						By("setting up the mock client to return a YAML with an updated ConfigMap")
						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(bytes.NewReader([]byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: example-config
data:
  key1: updated-value
`)), nil)

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						By("reconciling the resource on the destination server")
						_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
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

						deployK8sClient := setupClientFactoryMock(ctx, k8sClient, mockClientFactory, namespacedName, secondK8sClient)

						Expect(deployK8sClient.Create(ctx, existingConfigMap)).To(Succeed())

						defer (func() {
							err := deployK8sClient.Delete(ctx, existingConfigMap)
							Expect(err).NotTo(HaveOccurred())
						})()

						By("setting up the mock client to return a YAML with an updated ConfigMap")
						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(bytes.NewReader([]byte(`
apiVersion: v1
kind: ConfigMap
metadata:
  name: example-config
data:
  key1: updated-value
`)), nil)

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						By("reconciling the resource on the destination server")
						_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
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
						mockClient.EXPECT().
							DownloadArtifact(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
							Return(io.NopCloser(strings.NewReader(yamlData)), nil)

						mockRESTMapper.EXPECT().
							RESTMapping(schema.GroupKind{Group: "", Kind: "ConfigMap"}, "v1").
							Return(&meta.RESTMapping{
								Resource: schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"},
								Scope:    meta.RESTScopeNamespace,
							}, nil).AnyTimes()

						deployK8sClient := setupClientFactoryMock(ctx, k8sClient, mockClientFactory, namespacedName, secondK8sClient)

						By("reconciling the resource on the destination server")
						_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
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
						yamlDataNew := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: resource-a
  namespace: ` + namespace + `
`
						mockClient.EXPECT().
							DownloadArtifact(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
							Return(io.NopCloser(strings.NewReader(yamlDataNew)), nil)

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
						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(io.NopCloser(strings.NewReader(yamlData)), nil)

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()
						By("reconciling the resource on the destination server")
						deployK8sClient := setupClientFactoryMock(ctx, k8sClient, mockClientFactory, namespacedName, secondK8sClient)

						_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: namespacedName})
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
						yaml := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: resource-b
  namespace: ` + namespace + `
data:
  key1: value1
`
						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(io.NopCloser(strings.NewReader(yaml)), nil)

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
						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(io.NopCloser(strings.NewReader(yamlData)), nil)

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						By("reconciling the resource on the destination server")
						deployK8sClient := setupClientFactoryMock(ctx, k8sClient, mockClientFactory, namespacedName, secondK8sClient)

						_, err := reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: namespacedName})
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
						yaml := `
apiVersion: v1
kind: ConfigMap
metadata:
  name: resource-b
  namespace: ` + namespace + `
data:
  key1: value1
`
						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(io.NopCloser(strings.NewReader(yaml)), nil)

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

					It("should handle it grasefully if the managed resource is already managed by another instance of inrfrahubResource (managed by two infrahubResources)", func() {
						By("setting up the mock client to return a YAML with resources a in a namespace")
						infrahubRes := &infrahubv1alpha1.InfrahubResource{}
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
						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(io.NopCloser(strings.NewReader(yamlData)), nil)
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
						deployK8sClient := setupClientFactoryMock(ctx, k8sClient, mockClientFactory, namespacedName, secondK8sClient)

						_, err = reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						Eventually(func() error {
							cm := &v1.ConfigMap{}
							return deployK8sClient.Get(ctx, types.NamespacedName{Name: "resource-a", Namespace: "test"}, cm)
						}).Should(Succeed())

						By("creating a new InfrahubResource that manages the same resource")
						newInfrahubRes := &infrahubv1alpha1.InfrahubResource{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "new-resource",
								Namespace: namespace,
							},
							Spec: infrahubv1alpha1.InfrahubResourceSpec{
								Source: infrahubv1alpha1.InfrahubSyncSource{
									InfrahubAPIURL: apiURL,
									TargetBranch:   targetBranche,
									TargetDate:     targetDate,
									ArtifactName:   artifactName,
								},
								Destination: infrahubv1alpha1.InfrahubSyncDestination{
									Server:    destinationServer,
									Namespace: namespace,
								},
								IDs: infrahubv1alpha1.InfrahubResourceIDs{
									ArtifactID: artifactID,
									Checksum:   checksum,
									StorageID:  storageID,
								},
							},
						}

						Expect(k8sClient.Create(ctx, newInfrahubRes)).To(Succeed())
						By("reconciling the same namespace and an other resource in the second InfrahubResource")
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
						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(io.NopCloser(strings.NewReader(yamlData2)), nil)

						deployK8sClient = setupClientFactoryMock(ctx, k8sClient, mockClientFactory, namespacedName, secondK8sClient)

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
						updatedNewInfrahubRes := &infrahubv1alpha1.InfrahubResource{}
						err = k8sClient.Get(ctx, types.NamespacedName{Name: "new-resource", Namespace: namespace}, updatedNewInfrahubRes)
						Expect(err).NotTo(HaveOccurred())

						// Expect the updated newInfrahubRes to have a warning in Status.LastError
						Expect(updatedNewInfrahubRes.Status.LastError).To(ContainSubstring(fmt.Sprintf(
							"Warning: resource is already managed by infrahubResource: %s", infrahubRes.Name,
						)))

						By("Deleting the new infrahubResources again / namespace should stay")
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

						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(io.NopCloser(strings.NewReader("")), nil)
						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						By("reconciling the first infrahubResource on the destination server again with empty yaml")
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

				})
				Context("When the infrahubresource is deleted", func() {
					It("should ignore not found error of a resource", func() {
						By("calling the reconcile function on a non-existent resource")
						result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: "nonexistent", Namespace: "default"}})

						// Expect no error and empty result (reconciliation ended)
						Expect(err).NotTo(HaveOccurred())
						Expect(result).To(Equal(reconcile.Result{}))
					})

					It("should not delete the managed resource if it was overwritten before finalizer handling", func() {
						By("setting up the mock client to return a YAML with resources a and b")
						res := &infrahubv1alpha1.InfrahubResource{}
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
						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(io.NopCloser(strings.NewReader(yamlData)), nil)

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						deployK8sClient := setupClientFactoryMock(ctx, k8sClient, mockClientFactory, namespacedName, secondK8sClient)

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

						By("deleting the InfrahubResource and reconciling again")
						Expect(k8sClient.Delete(ctx, res)).To(Succeed())

						_, err = reconciler.Reconcile(ctx, ctrl.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						// res should be deleted
						Eventually(func() error {
							res := &infrahubv1alpha1.InfrahubResource{}
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

					It("should return error if DownloadArtifact fails", func() {
						By("setting up the mock client to return an error and reconcile")
						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(nil, fmt.Errorf("download error"))

						setupClientFactoryMock(ctx, k8sClient, mockClientFactory, namespacedName, secondK8sClient)

						_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("download error"))

						// Check that the InfrahubResource status is updated to failed
						instance := &infrahubv1alpha1.InfrahubResource{}
						err = k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						Expect(instance.Status.DeployState).To(Equal(infrahubv1alpha1.StateFailed))
						Expect(instance.Status.LastError).To(Equal("download error"))
					})

					It("should return error if decoding fails", func() {
						By("setting up the mock client to return invalid YAML and reconcile")
						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(bytes.NewReader([]byte("\x00\x00\x00invalid")), nil)
						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						setupClientFactoryMock(ctx, k8sClient, mockClientFactory, namespacedName, secondK8sClient)

						_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("decode artifact: error converting YAML to JSON: yaml"))
					})

					It("should return error if RESTMapping fails", func() {
						By("setting up the mock client to return a valid YAML but RESTMapping fails")
						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(bytes.NewReader([]byte(`{"apiVersion":"v1","kind":"ConfigMap","metadata":{"name":"test"}}`)), nil)

						mockRESTMapper.EXPECT().
							RESTMapping(schema.GroupKind{Group: "", Kind: "ConfigMap"}, "v1").
							Return(nil, fmt.Errorf("REST mapping error"))

						setupClientFactoryMock(ctx, k8sClient, mockClientFactory, namespacedName, secondK8sClient)

						_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("REST mapping error"))
					})

					It("should return an error when resource creation fails", func() {
						By("setting up the mock client to return a valid YAML and simulate a failure during resource creation")
						reconciler := &InfrahubResourceReconciler{
							Client:         failingK8sClient,
							Scheme:         k8sClient.Scheme(),
							InfrahubClient: mockClient,
							RESTMapper:     mockRESTMapper,
							ClientFactory:  mockClientFactory,
						}

						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(bytes.NewReader([]byte(`{
						"apiVersion": "v1",
						"kind": "ConfigMap",
						"metadata": {
							"name": "example-config",
							"namespace": "default"
						},
						"data": {
							"key1": "value1"
						}
					}`)), nil)
						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						// Act
						if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
							failingClient.FailingMethod = "Create"
						}

						setupClientFactoryMock(ctx, failingK8sClient, mockClientFactory, namespacedName, failingK8sClient)

						result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})

						// Assert: Check for the expected error
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("simulated failure: Create - &{map[apiVersion:v1 data:map[key1:value1]"))
						// Assert: Ensure the state of the InfrahubResource is marked as failed
						instance := &infrahubv1alpha1.InfrahubResource{}
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
									"managed-by":    infrahubOperator,
									OwnerAnnotation: resourceName,
								},
							},
							Data: map[string]string{
								"key1": "value1",
							},
						}

						deployK8sClient := setupClientFactoryMock(ctx, failingK8sClient, mockClientFactory, namespacedName, failingK8sClient)

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
						instance := &infrahubv1alpha1.InfrahubResource{}
						err := k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						instance.SetFinalizers([]string{FinalizerName})
						Expect(k8sClient.Update(ctx, instance)).To(Succeed())

						// Mock InfrahubClient behavior
						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(bytes.NewReader([]byte(`{
							"apiVersion": "v1", 
							"kind": "ConfigMap", 
							"metadata": {
								"name": "example-config",
								"namespace": "default"
							},
							"data": {
								"key1": "updated-value"
							}
						}`)), nil)

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						reconciler := &InfrahubResourceReconciler{
							Client:         failingK8sClient,
							Scheme:         k8sClient.Scheme(),
							InfrahubClient: mockClient,
							RESTMapper:     mockRESTMapper,
							ClientFactory:  mockClientFactory,
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

					It("should return an error when patching the state of infrahubResource with finalizer fails", func() {
						By("creating an existing ConfigMap and simulating a failure during Patch (State)")
						reconciler := &InfrahubResourceReconciler{
							Client:         failingK8sClient,
							Scheme:         k8sClient.Scheme(),
							InfrahubClient: mockClient,
							RESTMapper:     mockRESTMapper,
							ClientFactory:  mockClientFactory,
						}
						if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
							failingClient.FailingMethod = "Patch"
						}
						setupClientFactoryMock(ctx, failingK8sClient, mockClientFactory, namespacedName, failingK8sClient)

						_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).To(HaveOccurred())
						Expect(err.Error()).To(ContainSubstring("simulated failure: Patch - &{{ } {test-resource"))

						if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
							failingClient.FailingMethod = ""
						}
					})

					It("should return an error when deleting an existing managed resource fails and mark the resource stail", func() {
						By("creating an existing ConfigMap by reconcyling it and simulating a failure during update")
						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(bytes.NewReader([]byte(`{
							"apiVersion": "v1", 
							"kind": "ConfigMap", 
							"metadata": {
								"name": "example-config",
								"namespace": "default"
							},
							"data": {
								"key1": "value1"
							}
						}`)), nil)

						mockRESTMapper.EXPECT().
							RESTMapping(gomock.Any(), gomock.Any()).
							Return(&meta.RESTMapping{Scope: meta.RESTScopeNamespace}, nil).
							AnyTimes()

						// Set up the initial resource in the cluster
						reconciler := &InfrahubResourceReconciler{
							Client:         failingK8sClient,
							Scheme:         k8sClient.Scheme(),
							InfrahubClient: mockClient,
							RESTMapper:     mockRESTMapper,
							ClientFactory:  mockClientFactory,
						}
						deployK8sClient := setupClientFactoryMock(ctx, failingK8sClient, mockClientFactory, namespacedName, failingK8sClient)
						_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
						Expect(err).NotTo(HaveOccurred())

						By("updating the existing ConfigMap and recycling it with Delete failure")
						mockClient.EXPECT().
							DownloadArtifact(apiURL, artifactID, targetBranche, targetDate).
							Return(bytes.NewReader([]byte(`{
							"apiVersion": "v1", 
							"kind": "ConfigMap", 
							"metadata": {
								"name": "example-config2",
								"namespace": "default"
							},
							"data": {
								"key1": "updated-value"
							}
						}`)), nil)

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

						instance := &infrahubv1alpha1.InfrahubResource{}
						// Verify that the DeployState is set to StateStale
						err = k8sClient.Get(ctx, namespacedName, instance)
						Expect(err).NotTo(HaveOccurred())
						Expect(instance.Status.DeployState).To(Equal(infrahubv1alpha1.StateStale))

					})

					It("should return error if Get fails with unexpected error", func() {
						By("setting up the mock client to return a valid YAML and simulate a failure during Get")
						reconciler := &InfrahubResourceReconciler{
							Client:         failingK8sClient,
							Scheme:         nil, // Not needed for this test
							InfrahubClient: mockClient,
							RESTMapper:     mockRESTMapper,
							ClientFactory:  mockClientFactory,
						}
						if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
							failingClient.FailingMethod = "Get"
						}
						setupClientFactoryMock(ctx, k8sClient, mockClientFactory, namespacedName, secondK8sClient)

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

var _ = Describe("InfrahubResourceReconciler SetupWithManager", func() {
	var (
		mgr        manager.Manager
		reconciler *InfrahubResourceReconciler
		mockClient *mock.MockInfrahubClient
		mockCtrl   *gomock.Controller
	)

	BeforeEach(func() {
		var err error
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = mock.NewMockInfrahubClient(mockCtrl)

		mgr, err = ctrl.NewManager(cfg, ctrl.Options{
			Scheme: k8sClient.Scheme(),
		})
		Expect(err).NotTo(HaveOccurred())

		reconciler = &InfrahubResourceReconciler{
			Client:         mgr.GetClient(),
			Scheme:         mgr.GetScheme(),
			InfrahubClient: mockClient,
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should set up the controller with the manager successfully", func() {
		err := reconciler.SetupWithManager(mgr)
		Expect(err).NotTo(HaveOccurred())
	})
})
