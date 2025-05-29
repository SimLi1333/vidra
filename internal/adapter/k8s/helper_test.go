package k8s

import (
	"context"
	"time"

	mock "github.com/infrahub-operator/vidra/internal/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("GetSortedListByLabel", func() {
	var (
		ctx        context.Context
		fakeClient client.Client
		scheme     *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.TODO()

		scheme = runtime.NewScheme()
		Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())

		fakeClient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithRuntimeObjects().
			Build()
	})

	It("should return error when no objects are found", func() {
		secretList := &corev1.SecretList{}

		err := GetSortedListByLabel(ctx, fakeClient, "infrahub-url", "missing-url", secretList)
		Expect(err).To(MatchError(ContainSubstring("no resources found with label")))
	})

	It("should return sorted objects by creation timestamp descending", func() {
		now := time.Now()

		s1 := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "older",
				Namespace:         "default",
				CreationTimestamp: metav1.NewTime(now.Add(-1 * time.Hour)),
				Labels: map[string]string{
					"infrahub-url": "https://example.com",
				},
			},
		}
		s2 := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:              "newer",
				Namespace:         "default",
				CreationTimestamp: metav1.NewTime(now),
				Labels: map[string]string{
					"infrahub-url": "https://example.com",
				},
			},
		}

		fakeClient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithRuntimeObjects(s1, s2).
			Build()

		list := &corev1.SecretList{}
		err := GetSortedListByLabel(ctx, fakeClient, "infrahub-url", "https://example.com", list)
		Expect(err).NotTo(HaveOccurred())
		Expect(list.Items).To(HaveLen(2))
		Expect(list.Items[0].Name).To(Equal("newer"))
		Expect(list.Items[1].Name).To(Equal("older"))
	})

	It("should return error if client.List fails", func() {
		if failingClient, ok := failingK8sClient.(*mock.FailingUpdateClient); ok {
			failingClient.FailingMethod = "List"
		}

		list := &corev1.SecretList{}
		err := GetSortedListByLabel(ctx, failingK8sClient, "label", "value", list)
		Expect(err).To(MatchError(ContainSubstring("failed to list resources")))
	})

	It("should return error if ExtractList fails (malformed list)", func() {
		bad := &badList{}

		err := GetSortedListByLabel(ctx, fakeClient, "label", "value", bad)
		Expect(err).To(MatchError(ContainSubstring("no kind is registered for the type")))
	})

})

type badList struct {
	metav1.TypeMeta
	metav1.ListMeta
}

func (b *badList) GetObjectKind() schema.ObjectKind  { return &b.TypeMeta }
func (b *badList) DeepCopyObject() runtime.Object    { return &badList{} } // Still implements the interface
func (b *badList) DeepCopyList() client.ObjectList   { return &badList{} }
func (b *badList) GetListMeta() metav1.ListInterface { return &b.ListMeta }
