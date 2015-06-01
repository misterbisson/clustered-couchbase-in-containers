# Couchbase cluster in Docker containers

This is a Docker Compose file and shell script that will deploy a Couchbase cluster that can be scaled easily using `docker compose scale couchbase=$n`.

## Prep your environment

1. [Get a Joyent account](https://my.joyent.com/landing/signup/), add your SSH key, and [get a beta invite](http://joyent.com/lp/preview).
1. Install and configure the [Joyent CloudAPI CLI tools](https://apidocs.joyent.com/cloudapi/#getting-started).
1. Install the Docker CLI and Docker Compose.
1. [Configure your Docker CLI and Compose for use with Joyent](https://github.com/joyent/sdc-docker/tree/master/docs/api#the-helper-script):

```
curl -O https://raw.githubusercontent.com/joyent/sdc-docker/master/tools/sdc-docker-setup.sh && chmod +x sdc-docker-setup.sh
 ./sdc-docker-setup.sh -k us-east-3b.api.joyent.com <ACCOUNT> ~/.ssh/<PRIVATE_KEY_FILE>
```

## Easy instructions

1. [Clone](git@github.com:misterbisson/clustered-couchbase-in-containers.git) or [download](https://github.com/misterbisson/clustered-couchbase-in-containers/archive/master.zip) this repo.
1. `cd` into the cloned or downloaded directory.
1. Execute `bash start.sh` to start everything up.
    1. The script will use Docker compose to start the first containers, then execute a command to do the configuration of the first Couchbase node and start the cluster.
1. Go to the Couchbase dashboard to see the working, one-node cluster.
1. Scale the cluster using `docker-compose --project-name=ccic scale up couchbase=$n` and watch the node(s) join the cluster in the Couchbase dashboard.

## Manual instructions

The ambition is for this to work, though it is now blocked by DOCKER-409

## Start the music

```bash
docker-compose pull
docker-compose --verbose --project-name=ccic up -d --no-recreate
```

## Make it louder

```bash
docker-compose --verbose --project-name=ccic scale couchbase=5
```

---

### On each Couchbase container automatically

Inside each Couchbase container

```bash
couchbase-cli node-init -c 127.0.0.1:8091 -u access -p password \
    --node-init-data-path=/opt/couchbase/var/lib/couchbase/data \
    --node-init-index-path=/opt/couchbase/var/lib/couchbase/data \
    --node-init-hostname=$(ip addr show eth0 | grep -o '[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}')
```

1. Check Consul to determine the IP address of one or more machines in the cluster
    1. If no containers are registered yet, continue with cluster setup
1. If one or more machines are registered, continue with joining the cluster

### Consul notes

[Bootstrapping](https://www.consul.io/docs/guides/bootstrapping.html), [Consul clusters](https://www.consul.io/intro/getting-started/join.html), and the details about [adding and removing nodes](https://www.consul.io/docs/guides/servers.html). The [CLI](https://www.consul.io/docs/commands/index.html) and [HTTP](https://www.consul.io/docs/agent/http.html) API are also documented.

[Check for registered instances of a named service](https://www.consul.io/docs/agent/http/catalog.html#catalog_service)

```bash
curl -v http://165.225.190.200:8500/v1/catalog/service/couchbase | json -aH ServiceAddress
```

[Register an instance of a service](https://www.consul.io/docs/agent/http/catalog.html#catalog_register)

```bash
export MYIP=$(ip addr show eth0 | grep -o '[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}')
curl http://165.225.190.200:8500/v1/agent/service/register -d "$(printf '{"ID": "couchbase-%s","Name": "couchbase","Address": "%s"}' $MYIP $MYIP)"
```

### Cluster setup

Call the following to determine if the cluster is setup

```bash
curl -v http://165.225.190.200:8500/v1/catalog/service/couchbase | json -aH ServiceAddress
```

Inside one Couchbase container, if the cluster is not yet setup


```bash
couchbase-cli cluster-init -c 127.0.0.1:8091 -u access -p password \
    --cluster-init-username=Administrator \
    --cluster-init-password=password \
    --cluster-init-port=8091 \
    --cluster-init-ramsize=800
```

```bash
couchbase-cli bucket-create -c 127.0.01:8091 -u Administrator -p password \
   --bucket=sync_gateway \
   --bucket-type=couchbase \
   --bucket-port=11222 \
   --bucket-ramsize=800 \
   --bucket-replica=1
```

```bash
export MYIP=$(ip addr show eth0 | grep -o '[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}')
curl http://165.225.190.200:8500/v1/agent/service/register -d "$(printf '{"ID": "couchbase-%s","Name": "couchbase","Address": "%s"}' $MYIP $MYIP)"
```

### Join the cluster

Inside each Couchbase container after the first one, which presumably setup the cluster

Call the [Couchbase HTTP API](http://docs.couchbase.com/admin/admin/REST/rest-cluster-addnodes.html) to add this node. `http://192.168.129.188` below is the IP address of the first Couchbase container.

```bash
curl -i -u Administrator:password \
    http://192.168.129.188:8091/controller/addNode \
    -d "hostname=$(ip addr show eth0 | grep -o '[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}')&user=admin&password=password"
```

Then call the [CLI to rebalance](http://docs.couchbase.com/admin/admin/CLI/CBcli/cbcli-cluster-rebalance.html)

```bash
couchbase-cli rebalance -c 127.0.0.1:8091 -u Administrator -p password
```

Finally, [tell Consul](https://www.consul.io/docs/agent/http/catalog.html#catalog_register) about this newly added node

```bash
export MYIP=$(ip addr show eth0 | grep -o '[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}\.[0-9]\{1,3\}')
curl http://165.225.190.200:8500/v1/agent/service/register -d "$(printf '{"ID": "couchbase-%s","Name": "couchbase","Address": "%s"}' $MYIP $MYIP)"
```
