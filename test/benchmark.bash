#!/bin/bash

# create a list of our linked couchbase nodes
echo Querying for couchbase nodes...
HOSTS=()
for ip in `env | grep "CCIC_COUCHBASE_.*_PORT_8092_TCP_ADDR" | cut -d'=' -f2`
do
    HOSTS+=("$ip")
done
HOSTSTRING=$(IFS="," ; echo "${HOSTS[*]}")
NUM_HOSTS=${#HOSTS[@]}

echo Polling couchbase nodes...
CLUSTERFOUND=0
while [ "$CLUSTERFOUND" -lt ${NUM_HOSTS} ]; do
    echo -n '.'
    sleep 19

    CLUSTERFOUND=$(curl -sL http://consul:8500/v1/catalog/service/couchbase | json -aH ServiceAddress | wc -l)
done
sleep 3

CLUSTERFOUND=0
while [ $CLUSTERFOUND != 1 ]; do
    echo -n '.'

    CLUSTERIP=$(curl -sL http://consul:8500/v1/catalog/service/couchbase | json -aH ServiceAddress | head -1)
    if [ -n "$CLUSTERIP" ]
    then
        let CLUSTERFOUND=1
    else
        sleep 3
    fi
done
sleep 3


echo Running benchmark vs $HOSTSTRING...
cbc-pillowfight \
    --spec couchbase://$HOSTSTRING/benchmark \
    --min-size=800 \
    --max-size=800 \
    --num-threads=32 \
    --num-items=1000000 \
    --batch-size=5000 \
    --num-cycles=-1 \
    --set-pct=10 \
    --rate-limit=1000000 \
    --timings \
    --verbose
