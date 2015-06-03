#!/bin/bash

echo 'Starting Couchbase cluster'

echo
echo 'Pulling the most recent images'
docker-compose pull

echo
echo 'Starting containers'
docker-compose --project-name=ccic up -d --no-recreate --timeout=300

echo
echo -n 'Initilizing cluster.'

sleep 1.3
COUCHBASERESPONSIVE=0
while [ $COUCHBASERESPONSIVE != 1 ]; do
    echo -n '.'

    RUNNING=$(docker inspect ccic_couchbase_1 | json -a State.Running)
    if [ "$RUNNING" == "true" ]
    then
        docker exec -it ccic_couchbase_1 triton-bootstrap bootstrap benchmark
        let COUCHBASERESPONSIVE=1
    else
        sleep 1.3
    fi
done
echo

CBDASHBOARD="$(sdc-listmachines | json -aH -c "'ccic_couchbase_1' == this.name" ips.1):8091"
echo
echo 'Couchbase cluster running and bootstrapped'
echo "Dashboard: $CBDASHBOARD"
echo "username=Administrator"
echo "password=password"
`open http://$CBDASHBOARD/index.html#sec=servers`

echo
echo "Scale the cluster using the following command:"
echo "docker-compose --project-name=ccic scale couchbase=\$COUNT"
