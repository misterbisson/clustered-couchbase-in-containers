# Container for running cbc-pillowfight Couchbase stress test
FROM debian:jessie

# add the binary and lib we got from our builder container
ADD build/cbc-pillowfight /bin
ADD build/cbc-n1qlback /bin
ADD build/libcouchbase.so.2.0.28 /lib/x86_64-linux-gnu
RUN ln -s /lib/x86_64-linux-gnu/libcouchbase.so.2.0.28 /lib/x86_64-linux-gnu/libcouchbase.so.2
