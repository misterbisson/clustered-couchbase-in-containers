# Container for running cbc-pillowfight Couchbase stress test
FROM 		debian:wheezy

# set up couchbase apt sources
RUN apt-key adv --keyserver keyserver.ubuntu.com --recv-keys CD406E62 D9223EDA
RUN echo "deb http://packages.couchbase.com/ubuntu wheezy wheezy/main" > /etc/apt/sources.list.d/couchbase.list

# get cbc-pillowfight
RUN apt-get update -qq
RUN apt-get install -y libcouchbase2-bin
