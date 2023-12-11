test-example:
	@go run github.com/onsi/ginkgo/v2/ginkgo -v ./example/testing

test:
	@go run github.com/onsi/ginkgo/v2/ginkgo -v -p --race ./tests
