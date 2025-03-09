package xxhash32_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestXxhash32(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "XXHash32 Suite")
}
