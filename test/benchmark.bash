#!/bin/bash

# create a list of our couchbase nodes
echo Querying for couchbase nodes...
HOSTS=()
for container in `docker ps | grep ccic_couchbase | awk -F' +' '{print $NF}'`
do
    HOSTS+=(`docker inspect $container | json -a NetworkSettings.IPAddress`)
done
HOSTSTRING=$(IFS="," ; echo "${HOSTS[*]}")

echo Running benchmark vs $HOSTSTRING...

docker run -d -m 8G 0x74696d/pillowfight cbc-pillowfight \
       --spec couchbase://$HOSTSTRING/benchmark \
       --min-size=800 \
       --max-size=800 \
       --num-threads=64 \
       --num-items=1000000 \
       --batch-size=5000 \
       --num-cycles=10000 \
       --set-pct=10 \
       --rate-limit=1000000 \
       --timings \
       --verbose
