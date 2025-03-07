package resources_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/arisu-archive/assets-dumper/pkg/resources"
)

var _ = Describe("Client", func() {
	Describe("NewClient", func() {
		It("should create a Global client", func() {
			client, err := resources.NewClient(resources.ServerGlobal)
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())
		})

		It("should create a Japan client", func() {
			client, err := resources.NewClient(resources.ServerJapan)
			Expect(err).NotTo(HaveOccurred())
			Expect(client).NotTo(BeNil())
		})

		It("should return error for unknown server", func() {
			client, err := resources.NewClient(resources.ServerUnknown)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown server"))
			Expect(client).To(BeNil())
		})
	})
})
