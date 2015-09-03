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

MIN_SIZE=${MIN_SIZE:-800}
MAX_SIZE=${MAX_SIZE:-800}
THREADS=${THREADS:-32}
ITEMS=${ITEMS:-2500000}
BATCH=${BATCH:-5000}
CYCLES=${CYCLES:-10000}
SET_PCT=${SET_PCT:-10}
RATE_LIMIT=${RATE_LIMIT:-1000000}

echo Running benchmark vs $HOSTSTRING...
cbc-pillowfight \
    --spec couchbase://$HOSTSTRING/benchmark \
    --min-size=${MIN_SIZE} \
    --max-size=${MAX_SIZE} \
    --num-threads=${THREADS} \
    --num-items=${ITEMS} \
    --batch-size=${BATCH} \
    --num-cycles=${CYCLES} \
    --set-pct=${SET_PCT} \
    --rate-limit=${RATE_LIMIT} \
    --timings \
    --verbose
