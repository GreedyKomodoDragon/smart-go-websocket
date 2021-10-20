package websocket

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSmartHandler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "WSDB Suite")
}
