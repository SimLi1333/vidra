package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	infrahubv1alpha1 "github.com/infrahub-operator/vidra/api/v1alpha1"
)

var _ = Describe("MarkState and MarkStateFailed with real client", func() {
	var (
		ctx          context.Context
		vidraRes     *infrahubv1alpha1.VidraResource
		statusClient client.StatusClient
	)

	BeforeEach(func() {
		ctx = context.TODO()

		vidraRes = &infrahubv1alpha1.VidraResource{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-resource",
				Namespace: "default",
			},
		}

		// Create the resource first in the cluster/fake client
		Expect(k8sClient.Create(ctx, vidraRes)).To(Succeed())

		// Now assign the full client (which implements StatusClient)
		statusClient = k8sClient
	})
	AfterEach(func() {
		if vidraRes != nil {
			_ = k8sClient.Delete(ctx, vidraRes)
		}
	})

	It("should error if resource is nil in MarkState", func() {
		err := MarkState(ctx, statusClient, nil, func() {})
		Expect(err).To(MatchError("resource is nil"))
	})

	It("should patch the resource status successfully in MarkState", func() {
		err := MarkState(ctx, statusClient, vidraRes, func() {
			// For example, update status fields here
			vidraRes.Status.DeployState = "Succeeded"
		})
		// Since vidraRes is not created in the fake API server, this may fail unless you create it first.
		// So you can create the resource before patching, or expect error accordingly.
		Expect(err).NotTo(HaveOccurred())
	})
})
