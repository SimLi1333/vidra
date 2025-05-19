package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	infrahubv1alpha1 "github.com/simli1333/vidra/api/v1alpha1"
	"github.com/simli1333/vidra/internal/domain"
	mock "github.com/simli1333/vidra/internal/mocks"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("InfrahubSync Controller", func() {
	var (
		mockCtrl       *gomock.Controller
		mockClient     *mock.MockInfrahubClient
		ctx            context.Context
		namespacedName types.NamespacedName
		reconciler     *InfrahubSyncReconciler
	)
	const (
		resourceName      = "test-resource"
		apiURL            = "https://example.com"
		artifactName      = "test-artifact"
		namespace         = "default"
		targetBranche     = "main"
		targetDate        = "2025-01-01T00:00:00Z"
		destinationServer = "https://kubernetes.default.svc"
	)
	artifact1 := &domain.Artifact{
		ID:        "artifact-123",
		StorageID: "storage-456",
		Checksum:  "checksum-789",
	}
	artifact2 := &domain.Artifact{
		ID:        "artifact-456",
		StorageID: "storage-789",
		Checksum:  "checksum-012",
	}

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = mock.NewMockInfrahubClient(mockCtrl)
		ctx = context.Background()
		namespacedName = types.NamespacedName{
			Name: resourceName,
		}
		By("creating the secret with credentials")
		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "infrahub-credentials",
				Namespace: namespace,
			},
			Data: map[string][]byte{
				"username": []byte("test-user"),
				"password": []byte("test-pass"),
			},
		}
		// Add labels to the secret
		secret.Labels = map[string]string{
			"infrahub-api-url": "example.com",
		}
		err := k8sClient.Get(ctx, types.NamespacedName{
			Name:      "infrahub-credentials",
			Namespace: namespace,
		}, secret)
		if errors.IsNotFound(err) {
			Expect(k8sClient.Create(ctx, secret)).To(Succeed())
		} else if err == nil {
			Expect(k8sClient.Update(ctx, secret)).To(Succeed())
		} else {
			Fail(fmt.Sprintf("unexpected error while getting secret: %v", err))
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()

		By("deleting the secret for cleanup")
		secret := &v1.Secret{}
		err := k8sClient.Get(ctx, types.NamespacedName{
			Name:      "infrahub-credentials",
			Namespace: namespace,
		}, secret)
		if !errors.IsNotFound(err) {
			Expect(k8sClient.Delete(ctx, secret)).To(Succeed())
		}
		By("reseting failingK8sClient for cleanup")
		if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
			failingClient.FailingMethod = ""
		}

	})

	Context("When reconciling a resource", func() {
		BeforeEach(func() {
			By("creating the custom resource for the Kind InfrahubSync if it doesn't exist")
			instance := &infrahubv1alpha1.InfrahubSync{}
			err := k8sClient.Get(ctx, namespacedName, instance)
			if err != nil && errors.IsNotFound(err) {
				err = k8sClient.Create(ctx, &infrahubv1alpha1.InfrahubSync{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: namespace,
					},
					Spec: infrahubv1alpha1.InfrahubSyncSpec{
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
					},
				})
				Expect(err).NotTo(HaveOccurred())
			}
			reconciler = &InfrahubSyncReconciler{
				Client:         k8sClient,
				Scheme:         k8sClient.Scheme(),
				InfrahubClient: mockClient,
				RequeueAfter:   time.Minute,
				QueryName:      "test-query",
			}
		})

		AfterEach(func() {
			By("deleting the custom resource for cleanup")
			instance := &infrahubv1alpha1.InfrahubSync{}
			Expect(k8sClient.Get(ctx, namespacedName, instance)).To(Succeed())
			Expect(k8sClient.Delete(ctx, instance)).To(Succeed())
			By("deleting the infrahub resource for cleanup")
			res := &infrahubv1alpha1.InfrahubResource{}
			err := k8sClient.Get(ctx, types.NamespacedName{
				Name:      artifact1.ID,
				Namespace: namespace,
			}, res)
			if err == nil {
				Expect(k8sClient.Delete(ctx, res)).To(Succeed())
			}
			res2 := &infrahubv1alpha1.InfrahubResource{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      artifact2.ID,
				Namespace: namespace,
			}, res2)
			if err == nil {
				Expect(k8sClient.Delete(ctx, res2)).To(Succeed())
			}
		})
		Context("Once the resource exists", func() {
			It("should successfully reconcile the resource and call InfrahubClient methods if no artefacts are in infrahub", func() {
				By("setting up mock expectations")
				mockClient.EXPECT().
					Login(apiURL, "test-user", "test-pass").
					Return("mock-token", nil)

				mockClient.EXPECT().
					RunQuery("test-query", apiURL, artifactName, targetBranche, targetDate, "mock-token").
					Return(&[]domain.Artifact{}, nil)

				By("reconciling the resource")
				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})
				Expect(err).NotTo(HaveOccurred())
			})
			It("should creat the infrahubResource if the artifact id is present", func() {
				By("setting up mock expectations")
				mockClient.EXPECT().
					Login(apiURL, "test-user", "test-pass").
					Return("mock-token", nil)

				mockClient.EXPECT().
					RunQuery("test-query", apiURL, artifactName, targetBranche, targetDate, "mock-token").
					Return(&[]domain.Artifact{*artifact1}, nil)

				By("reconciling the resource")
				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})
				Expect(err).NotTo(HaveOccurred())
				infrahubResource := &infrahubv1alpha1.InfrahubResource{}
				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      artifact1.ID,
					Namespace: namespace,
				}, infrahubResource)
				Expect(err).NotTo(HaveOccurred())
				Expect(infrahubResource.Name).To(Equal(artifact1.ID))
				Expect(infrahubResource.Spec.IDs.ArtifactID).To(Equal(artifact1.ID))
				Expect(infrahubResource.Spec.IDs.StorageID).To(Equal(artifact1.StorageID))
				Expect(infrahubResource.Spec.IDs.Checksum).To(Equal(artifact1.Checksum))
				Expect(infrahubResource.Spec.Source.InfrahubAPIURL).To(Equal(apiURL))
				Expect(infrahubResource.Spec.Source.TargetBranch).To(Equal(targetBranche))
				Expect(infrahubResource.Spec.Source.TargetDate).To(Equal(targetDate))
				Expect(infrahubResource.Spec.Source.ArtifactName).To(Equal(artifactName))
				Expect(infrahubResource.Spec.Destination.Server).To(Equal(destinationServer))
				Expect(infrahubResource.Spec.Destination.Namespace).To(Equal(namespace))
				Expect(infrahubResource.Status.DeployState).To(BeEmpty())
				Expect(infrahubResource.Status.LastError).To(BeEmpty())
				infrahubSync := &infrahubv1alpha1.InfrahubSync{}
				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      resourceName,
					Namespace: namespace,
				}, infrahubSync)
				Expect(err).NotTo(HaveOccurred())
				Expect(infrahubSync.Status.SyncState).To(Equal(infrahubv1alpha1.StateSucceeded))
			})

			It("should delete the infrahubResource if the artifact id is not present", func() {
				By("setting up mock expectations")
				mockClient.EXPECT().
					Login(apiURL, "test-user", "test-pass").
					Return("mock-token", nil)
				mockClient.EXPECT().
					RunQuery("test-query", apiURL, artifactName, targetBranche, targetDate, "mock-token").
					Return(&[]domain.Artifact{*artifact1, *artifact2}, nil)

				By("reconciling the resource with two artifacts")
				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})
				Expect(err).NotTo(HaveOccurred())
				infrahubResource := &infrahubv1alpha1.InfrahubResource{}
				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      artifact1.ID,
					Namespace: namespace,
				}, infrahubResource)
				Expect(err).NotTo(HaveOccurred())
				Expect(infrahubResource.Name).To(Equal(artifact1.ID))

				infrahubResource2 := &infrahubv1alpha1.InfrahubResource{}
				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      artifact2.ID,
					Namespace: namespace,
				}, infrahubResource)
				Expect(err).NotTo(HaveOccurred())
				Expect(infrahubResource.Name).To(Equal(artifact2.ID))

				mockClient.EXPECT().
					Login(apiURL, "test-user", "test-pass").
					Return("mock-token", nil)
				mockClient.EXPECT().
					RunQuery("test-query", apiURL, artifactName, targetBranche, targetDate, "mock-token").
					Return(&[]domain.Artifact{*artifact2}, nil)

				By("reconciling the resource with one artifact")
				_, err = reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})
				Expect(err).NotTo(HaveOccurred())

				// Wait for the resource to be deleted
				Eventually(func() error {
					err := k8sClient.Get(ctx, types.NamespacedName{
						Name:      artifact1.ID,
						Namespace: namespace,
					}, infrahubResource)
					if errors.IsNotFound(err) {
						return nil
					}
					return err
				}).Should(Succeed())

				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      artifact2.ID,
					Namespace: namespace,
				}, infrahubResource2)
				Expect(err).NotTo(HaveOccurred())
				Expect(infrahubResource2.Name).To(Equal(artifact2.ID))
			})

			It("shold read the newest secret with the infrahub credentials for that url", func() {
				time.Sleep(2 * time.Second)
				By("creating the secret with credentials")
				secret := &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "infrahub-credentials2",
						Namespace: namespace,
					},
					Data: map[string][]byte{
						"username": []byte("test-user2"),
						"password": []byte("test-pass2"),
					},
				}
				// Add labels to the secret
				secret.Labels = map[string]string{
					"infrahub-api-url": "example.com",
				}
				Expect(k8sClient.Create(ctx, secret)).To(Succeed())

				defer func() {
					By("deleting the secret for cleanup")
					Expect(k8sClient.Delete(ctx, secret)).To(Succeed())
				}()

				Eventually(func(g Gomega) {
					var secretList v1.SecretList
					err := k8sClient.List(ctx, &secretList, client.InNamespace(namespace),
						client.MatchingLabels{"infrahub-api-url": "example.com"})
					g.Expect(err).NotTo(HaveOccurred())
					g.Expect(secretList.Items).To(HaveLen(2)) // one old, one new
				}, 5*time.Second, 100*time.Millisecond).Should(Succeed())

				By("setting up mock expectations")
				mockClient.EXPECT().
					Login(apiURL, "test-user2", "test-pass2").
					Return("mock-token", nil)

				mockClient.EXPECT().
					RunQuery("test-query", apiURL, artifactName, targetBranche, targetDate, "mock-token").
					Return(&[]domain.Artifact{*artifact1}, nil)

				By("reconciling the resource")
				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})
				Expect(err).NotTo(HaveOccurred())
			})

			It("should update the infrahubResource if the artifact checksum is changed", func() {
				By("setting up mock expectations")
				mockClient.EXPECT().
					Login(apiURL, "test-user", "test-pass").
					Return("mock-token", nil)
				mockClient.EXPECT().
					RunQuery("test-query", apiURL, artifactName, targetBranche, targetDate, "mock-token").
					Return(&[]domain.Artifact{*artifact1}, nil)

				By("reconciling the resource")
				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})
				Expect(err).NotTo(HaveOccurred())
				By("checking if the infrahubResource is created")
				infrahubResource := &infrahubv1alpha1.InfrahubResource{}
				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      artifact1.ID,
					Namespace: namespace,
				}, infrahubResource)
				Expect(err).NotTo(HaveOccurred())
				Expect(infrahubResource.Name).To(Equal(artifact1.ID))
				Expect(infrahubResource.Spec.IDs.StorageID).To(Equal(artifact1.StorageID))

				By("updating the infrahubResource with new checksum and storage id")
				mockClient.EXPECT().
					Login(apiURL, "test-user", "test-pass").
					Return("mock-token", nil)

				artifact1Updated := *artifact1
				artifact1Updated.Checksum = "new-checksum-123"
				artifact1Updated.StorageID = "new-storage-456"

				mockClient.EXPECT().
					RunQuery("test-query", apiURL, artifactName, targetBranche, targetDate, "mock-token").
					Return(&[]domain.Artifact{artifact1Updated}, nil)

				By("reconciling the resource with updated artifact")
				_, err = reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})
				Expect(err).NotTo(HaveOccurred())
				infrahubResource = &infrahubv1alpha1.InfrahubResource{}
				err = k8sClient.Get(ctx, types.NamespacedName{
					Name:      artifact1.ID,
					Namespace: namespace,
				}, infrahubResource)
				Expect(err).NotTo(HaveOccurred())
				Expect(infrahubResource.Name).To(Equal(artifact1.ID))
				Expect(infrahubResource.Spec.IDs.StorageID).To(Equal(artifact1Updated.StorageID))
				Expect(infrahubResource.Spec.IDs.Checksum).To(Equal(artifact1Updated.Checksum))
			})

		})

		Context("Error handling", func() {

			It("should return an error when resource creation or update fails", func() {
				By("setting up mock expectations")
				mockClient.EXPECT().
					Login(apiURL, "test-user", "test-pass").
					Return("mock-token", nil)
				mockClient.EXPECT().
					RunQuery("test-query", apiURL, artifactName, targetBranche, targetDate, "mock-token").
					Return(&[]domain.Artifact{*artifact1}, nil)

				By("reconciling the resource with failing client (Update)")
				reconciler := &InfrahubSyncReconciler{
					Client:         failingK8sClient,
					Scheme:         k8sClient.Scheme(),
					InfrahubClient: mockClient,
					RequeueAfter:   time.Minute,
					QueryName:      "test-query",
				}
				if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
					failingClient.FailingMethod = "Update"
				}

				result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to create or update InfrahubResource artifact-123: simulated failure: Update"))

				Expect(result).To(Equal(reconcile.Result{}))
				if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
					failingClient.FailingMethod = ""
				}
			})

			It("should return an error when resource deletion fails", func() {
				By("setting up mock expectations")
				mockClient.EXPECT().
					Login(apiURL, "test-user", "test-pass").
					Return("mock-token", nil)
				mockClient.EXPECT().
					RunQuery("test-query", apiURL, artifactName, targetBranche, targetDate, "mock-token").
					Return(&[]domain.Artifact{*artifact1, *artifact2}, nil)

				By("reconciling the resource with failing client ()")
				reconciler := &InfrahubSyncReconciler{
					Client:         failingK8sClient,
					Scheme:         k8sClient.Scheme(),
					InfrahubClient: mockClient,
					RequeueAfter:   time.Minute,
					QueryName:      "test-query",
				}
				_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
				Expect(err).NotTo(HaveOccurred())

				mockClient.EXPECT().
					Login(apiURL, "test-user", "test-pass").
					Return("mock-token", nil)
				mockClient.EXPECT().
					RunQuery("test-query", apiURL, artifactName, targetBranche, targetDate, "mock-token").
					Return(&[]domain.Artifact{*artifact1}, nil)

				By("reconciling the resource with failing client (Delete)")
				if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
					failingClient.FailingMethod = "Delete"
				}

				result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to delete stale InfrahubResource artifact-456: simulated failure: Delete"))

				Expect(result).To(Equal(reconcile.Result{}))
				if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
					failingClient.FailingMethod = ""
				}
			})

			It("should return an error when Listing the Secret fails", func() {
				By("reconciling the resource with failing client (List)")
				reconciler := &InfrahubSyncReconciler{
					Client:         failingK8sClient,
					Scheme:         k8sClient.Scheme(),
					InfrahubClient: mockClient,
					RequeueAfter:   time.Minute,
					QueryName:      "test-query",
				}
				if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
					failingClient.FailingMethod = "List"
				}

				result, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: namespacedName})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(fmt.Sprintf("no secret found with InfrahubAPIURL: %s, error: failed to list resources: simulated failure: List - [map[infrahub-api-url:", apiURL)))

				Expect(result).To(Equal(reconcile.Result{}))
				if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
					failingClient.FailingMethod = ""
				}
			})

			It("should return error if login fails", func() {
				mockClient.EXPECT().
					Login(apiURL, gomock.Any(), gomock.Any()).
					Return("", fmt.Errorf("login failed"))

				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("login failed"))
			})

			It("should return error if query fails", func() {
				mockClient.EXPECT().
					Login(apiURL, gomock.Any(), gomock.Any()).
					Return("mock-token", nil)

				mockClient.EXPECT().
					RunQuery("test-query", apiURL, artifactName, targetBranche, targetDate, "mock-token").
					Return(nil, fmt.Errorf("query failed"))

				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("query failed"))
			})

			It("should return error if the secret is invalid", func() {
				By("creating the secret with invalid credentials")
				secret := &v1.Secret{}
				Expect(k8sClient.Get(ctx, types.NamespacedName{
					Name:      "infrahub-credentials",
					Namespace: namespace,
				}, secret)).To(Succeed())
				Expect(k8sClient.Delete(ctx, secret)).To(Succeed())

				secret = &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "infrahub-credentials",
						Namespace: namespace,
					},
					Data: map[string][]byte{
						"username": []byte("invalid-user"),
						"password": []byte("invalid-pass"),
					},
				}
				secret.Labels = map[string]string{
					"infrahub-api-url": strings.TrimPrefix(apiURL, "https://"),
				}
				Expect(k8sClient.Create(ctx, secret)).To(Succeed())

				defer func() {
					By("deleting the secret for cleanup")
					Expect(k8sClient.Delete(ctx, secret)).To(Succeed())
				}()

				mockClient.EXPECT().
					Login(apiURL, gomock.Any(), gomock.Any()).
					Return("", fmt.Errorf("invalid secret"))

				By("reconciling the resource with invalid secret")
				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})
				// Check that the error is as expected
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid secret"))
			})

			It("should return error if the secret is not found", func() {
				By("deleting the secret to simulate the error case")
				secret := &v1.Secret{}
				Expect(k8sClient.Get(ctx, types.NamespacedName{
					Name:      "infrahub-credentials",
					Namespace: namespace,
				}, secret)).To(Succeed())
				Expect(k8sClient.Delete(ctx, secret)).To(Succeed())

				By("reconciling the resource with missing secret")
				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no secret found with InfrahubAPIURL: " + apiURL))
			})

			It("should return error if the secret has no user or password", func() {
				By("deleting the password to simulate the error case")
				secret := &v1.Secret{}
				Expect(k8sClient.Get(ctx, types.NamespacedName{
					Name:      "infrahub-credentials",
					Namespace: namespace,
				}, secret)).To(Succeed())
				delete(secret.Data, "password")
				Expect(k8sClient.Update(ctx, secret)).To(Succeed())

				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no secret found with both username and password fields"))
			})

			It("should return error if the secret is empty", func() {
				By("deleting the secret to simulate the error case")
				secret := &v1.Secret{}
				Expect(k8sClient.Get(ctx, types.NamespacedName{
					Name:      "infrahub-credentials",
					Namespace: namespace,
				}, secret)).To(Succeed())
				Expect(k8sClient.Delete(ctx, secret)).To(Succeed())

				By("creating a secret with empty data")
				emptySecret := &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "infrahub-credentials",
						Namespace: namespace,
					},
					Data: map[string][]byte{
						"username": []byte(""),
						"password": []byte(""),
					},
				}
				emptySecret.Labels = map[string]string{
					"infrahub-api-url": strings.TrimPrefix(apiURL, "https://"),
				}
				Expect(k8sClient.Create(ctx, emptySecret)).To(Succeed())

				mockClient.EXPECT().
					Login(apiURL, "", "").
					Return("", fmt.Errorf("missing username, password in the secret"))

				By("reconciling the resource with empty secret")
				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: namespacedName,
				})

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("missing username, password in the secret"))
			})

			It("should ignore the resource if it is not found", func() {
				_, err := reconciler.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: "nonexistent", Namespace: "default"}})
				Expect(err).ToNot(HaveOccurred())
			})
			It("should return error if Get fails with unexpected error", func() {
				By("reconciling the resource with failing client (Get)")
				reconciler := &InfrahubSyncReconciler{
					Client:         failingK8sClient,
					Scheme:         nil, // Not needed for this test
					InfrahubClient: mockClient,
				}
				if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
					failingClient.FailingMethod = "Get"
				}

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
var _ = Describe("InfrahubSyncReconciler SetupWithManager", func() {
	var (
		mgr        manager.Manager
		reconciler *InfrahubSyncReconciler
		mockClient *mock.MockInfrahubClient
		mockCtrl   *gomock.Controller
	)

	BeforeEach(func() {
		var err error
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = mock.NewMockInfrahubClient(mockCtrl)

		mgr, err = ctrlruntime.NewManager(cfg, ctrlruntime.Options{
			Scheme: k8sClient.Scheme(),
		})
		Expect(err).NotTo(HaveOccurred())

		reconciler = &InfrahubSyncReconciler{
			Client:         mgr.GetClient(),
			Scheme:         mgr.GetScheme(),
			InfrahubClient: mockClient,
			RequeueAfter:   time.Minute,
			QueryName:      "test-query",
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should load the configMap of the controller and set up the controller with the manager successfully", func() {
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
				"requeueAfter": "1m",
				"queryName":    "ArtifactIDs",
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
		Expect(reconciler.RequeueAfter).To(Equal(time.Minute))
		Expect(reconciler.QueryName).To(Equal("ArtifactIDs"))
	})

})
