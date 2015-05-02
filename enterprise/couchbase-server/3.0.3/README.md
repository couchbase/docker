
This README will guide you through running Couchbase Server under Docker.

# Background Information

## Networking

Couchbase Server communicates on a number of different ports (see the [Couchbase Server documentation](http://docs.couchbase.com/admin/admin/Install/install-networkPorts.html "Network ports page on Couchbase Server documentation")). It also is not generally supported for nodes in a cluster to be behind any kind of NAT. For these reasons, Docker's default networking configuration is not ideally suited to Couchbase Server deployments.

There are several deployment scenarios which this Docker image can easily support. These will be detailed below, along with recommended network arrangements for each.

## Volumes

A Couchbase Server Docker container will write all persistent and node-specific data under the directory `/opt/couchbase/var`. We recommend mapping this directory to a directory on the host filesystem (using the `-v` option to `docker run`) for the following reasons:

* Persistence. Storing `/opt/couchbase/var` outside the container allows you to delete the container and re-create it later. You can even update to a container running a later point release of Couchbase Server without losing your data.
* Performance. In a standard Docker environment using a union filesystem, leaving `/opt/couchbase/var` "inside" the container will result in some amount of performance degradation.

*SELinux workaround*

If you have SELinux enabled, mounting host volumes in a container requires an extra step.  Assuming you are mounting the `~/couchbase` directory on the host filesystem, you will need to run the following command once before running your first container on that host:

```
mkdir ~/couchbase && chcon -Rt svirt_sandbox_file_t ~/couchbase
```

## Ulimits

Couchbase normally expects the following changes to ulimits:

```
ulimit -n 40960        # nofile: max number of open files
ulimit -c unlimited    # core: max core file size
ulimit -l unlimited    # memlock: maximum locked-in-memory address space
```

These parameters come into play when running under heavy load, so if you are just doing light testing and development, you can ignore these.

In order to set the ulimits in your container, you will need to run Couchbase Docker containers with the following additional `--ulimit` flags:

```
docker run -d --ulimit nofile=40960:40960 --ulimit core=100000000:100000000 --ulimit memlock=100000000:100000000 couchbase/server
```

Since `unlimited` is not supported as a value, it sets the core and memlock values to 100 GB.  If your system has more than 100 GB RAM, you will want to increase this value to match the available RAM on the system.

NOTE: the `--ulimit` flags only work on Docker 1.6 or later.

# Common Deployment Scenarios

## Single container on single host (easy)

This is a quick way to try out Couchbase Server on your own machine with no installation overhead - download and run! In this case, any networking configuration will work; the only real requirement is that port 8091 be exposed so that you can access the Couchbase Admin Console.

Start the container:

```
docker run -d -v ~/couchbase:/opt/couchbase/var -p 8091:8091 couchbase/server
```

Resulting container architecture:

```
┌───────────────────────┐                                                      
│   Host OS (Linux)     │                                                      
│  ┌─────────────────┐  │                                                      
│  │  Container OS   │  │                                                      
│  │    (CentOS)     │  │                                                      
│  │  ┌───────────┐  │  │                                                      
│  │  │ Couchbase │  │  │                                                      
│  │  │  Server   │  │  │                                                      
│  │  └───────────┘  │  │                                                      
│  └─────────────────┘  │                                                      
└───────────────────────┘                                                      
```

## Multiple hosts in a single datacenter, single container on each host (easy)

This is a typical Couchbase Server cluster, where each node runs on a dedicated host. We assume that the datacenter LAN configuration allows each host in the cluster to see each other host via known IPs.

In this case, the most efficient way to run your cluster in Docker is to use the host's own networking stack, by running each container with the `--net=host` option. There is no need to use `-p` to "expose" any ports. Each container will use the IP address(es) of its host.

Start a container on each host via:

```
docker run -d -v ~/couchbase:/opt/couchbase/var --net=host couchbase/server
```

You can access the Couchbase Server Admin Console via port 8091 on any of the hosts.  When configuring Couchbase, you will need to use the host IP address.

In addition to being easy to set up, this is also likely to be the most performant way to deploy a Docker-based cluster as there will be no Docker-imposed networking overhead.

Resulting container architecture:

```
┌───────────────────────┐  ┌───────────────────────┐  ┌───────────────────────┐
│   Host OS (Linux)     │  │   Host OS (Linux)     │  │   Host OS (Linux)     │
│  ┌─────────────────┐  │  │  ┌─────────────────┐  │  │  ┌─────────────────┐  │
│  │  Container OS   │  │  │  │  Container OS   │  │  │  │  Container OS   │  │
│  │    (CentOS)     │  │  │  │    (CentOS)     │  │  │  │    (CentOS)     │  │
│  │  ┌───────────┐  │  │  │  │  ┌───────────┐  │  │  │  │  ┌───────────┐  │  │
│  │  │ Couchbase │  │  │  │  │  │ Couchbase │  │  │  │  │  │ Couchbase │  │  │
│  │  │  Server   │  │  │  │  │  │  Server   │  │  │  │  │  │  Server   │  │  │
│  │  └───────────┘  │  │  │  │  └───────────┘  │  │  │  │  └───────────┘  │  │
│  └─────────────────┘  │  │  └─────────────────┘  │  │  └─────────────────┘  │
└───────────────────────┘  └───────────────────────┘  └───────────────────────┘
```

## Running in container clouds with SDN (easy)

Some cloud providers, such as:

* Joyent Triton Container Cloud
* Amazon ECS
* Google Container Engine

all provide Software Defined Networking (SDN) which simplifies the networking setup required to run Couchbase Server.  We have experimented with Couchbase Server deployments on Joyent's Triton offering and have been very pleased with the performance and ease of use, so this section will be based on those experiences.

Within Joyent, a container is itself a first-class citizen; there is no "host" for the container. This is how they achieve bare-metal speeds while keeping the advantages of containerization.

**Networking**

* Each container is given an IP on an account-wide LAN.
* Every container can see every other container on these internal IP addresses.
* When configuring a Couchbase cluster, internal IP addresses should be used.
* The network infrastructure between containers is handled automatically and efficiently.
* By specifying the `-P` option to `docker run`, you can request that a container be given a public IP that is visible from the internet.  You should specify this option for at least one node in your cluster so that you can access the Admin Console, Client API ports, and so on.  It is not necessary or desirable to specify `-P` for every container.

**Volumes**

* There is no "host" for a container in Joyent. Therefore the `-v` option is not used.
* All storage must be inside a container.
* Joyent does not use a union filesystem for its Docker layer, but rather a highly efficient ZFS implementation.
* There is no performance penalty to using in-container storage.

**Persistent data and upgrades**

* Joyent does support volume links between containers. You could therefore launch two containers per node in your cluster - one simply to host the storage, and the other running Couchbase Server.  (downside: extra cost)
* Another option is to stick with a single container per node, and use [rolling upgrades](http://blog.couchbase.com/Couchbase-rolling-upgrades "Couchbase blog on rolling upgrades") when you wish to upgrade to a newer Couchbase Server version.

So the `docker run` command for nodes in Joyent becomes very easy:

```
docker run -d couchbase/server
```

Just remember to also specify `-P` for one or two nodes so you can connect to port 8091 for the Admin Console.

Resulting container architecture:

```
┌───────────────────────────────────────────────────────────────┐                         
│                        Container Cloud                        │                         
│ ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │                         
│ │  Container OS   │  │  Container OS   │  │  Container OS   │ │                         
│ │    (CentOS)     │  │    (CentOS)     │  │    (CentOS)     │ │                         
│ │ ┌─────────────┐ │  │ ┌─────────────┐ │  │ ┌─────────────┐ │ │                         
│ │ │  Couchbase  │ │  │ │  Couchbase  │ │  │ │  Couchbase  │ │ │                         
│ │ │   Server    │ │  │ │   Server    │ │  │ │   Server    │ │ │                         
│ │ │             │ │  │ │             │ │  │ │             │ │ │                         
│ │ │ private ip: │ │  │ │ private ip: │ │  │ │ private ip: │ │ │                         
│ │ │ 10.20.21.1  │ │  │ │ 10.20.21.2  │ │  │ │ 10.20.21.3  │ │ │                         
│ │ │             │ │  │ │             │ │  │ │             │ │ │                         
│ │ │ public ip:  │ │  │ │             │ │  │ │             │ │ │                         
│ │ │ 62.87.22.8  │ │  │ │             │ │  │ │             │ │ │                         
│ │ └─────────────┘ │  │ └─────────────┘ │  │ └─────────────┘ │ │                         
│ └─────────────────┘  └─────────────────┘  └─────────────────┘ │                         
└───────────────────────────────────────────────────────────────┘                         
```

## Multiple containers on a single host (medium)

* Useful for testing out a multi-node cluster on your local workstation.
* Not recommended for production use.  (the norm for a production cluster is that each node runs on dedicated hardware)
* Allows you to experiment with cluster rebalancing and failover.
* The networking is effectively the same as described the Software-Defined Network section: each container is given an internal IP address by Docker, and each of these IPs is visible to all other containers running on the same host
* Internal IPs should be used in the Admin Console when adding new nodes to the cluster
* For external access to the admin console, you should expose port 8091 of exactly one of the containers when you start it.

You can choose to mount `/opt/couchbase/var` from the host as you like. If you do so, though, remember to give each container a separate host directory!

```
docker run -d -v ~/couchbase/node1:/opt/couchbase/var couchbase/server
docker run -d -v ~/couchbase/node2:/opt/couchbase/var couchbase/server
docker run -d -v ~/couchbase/node3:/opt/couchbase/var -p 8091:8091 couchbase/server
```

Resulting container architecture:

```
┌──────────────────────────────────────────────────────────┐                   
│                     Host OS (Linux)                      │                   
│                                                          │                   
│  ┌───────────────┐ ┌───────────────┐  ┌───────────────┐  │                   
│  │ Container OS  │ │ Container OS  │  │ Container OS  │  │                   
│  │   (CentOS)    │ │   (CentOS)    │  │   (CentOS)    │  │                   
│  │ ┌───────────┐ │ │ ┌───────────┐ │  │ ┌───────────┐ │  │                   
│  │ │ Couchbase │ │ │ │ Couchbase │ │  │ │ Couchbase │ │  │                   
│  │ │  Server   │ │ │ │  Server   │ │  │ │  Server   │ │  │                   
│  │ └───────────┘ │ │ └───────────┘ │  │ └───────────┘ │  │                   
│  └───────────────┘ └───────────────┘  └───────────────┘  │                   
└──────────────────────────────────────────────────────────┘                   
```

**Setting up your Couchbase cluster**

1. After running the last `docker run` command above, get the <container_id>.  Lets call that `<node_3_container_id>`

1. Get the ip address of the node 3 container by running `docker inspect --format '{{ .NetworkSettings.IPAddress }}' <node_3_container_id>`.  Lets call that `<node_3_ip_addr>`.

1. From the host, connect to http://localhost:8091 in your browser and click the "Setup" button.

1. In the hostname field, enter `<node_3_ip_addr>`

1. Accept all default values in the setup wizard.  Choose a password that you will remember.

1. Choose the Add Servers button in the web UI

1. For the two remaining containers

    1. Get the ip address of the container by running `docker inspect --format '{{ .NetworkSettings.IPAddress }}' <node_x_container_id>`.  Lets call that `<node_x_ip_addr>`

    1. In the Server IP Address field, use `<node_x_ip_addr>` 

    1. In the password field, use the password created above.

## Multiple hosts, multiple containers per host (hard)

```
┌─────────────────────────────────────────┐  ┌─────────────────────────────────────────┐
│            Host OS (Linux)              │  │            Host OS (Linux)              │
│ ┌─────────────────┐ ┌─────────────────┐ │  │ ┌─────────────────┐ ┌─────────────────┐ │
│ │  Container OS   │ │  Container OS   │ │  │ │  Container OS   │ │  Container OS   │ │
│ │    (CentOS)     │ │    (CentOS)     │ │  │ │    (CentOS)     │ │    (CentOS)     │ │
│ │  ┌───────────┐  │ │  ┌───────────┐  │ │  │ │  ┌───────────┐  │ │  ┌───────────┐  │ │
│ │  │ Couchbase │  │ │  │ Couchbase │  │ │  │ │  │ Couchbase │  │ │  │ Couchbase │  │ │
│ │  │  Server   │  │ │  │  Server   │  │ │  │ │  │  Server   │  │ │  │  Server   │  │ │
│ │  └───────────┘  │ │  └───────────┘  │ │  │ │  └───────────┘  │ │  └───────────┘  │ │
│ └─────────────────┘ └─────────────────┘ │  │ └─────────────────┘ └─────────────────┘ │
└─────────────────────────────────────────┘  └─────────────────────────────────────────┘
```

* Difficult to achieve with plain vanilla Docker, as there is no native way to allow each container unrestricted access to the internal IPs of containers running on other hosts.
* There are software networking layers such as [Flannel](https://github.com/coreos/flannel "Flannel") and [Weave](https://github.com/weaveworks/weave "Weave"), but it is beyond the scope of this README to cover how those might be configured.
* This is not a particularly useful deployment scenario for either testing or production use, so we will simply suggest that you not try this.

