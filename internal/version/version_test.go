package version

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFactory(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Version Test Suite")
}

var _ = Describe("Version tests", func() {
	It("Should true be true", func() {
		Expect(true).To(BeTrue())
	})
})
