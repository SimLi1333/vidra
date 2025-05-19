package k8s_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	infrahubv1alpha1 "github.com/simli1333/vidra/api/v1alpha1"
	"github.com/simli1333/vidra/internal/adapter/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("DynamicClientFactory", func() {
	var (
		factory   *k8s.DynamicClientFactory
		ctx       context.Context
		serverURL string
		k8sClient client.Client
	)
	var namespace = "default"

	BeforeEach(func() {
		factory = k8s.NewDynamicClientFactory()
		ctx = context.Background()
		serverURL = "https://my-cluster.example.com"

		// Mock kubeconfig
		rawConfig := &clientcmdapi.Config{
			Clusters: map[string]*clientcmdapi.Cluster{
				"my-context": {
					Server: "https://my-cluster.example.com",
				},
			},
			Contexts: map[string]*clientcmdapi.Context{
				"my-context": {
					Cluster: "my-context",
				},
			},
			CurrentContext: "my-context",
		}

		configBytes, err := clientcmd.Write(*rawConfig)
		Expect(err).ToNot(HaveOccurred())

		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster-kubeconfig",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"kubeconfig": configBytes,
			},
			Type: v1.SecretTypeOpaque,
		}
		// Add labels to the secret
		secret.Labels = map[string]string{
			"infrahub-operator-kubeconf": "my-cluster.example.com",
		}
		Expect(infrahubv1alpha1.AddToScheme(scheme.Scheme)).To(Succeed())
		k8sClient = fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()

		err = k8sClient.Create(ctx, secret)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster-kubeconfig",
				Namespace: namespace,
			},
		}
		err := k8sClient.Get(ctx, client.ObjectKeyFromObject(secret), secret)
		if err == nil {
			err = k8sClient.Delete(ctx, secret)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	It("should return a cached client on second call", func() {
		c1, err := factory.GetCachedClientFor(ctx, serverURL, k8sClient)
		Expect(err).ToNot(HaveOccurred())
		Expect(c1).ToNot(BeNil())

		c2, err := factory.GetCachedClientFor(ctx, serverURL, k8sClient)
		Expect(err).ToNot(HaveOccurred())
		Expect(c2).To(Equal(c1)) // should be the same cached client
	})

	It("should fail if GetSortedListByLabel fails", func() {
		By("Deleting the secret to simulate failure")
		err := k8sClient.Delete(ctx, &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster-kubeconfig",
				Namespace: "default",
			},
		})
		Expect(err).ToNot(HaveOccurred())

		_, err = factory.GetCachedClientFor(ctx, serverURL, k8sClient)
		Expect(err).To(HaveOccurred())
		Expect(err).To(MatchError(ContainSubstring("failed to get secrets by label: no resources found with label infrahub-operator-kubeconf=my-cluster.example.com")))
	})

	It("should fail if kubeconfig is missing", func() {
		By("deleting the kubeconfig data from the secret")
		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster-kubeconfig",
				Namespace: "default",
			},
			Data: map[string][]byte{},
			Type: v1.SecretTypeOpaque,
		}
		// Add labels to the secret
		secret.Labels = map[string]string{
			"infrahub-operator-kubeconf": "my-cluster.example.com",
		}
		err := k8sClient.Update(ctx, secret)
		Expect(err).ToNot(HaveOccurred())
		By("Creating a secret without kubeconfig data")
		secret2 := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "invalid-cluster-kubeconfig",
				Namespace: "default",
			},
			Data: map[string][]byte{},
			Type: v1.SecretTypeOpaque,
		}
		// Add labels to the secret
		secret2.Labels = map[string]string{
			"infrahub-operator-kubeconf": "my-cluster.example.com",
		}
		err = k8sClient.Create(ctx, secret2)
		Expect(err).ToNot(HaveOccurred())
		Eventually(func(g Gomega) {
			var secretList v1.SecretList
			err := k8sClient.List(ctx, &secretList, client.InNamespace(namespace),
				client.MatchingLabels{"infrahub-operator-kubeconf": "my-cluster.example.com"})
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(secretList.Items).To(HaveLen(2)) // one old, one new
		}, 5*time.Second, 100*time.Millisecond).Should(Succeed())

		defer func() {
			By("Deleting the mismatched secret")
			err := k8sClient.Delete(ctx, secret)
			Expect(err).ToNot(HaveOccurred())
		}()

		_, err = factory.GetCachedClientFor(ctx, serverURL, k8sClient)
		Expect(err).To(MatchError(ContainSubstring("kubeconfig not found in any secret")))

	})

	It("should fail if no matching context is found", func() {
		By("deleting the context from the kubeconfig")
		secret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cluster-kubeconfig",
				Namespace: "default",
			},
			Data: map[string][]byte{},
			Type: v1.SecretTypeOpaque,
		}
		// Add labels to the secret
		secret.Labels = map[string]string{
			"infrahub-operator-kubeconf": "my-cluster.example.com",
		}
		err := k8sClient.Update(ctx, secret)
		Expect(err).ToNot(HaveOccurred())

		By("Creating another invalid secret with mismatched context")
		rawConfig := &clientcmdapi.Config{
			Clusters: map[string]*clientcmdapi.Cluster{
				"some-other": {Server: "https://another.example.com"},
			},
			Contexts: map[string]*clientcmdapi.Context{
				"some-other": {Cluster: "some-other"},
			},
		}

		configBytes, err := clientcmd.Write(*rawConfig)
		Expect(err).ToNot(HaveOccurred())
		invalidSecret := &v1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mismatched-cluster-kubeconfig",
				Namespace: "default",
			},
			Data: map[string][]byte{
				"kubeconfig": configBytes,
			},
			Type: v1.SecretTypeOpaque,
		}
		// Add labels to the secret
		invalidSecret.Labels = map[string]string{
			"infrahub-operator-kubeconf": "my-cluster.example.com",
		}
		err = k8sClient.Create(ctx, invalidSecret)
		Expect(err).ToNot(HaveOccurred())
		Eventually(func(g Gomega) {
			var secretList v1.SecretList
			err := k8sClient.List(ctx, &secretList, client.InNamespace(namespace),
				client.MatchingLabels{"infrahub-operator-kubeconf": "my-cluster.example.com"})
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(secretList.Items).To(HaveLen(2))
		}, 5*time.Second, 100*time.Millisecond).Should(Succeed())

		defer func() {
			By("Deleting the mismatched secret")
			err = k8sClient.Delete(ctx, invalidSecret)
			Expect(err).ToNot(HaveOccurred())
		}()

		_, err = factory.GetCachedClientFor(ctx, serverURL, k8sClient)
		Expect(err).To(MatchError(ContainSubstring("no matching context found")))
	})
})
