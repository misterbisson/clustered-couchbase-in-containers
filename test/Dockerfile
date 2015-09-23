# Container for running Couchbase benchmarks
FROM node:slim

# for parsing JSON output from Consul
RUN npm install -g json

# add the script we need to run
ADD cbc-benchmark /bin/cbc-benchmark

# add the binary and lib we got from our builder container
ADD build/cbc-pillowfight /bin
ADD build/libcouchbase.so.2.0.22 /lib/x86_64-linux-gnu
RUN ln -s /lib/x86_64-linux-gnu/libcouchbase.so.2.0.22 /lib/x86_64-linux-gnu/libcouchbase.so.2
