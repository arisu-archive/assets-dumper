package resources_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/arisu-archive/assets-dumper/pkg/resources"
)

var _ = Describe("ServerEnum", func() {
	Describe("GetServer", func() {
		It("should recognize global server variations", func() {
			Expect(resources.GetServer("g")).To(Equal(resources.ServerGlobal))
			Expect(resources.GetServer("gl")).To(Equal(resources.ServerGlobal))
			Expect(resources.GetServer("global")).To(Equal(resources.ServerGlobal))
		})

		It("should recognize japan server variations", func() {
			Expect(resources.GetServer("j")).To(Equal(resources.ServerJapan))
			Expect(resources.GetServer("jp")).To(Equal(resources.ServerJapan))
			Expect(resources.GetServer("japan")).To(Equal(resources.ServerJapan))
		})

		It("should return unknown for invalid inputs", func() {
			Expect(resources.GetServer("invalid")).To(Equal(resources.ServerUnknown))
			Expect(resources.GetServer("")).To(Equal(resources.ServerUnknown))
		})
	})

	Describe("IsValid", func() {
		It("should validate known servers", func() {
			Expect(resources.ServerGlobal.IsValid()).To(BeTrue())
			Expect(resources.ServerJapan.IsValid()).To(BeTrue())
		})

		It("should invalidate unknown servers", func() {
			Expect(resources.ServerUnknown.IsValid()).To(BeFalse())
			Expect(resources.Server("random").IsValid()).To(BeFalse())
		})
	})
})
