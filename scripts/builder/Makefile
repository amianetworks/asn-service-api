COMPILE_PLUGIN_IMAGE_NAME_DEV := asn-compiler-dev

compile-plugin: compile-plugin-clean
	@docker buildx build --platform linux/amd64 -f $(SERVICE_API_PATH)/scripts/builder/compiler-plugin.dockerfile -t $(COMPILE_PLUGIN_IMAGE_NAME_DEV):latest .
	@docker run -itd --platform linux/amd64 --name $(COMPILE_PLUGIN_IMAGE_NAME_DEV)-container $(COMPILE_PLUGIN_IMAGE_NAME_DEV):latest
	@mkdir -p build
	@docker cp $(COMPILE_PLUGIN_IMAGE_NAME_DEV)-container:/asn-compiler-dev/build .
	@docker stop $(COMPILE_PLUGIN_IMAGE_NAME_DEV)-container
	@docker rm $(COMPILE_PLUGIN_IMAGE_NAME_DEV)-container
	@docker rmi $(COMPILE_PLUGIN_IMAGE_NAME_DEV):latest

compile-plugin-clean:
	@rm -rf build/
