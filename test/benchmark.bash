#!/bin/bash
# Benchmarks, for fun and sparkles

HOSTS=()
CONTAINERS=()
for container in $(docker ps | awk -F' +' '/ccic_couchbase/{print $NF}')
do
    CONTAINERS=($container)
    HOSTS+=($(docker inspect $container | json -a NetworkSettings.IPAddress))
done
HOSTSTRING=$(IFS="," ; echo "${HOSTS[*]}")

echo Running benchmark vs $HOSTSTRING...

for i in $(seq 2)
do
    docker run -d \
           --name ccic_client_$i \
           0x74696d/pillowfight \
           cbc-pillowfight \
           --spec couchbase://${HOSTSTRING}/benchmark \
           --json --template id,0,200 -m 1 -M 1
done
