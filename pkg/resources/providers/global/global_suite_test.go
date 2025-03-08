package global_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGlobal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Global Suite")
}
