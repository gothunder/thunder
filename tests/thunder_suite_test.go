package tests

import (
	"os"
	"testing"

	"github.com/gothunder/thunder/tests/entInit"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	client *entInit.Client
)

func TestThunder(t *testing.T) {
	os.Setenv("TZ", "UTC")

	RegisterFailHandler(Fail)
	RunSpecs(t, "Thunder Suite")
}
