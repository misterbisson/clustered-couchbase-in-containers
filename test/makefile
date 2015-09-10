MAKEFLAGS += --warn-undefined-variables
SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail
.DEFAULT_GOAL := build

ROOT := $(shell pwd)

clean:
	# rm -f .build-builder
	rm -rf build

# builds the build container, which contains everything we need
# to build and test cbc-benchmark, but that we don't want to ship
# because the final image weighs in at close to 1GB
.build-builder:
	docker build -t=golang-alpine -f=Dockerfile-builder .
	@touch .build-builder

# grabs the binary and shared lib we need from the builder image
# and drops them into the much leaner production deployment image
build: .build-builder
	mkdir -p build
	docker run -it --name cbc-build \
		-e CGO_ENABLED=0 \
		-e GOPATH=/cbc-benchmark/.godeps:/cbc-benchmark \
		-v ${ROOT}:/cbc-benchmark -w /cbc-benchmark \
		golang-alpine go build -a -o /cbc-benchmark/build/cbc-benchmark main \
		&& docker wait cbc-build
	@docker kill cbc-build || true \
		&& docker rm cbc-build || true
	chmod +x ${ROOT}/build/cbc-benchmark
	cp ${ROOT}/Dockerfile ${ROOT}/build/
	cd ${ROOT}/build && docker build -t 0x74696d/cbc-benchmark .

# push our image to the public registry
ship: build
	docker push 0x74696d/cbc-benchmark
