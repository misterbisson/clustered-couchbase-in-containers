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

docker run -m 1G 0x74696d/pillowfight cbc-pillowfight \
       --spec couchbase://$HOSTSTRING/benchmark \
       --min-size=200 \
       --max-size=200 \
       --num-threads=10 \
       --num-items=1000 \
       --batch-size=500 \
       --num-cycles=10000 \
       --set-pct=10 \
       --rate-limit=500000 \
       --timings
