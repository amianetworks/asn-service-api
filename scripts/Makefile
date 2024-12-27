.PHONY: build-dev clean

IMAGE_NAME_DEV := asn-compiler-dev

compile-plugin: clean
	@docker buildx build --platform linux/amd64 -f compiler-plugin.dockerfile -t $(IMAGE_NAME_DEV):latest .
	@docker run -itd --platform linux/amd64 --name $(IMAGE_NAME_DEV)-container $(IMAGE_NAME_DEV):latest
	@mkdir -p build
	@docker cp $(IMAGE_NAME_DEV)-container:/asn-compiler-dev/controller/build ./build/controller
	@docker cp $(IMAGE_NAME_DEV)-container:/asn-compiler-dev/node/build ./build/node
	@docker stop $(IMAGE_NAME_DEV)-container
	@docker rm $(IMAGE_NAME_DEV)-container
	@docker rmi $(IMAGE_NAME_DEV):latest

clean:
	@rm -rf build/