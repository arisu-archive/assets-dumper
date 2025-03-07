package japan_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJapan(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Japan Suite")
}
