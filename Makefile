# Copyright 2026 Amiasys Corporation and/or its affiliates. All rights reserved.

PACKAGE := asn-service-api/v26
.PHONY: code-inspect check-tag

code-inspect:
	goimports -w -local "$(PACKAGE)" .
	go fmt ./...
	errcheck ./...
	go vet ./...
	staticcheck ./...
	golangci-lint run
	@echo "code-inspection completed"

# Release policy: only X.Y.0 is tagged (see README.md, "Versioning").
check-tag:
	@test -n "$(TAG)" || { echo "usage: make check-tag TAG=vX.Y.0" >&2; exit 1; }
	@echo "$(TAG)" | grep -Eq '^v[0-9]+\.[0-9]+\.0$$' || \
		{ echo "ERROR: $(TAG) is not vX.Y.0; asn-service-api has no patch releases" >&2; exit 1; }
	@! git rev-parse -q --verify "refs/tags/$(TAG)" >/dev/null || \
		{ echo "ERROR: tag $(TAG) already exists" >&2; exit 1; }
	@echo "$(TAG) ok"
