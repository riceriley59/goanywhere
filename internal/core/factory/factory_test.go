package factory

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFactory(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Factory Suite")
}

var _ = Describe("Plugin factory tests", func() {
	It("Should true be true", func() {
		Expect(true).To(BeTrue())
	})
})
