MAKEFLAGS += --warn-undefined-variables
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail
.DEFAULT_GOAL := build

clean:
	rm -f .build-builder
	rm -rf build/

# builds the build container, which contains everything we need
# to build and test cbc-pillowfight, but that we don't want to ship
# because the final image weighs in at close to 1GB
.build-builder:
	docker build -t="cbc-benchmark-builder" -f=Dockerfile-builder .
	@touch .build-builder

# grabs the binary and shared lib we need from the builder image
# and drops them into the much leaner production deployment image
build: .build-builder
	mkdir -p build
	docker cp $(shell docker create cbc-benchmark-builder):/libcouchbase/build/lib/libcouchbase.so.2.0.22 - > build/libcouchbase.so.2.0.22
	docker cp $(shell docker create cbc-benchmark-builder):/libcouchbase/build/bin/cbc-pillowfight - > build/cbc-pillowfight
cbc-benchmark-builder):/libcouchbase/build/bin/cbc-n1qlback - > build/cbc-n1qlback
	docker build -t="0x74696d/cbc-benchmark" -f=Dockerfile .

# push our image to the public registry
ship: build
	docker push 0x74696d/cbc-benchmark
