package integration

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFactory(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Integration Test Suite")
}

var _ = Describe("Integration tests", func() {
	It("Should true be true", func() {
		Expect(true).To(BeTrue())
	})
})
