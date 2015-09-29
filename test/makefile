MAKEFLAGS += --warn-undefined-variables
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail
.DEFAULT_GOAL := build

clean:
	rm -f .build-builder
	rm -rf build/perf build/durability

# builds the build container, which contains everything we need
# to build and test cbc-pillowfight, but that we don't want to ship
# because the final image weighs in at close to 1GB
.build-builder:
	docker build -t="cbc-benchmark-builder" -f=Dockerfile-builder .
	@touch .build-builder

build: build/perf build/durability

# Copy out the libcouchbase perf-external build to a local directory
# and ADD it during a Docker build
build/perf: .build-builder
	mkdir -p build/perf
	docker cp $(shell docker create cbc-benchmark-builder):/libcouchbase/perf/libcouchbase/build/lib/libcouchbase.so.2.0.28 - > build/perf/libcouchbase.so.2.0.28
	docker cp $(shell docker create cbc-benchmark-builder):/libcouchbase/perf/libcouchbase/build/bin/cbc-pillowfight - > build/perf/cbc-pillowfight
	docker build -t="0x74696d/cbc-benchmark-perf" -f=Dockerfile-perf .

# Copy out the libcouchbase durability build to a local directory
# and ADD it during a Docker build
build/durability: .build-builder
	mkdir -p build/durability
	docker cp $(shell docker create cbc-benchmark-builder):/libcouchbase/durability/libcouchbase/build/lib/libcouchbase.so.2.0.22 - > build/durability/libcouchbase.so.2.0.22
	docker cp $(shell docker create cbc-benchmark-builder):/libcouchbase/durability/libcouchbase/build/bin/cbc-pillowfight - > build/durability/cbc-pillowfight
	docker build -t="0x74696d/cbc-benchmark-durable" -f=Dockerfile-durable .

# push our images to the public registry
ship: build
	docker push 0x74696d/cbc-benchmark-durable
	docker push 0x74696d/cbc-benchmark-perf
