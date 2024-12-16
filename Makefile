PACKAGE := github.com/amianetworks/asn-service-api/v25

.PHONY: code-cleanup dependency-update code-inspection

code-cleanup: dependency-update code-inspection

dependency-update:
	go mod tidy
	go get -u ./...
	go mod tidy
	@echo "dependency-update completed"

code-inspection:
	goimports -w -local "$(PACKAGE)" .
	go fmt ./...
	errcheck ./...
	go vet ./...
	staticcheck ./...
	golangci-lint run
	@echo "code-inspection completed"