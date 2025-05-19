/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	infrahubv1alpha1 "gitlab.ost.ch/ins-stud/sa-ba/ba-fs25-infrahub/infrahub-operator/api/v1alpha1"
	mock "gitlab.ost.ch/ins-stud/sa-ba/ba-fs25-infrahub/infrahub-operator/internal/mocks"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var secondCfg *rest.Config
var k8sClient client.Client
var secondK8sClient client.Client
var failingK8sClient client.Client
var testEnv *envtest.Environment
var secondTestEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	// First test environment
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,

		// The BinaryAssetsDirectory is only required if you want to run the tests directly
		// without call the makefile target test. If not informed it will look for the
		// default path defined in controller-runtime which is /usr/local/kubebuilder/.
		// Note that you must have the required binaries setup under the bin directory to perform
		// the tests directly. When we run make test it will be setup and used automatically.
		BinaryAssetsDirectory: filepath.Join("..", "..", "bin", "k8s",
			fmt.Sprintf("1.31.0-%s-%s", runtime.GOOS, runtime.GOARCH)),
	}
	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = infrahubv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	failingK8sClient = &mock.FailingUpdateClient{
		Client:       k8sClient,
		UpdateErrMsg: "simulated failure",
		RestMapper:   nil,
	}

	// Second isolated environment
	secondTestEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
		BinaryAssetsDirectory: filepath.Join("..", "..", "bin", "k8s",
			fmt.Sprintf("1.31.0-%s-%s", runtime.GOOS, runtime.GOARCH)),
	}
	secondCfg, err = secondTestEnv.Start()
	Expect(err).NotTo(HaveOccurred())

	err = infrahubv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	secondK8sClient, err = client.New(secondCfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())

	Expect(k8sClient).NotTo(BeNil())
	Expect(secondK8sClient).NotTo(BeNil())
	Expect(failingK8sClient).NotTo(BeNil())

})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func setupClientFactoryMock(
	ctx context.Context,
	k8sClient client.Client,
	mockClientFactory *mock.MockClientFactory,
	namespacedName types.NamespacedName,
	secondK8sClient client.Client,
) client.Client {
	infrahubResource := &infrahubv1alpha1.InfrahubResource{}
	err := k8sClient.Get(ctx, namespacedName, infrahubResource)
	Expect(err).NotTo(HaveOccurred())

	if infrahubResource.Spec.Destination.Server != "" || infrahubResource.Spec.Destination.Server == "https://kubernetes.default.svc" {
		mockClientFactory.EXPECT().
			GetCachedClientFor(ctx, infrahubResource.Spec.Destination.Server, k8sClient).
			Return(secondK8sClient, nil).AnyTimes()
		return secondK8sClient
	}

	return k8sClient
}
