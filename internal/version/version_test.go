package version

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVersion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Version Suite")
}

var _ = Describe("Version", func() {
	Describe("GetVersion", func() {
		It("returns default version", func() {
			Expect(GetVersion()).To(Equal("v0.0.0"))
		})

		It("returns custom version when set", func() {
			oldVersion := VERSION
			VERSION = "v1.2.3"
			defer func() { VERSION = oldVersion }()

			Expect(GetVersion()).To(Equal("v1.2.3"))
		})
	})

	Describe("GIT_SHA", func() {
		It("has default value", func() {
			Expect(GIT_SHA).To(Equal("unknown"))
		})
	})
})
