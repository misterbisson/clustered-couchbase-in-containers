#!/bin/bash
# Benchmarks, for fun and sparkles

VERSION=${1:-4}
PREFIX=ccic${VERSION}
NUM=${2:-2}

HOSTS=()
CONTAINERS=()
for container in $(docker ps | awk -F' +' '/ '${PREFIX}'_couchbase'${VERSION}'_[0-9+]/{print $NF}')
do
    CONTAINERS=($container)
    HOSTS+=($(docker inspect $container | json -a NetworkSettings.IPAddress))
done
HOSTSTRING=$(IFS="," ; echo "${HOSTS[*]}")

echo Running benchmark vs $HOSTSTRING...

for i in $(seq $NUM)
do
    docker run -d \
           --name ${PREFIX}_client_$i \
           0x74696d/pillowfight \
           cbc-pillowfight \
           --spec couchbase://${HOSTSTRING}/benchmark \
           --json --template id,0,200 -m 1 -M 1
done
