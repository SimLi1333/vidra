package infrahub

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInfrahub(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Infrahub Adapter Suite")
}
