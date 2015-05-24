#!/bin/bash

CBHOSTS=5
NODERAM=800

docker-compose pull

docker-compose --project-name=ccic up -d --no-recreate

echo 'Waiting 13 seconds for services to start'
sleep 13

echo 'Initilizing node'
docker exec -it ccic_couchbase_1 \
    couchbase-cli node-init -c 127.0.0.1:8091 -u access -p password \
        --node-init-data-path=/opt/couchbase/var/lib/couchbase/data \
        --node-init-index-path=/opt/couchbase/var/lib/couchbase/data \
        --node-init-hostname=$(sdc-listmachines | json -aH -c "'ccic_couchbase_1' == this.name" ips.0)

echo 'Initilizing cluster'
docker exec -it ccic_couchbase_1 \
    couchbase-cli cluster-init -c 127.0.0.1:8091 -u access -p password \
        --cluster-init-username=Administrator \
        --cluster-init-password=password \
        --cluster-init-port=8091 \
        --cluster-init-ramsize=$NODERAM

echo 'Initilizing bucket'
docker exec -it ccic_couchbase_1 \
    couchbase-cli bucket-create -c 127.0.01:8091 -u Administrator -p password \
       --bucket=sync_gateway \
       --bucket-type=couchbase \
       --bucket-port=11222 \
       --bucket-ramsize=$NODERAM \
       --bucket-replica=1

CBDASHBOARD="$(sdc-listmachines | json -aH -c "'ccic_couchbase_1' == this.name" ips.1):8091"
echo "Couchbase: $CBDASHBOARD"
echo "username=Administrator"
echo "password=password"
`open http://$CBDASHBOARD/index.html#sec=servers`

echo "Growing cluster to $CBHOSTS hosts"
docker-compose --project-name=ccic scale couchbase=$CBHOSTS

echo 'Waiting 13 seconds for services to start'
sleep 13

for i in `seq 1 $CBHOSTS`; \
    do \
        echo 'Initilizing node'; \
        docker exec -it ccic_couchbase_$i \
            couchbase-cli node-init -c 127.0.0.1:8091 -u access -p password \
                --node-init-data-path=/opt/couchbase/var/lib/couchbase/data \
                --node-init-index-path=/opt/couchbase/var/lib/couchbase/data \
                --node-init-hostname=$(sdc-listmachines | json -aH -c "'ccic_couchbase_$i' == this.name" ips.0); \

        echo 'Joining cluster'; \
        docker exec -it ccic_couchbase_1 \
            couchbase-cli rebalance -c 127.0.0.1:8091 -u Administrator -p password \
                --server-add=$(sdc-listmachines | json -aH -c "'ccic_couchbase_$i' == this.name" ips.0):8091 \
                --server-add-username=Administrator \
                --server-add-password=password; \
done