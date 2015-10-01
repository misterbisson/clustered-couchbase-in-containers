#!/bin/bash

PREFIX=${PREFIX:-ccic}
COMPOSE=${COMPOSE:-docker-compose.yml}
VERSION=${VERSION:-4}
export DOCKER_CLIENT_TIMEOUT=300

echo 'Starting Couchbase cluster'

echo
echo 'Pulling the most recent images'
docker-compose -f ${COMPOSE} pull

echo
echo 'Starting containers'
# starts the couchbase containers and their deps but not the benchmark
# client container
docker-compose -f ${COMPOSE} --project-name=$PREFIX up -d --no-recreate --timeout=300 couchbase${VERSION}

echo
echo -n 'Initializing cluster.'

sleep 1.3
COUCHBASERESPONSIVE=0
while [ $COUCHBASERESPONSIVE != 1 ]; do
    echo -n '.'

    RUNNING=$(docker inspect ${PREFIX}_couchbase${VERSION}_1 | json -a State.Running)
    if [ "$RUNNING" == "true" ]
    then
        docker exec ${PREFIX}_couchbase${VERSION}_1 triton-bootstrap bootstrap benchmark
        let COUCHBASERESPONSIVE=1
    else
        sleep 1.3
    fi
done
echo

CBDASHBOARD=$(docker inspect ${PREFIX}_couchbase${VERSION}_1 | json -aH NetworkSettings.IPAddress)
echo
echo 'Couchbase cluster running and bootstrapped'
echo "Dashboard: $CBDASHBOARD"
echo "username=Administrator"
echo "password=password"
command -v open >/dev/null 2>&1 && `open http://$CBDASHBOARD:8091/index.html#sec=servers`

echo
echo 'Scaling Couchbase cluster to three nodes'
echo "docker-compose --project-name=$PREFIX scale couchbase${VERSION}=3"
docker-compose -f ${COMPOSE} --project-name=$PREFIX scale couchbase${VERSION}=3

echo
echo "Go ahead, try a lucky 7 node cluster:"
echo "docker-compose -f ${COMPOSE} --project-name=${PREFIX} scale couchbase${VERSION}=7"
