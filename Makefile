# Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

PACKAGE := asn-service-api/v25
.PHONY: code-inspect

code-inspect:
	goimports -w -local "$(PACKAGE)" .
	go fmt ./...
	errcheck ./...
	go vet ./...
	staticcheck ./...
	golangci-lint run
	@echo "code-inspection completed"
