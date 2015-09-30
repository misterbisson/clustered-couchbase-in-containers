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

# ------------------------------------------------------
# docker-machine / swarm tasks for running on AWS
# Make sure your AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment
# variables have been set.
# You'll also need the awscli installed: https://aws.amazon.com/cli/

aws: .aws/consul .aws/machines
	@echo All machines ready. To configure your Docker client run:
	@echo 'eval "$$(docker-machine env --swarm swarm-aws-master)"'

# initialize the swarm
.aws/swarm:
	mkdir -p .aws
	docker run --rm swarm create > .aws/swarm

# Create the swarm master in your local Virtualbox environment
.aws/master: .aws/swarm
	-docker-machine rm swarm-master-aws
	docker-machine create \
		-d virtualbox \
		--swarm \
		--swarm-master \
		--swarm-discovery token://$(shell cat .aws/swarm) \
		swarm-aws-master
	touch .aws/master

# Set up and get info about the AWS environment.
# This sets up a security group that allows all communication between the nodes
# of the cluster but only ssh and 8092 (CB console) access from your current IP.
.aws/aws: .aws/vpc .aws/subnet .aws/auth-secgroup
.aws/vpc:
	aws ec2 describe-vpcs | json -a Vpcs\[0\].VpcId > .aws/vpc
.aws/subnet:
	aws ec2 describe-subnets --filters Name=availabilityZone,Values=us-east-1a | json -a Subnets[0].SubnetId > .aws/subnet
.aws/secgroup:
	aws ec2 create-security-group --vpc-id $(shell cat .aws/vpc) --group-name docker-machine | json -a GroupName > .aws/secgroup
.aws/myip:
	curl http://ifconfig.co > .aws/myip
.aws/auth-secgroup: .aws/secgroup .aws/myip
	aws ec2 authorize-security-group-ingress --group-name docker-machine \
		--protocol tcp --port 0-65535 --source-group $(shell cat .aws/secgroup)
	aws ec2 authorize-security-group-ingress --group-name docker-machine \
		--protocol udp --port 0-65535 --source-group $(shell cat .aws/secgroup)
	aws ec2 authorize-security-group-ingress --group-name docker-machine \
		--protocol tcp --port 22 --cidr $(shell cat .aws/myip)/32
	aws ec2 authorize-security-group-ingress --group-name docker-machine \
		--protocol tcp --port 8092 --cidr $(shell cat .aws/myip)/32
	aws ec2 authorize-security-group-ingress --group-name docker-machine \
		--protocol tcp --port 2376 --cidr $(shell cat .aws/myip)/32
	aws ec2 describe-security-groups --group-names docker-machine > .aws/auth-secgroup

# Create the VM where we're going to run consul and register it to swarm.
# We need to capture the machine's IP so that we can pass this value to
# docker-compose so CB nodes can find consul.
.aws/consul: .aws/master .aws/aws
	docker-machine create \
		-d amazonec2 \
		--swarm \
		--amazonec2-instance-type t2.micro \
		--amazonec2-region us-east-1 \
		--amazonec2-security-group docker-machine \
		--amazonec2-subnet-id $(shell cat .aws/subnet) \
		--amazonec2-vpc-id $(shell cat .aws/vpc) \
		--swarm-discovery token://$(shell cat .aws/swarm) \
		swarm-aws-consul
	docker-machine ip swarm-consul > .aws/consul

# Dynamically generate a list of jobs to create VMs and add these to
# our swarm
AWS_VMS := $(shell seq 1 10)
AWS_VM_JOBS := $(addprefix aws-machine,${AWS_VMS})
.aws/machines: ${AWS_VM_JOBS}
${AWS_VM_JOBS}: aws-machine%:
	docker-machine create \
		-d amazonec2 \
		--swarm \
		--amazonec2-instance-type m4.4xlarge \
		--amazonec2-region us-east-1 \
		--amazonec2-security-group docker-machine \
		--amazonec2-spot-price 1.00 \
		--amazonec2-subnet-id $(shell cat .aws/subnet) \
		--amazonec2-vpc-id $(shell cat .aws/vpc) \
		--swarm-discovery token://$(shell cat .aws/swarm) \
		swarm-aws-$*


# ------------------------------------------------------
# docker-machine / swarm tasks for running on GCE
# Make sure you have gcloud compute tools set up with a default
# project. When machines are launched your browser will launch
# asking you to verify the OAuth permissions

gce: .gce/consul # .gce/machines
	@echo All machines ready. To configure your Docker client run:
	@echo 'eval "$$(docker-machine env --swarm swarm-master-gce)"'

# initialize the swarm
.gce/swarm:
	mkdir -p .gce
	docker run --rm swarm create > .gce/swarm

# Create the swarm master in your local Virtualbox environment
.gce/master: .gce/swarm
	-docker-machine rm swarm-master-gce
	docker-machine create \
		-d virtualbox \
		--swarm \
		--swarm-master \
		--swarm-discovery token://$(shell cat .gce/swarm) \
		swarm-master-gce
	@touch .gce/master

# Set up and get info about the GCE environment.
# This sets up a firewall rule that allows communication on ssh, 8092 (CB console),
# and 8500 (consul) from your current IP.
.gce/gce: .gce/network
	gcloud info --format flattened | awk '/config.project/{print $$2}' > .gce/gce
.gce/network:
	gcloud compute firewall-rules create remote-access \
		--allow tcp:22,tcp:8092,tcp:2376,tcp:8500 --network default \
		--source-ranges $(shell curl http://ifconfig.co)/32
	@touch .gce/network

# Create the VM where we're going to run consul and register it to swarm.
# We need to capture the machine's IP so that we can pass this value to
# docker-compose so CB nodes can find consul.
.gce/consul: .gce/master .gce/gce
	docker-machine create \
		-d google \
		--swarm \
		--google-machine-type n1-standard-1 \
		--google-project $(shell cat .gce/gce) \
		--swarm-discovery token://$(shell cat .gce/swarm) \
		swarm-gce-consul
	docker-machine ip swarm-gce-consul > .gce/consul

# Dynamically generate a list of jobs to create VMs and add these to
# our swarm
GCE_VMS := $(shell seq 1 10)
GCE_VM_JOBS := $(addprefix gce-machine,${GCE_VMS})
.gce/machines: ${GCE_VM_JOBS}
${GCE_VM_JOBS}: gce-machine%:
	docker-machine create \
		-d google \
		--swarm \
		--google-machine-type n1-standard-16 \
		--google-project $(shell cat .gce/gce) \
		--swarm-discovery token://$(shell cat .gce/swarm) \
		swarm-gce-$*

# ------------------------------------------------------

.PHONY: clean .aws/machines .gce/machines ${AWS_VM_JOBS} ${GCE_VM_JOBS}
