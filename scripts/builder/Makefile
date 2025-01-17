# Copyright 2025 Amiasys Corporation and/or its affiliates. All rights reserved.

COMPILE_PLUGIN_IMAGE_NAME_DEV := asn-compiler-dev

build: build-clean
	@docker buildx build --platform linux/amd64 -f service-api/scripts/builder/compiler-plugin.dockerfile -t $(COMPILE_PLUGIN_IMAGE_NAME_DEV):latest .
	@docker run -itd --platform linux/amd64 --name $(COMPILE_PLUGIN_IMAGE_NAME_DEV)-container $(COMPILE_PLUGIN_IMAGE_NAME_DEV):latest
	@mkdir -p build
	@docker cp $(COMPILE_PLUGIN_IMAGE_NAME_DEV)-container:/asn-compiler-dev/build .
	@docker stop $(COMPILE_PLUGIN_IMAGE_NAME_DEV)-container
	@docker rm $(COMPILE_PLUGIN_IMAGE_NAME_DEV)-container
	@docker rmi $(COMPILE_PLUGIN_IMAGE_NAME_DEV):latest

build-clean:
	@rm -rf build/
