package mid

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMid(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mid Suite")
}
