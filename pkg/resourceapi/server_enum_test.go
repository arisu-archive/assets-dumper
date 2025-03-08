package resourceapi_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/arisu-archive/assets-dumper/pkg/resourceapi"
)

var _ = Describe("ServerEnum", func() {
	Describe("GetServer", func() {
		It("should recognize global server variations", func() {
			Expect(resourceapi.GetServer("g")).To(Equal(resourceapi.ServerGlobal))
			Expect(resourceapi.GetServer("gl")).To(Equal(resourceapi.ServerGlobal))
			Expect(resourceapi.GetServer("global")).To(Equal(resourceapi.ServerGlobal))
		})

		It("should recognize japan server variations", func() {
			Expect(resourceapi.GetServer("j")).To(Equal(resourceapi.ServerJapan))
			Expect(resourceapi.GetServer("jp")).To(Equal(resourceapi.ServerJapan))
			Expect(resourceapi.GetServer("japan")).To(Equal(resourceapi.ServerJapan))
		})

		It("should return unknown for invalid inputs", func() {
			Expect(resourceapi.GetServer("invalid")).To(Equal(resourceapi.ServerUnknown))
			Expect(resourceapi.GetServer("")).To(Equal(resourceapi.ServerUnknown))
		})
	})

	Describe("IsValid", func() {
		It("should validate known servers", func() {
			Expect(resourceapi.ServerGlobal.IsValid()).To(BeTrue())
			Expect(resourceapi.ServerJapan.IsValid()).To(BeTrue())
		})

		It("should invalidate unknown servers", func() {
			Expect(resourceapi.ServerUnknown.IsValid()).To(BeFalse())
			Expect(resourceapi.Server("random").IsValid()).To(BeFalse())
		})
	})
})
