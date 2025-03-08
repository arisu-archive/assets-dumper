package resourceapi_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestResourceAPI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ResourceAPI Suite")
}
