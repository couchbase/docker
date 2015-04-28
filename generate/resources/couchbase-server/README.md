
This README will guide you through running Couchbase Server under Docker.

# Background Information

## Networking

Couchbase Server communicates on a number of different ports (see the [Couchbase Server documentation](http://docs.couchbase.com/admin/admin/Install/install-networkPorts.html "Network ports page on Couchbase Server documentation")). It also is not generally supported for nodes in a cluster to be behind any kind of NAT. For these reasons, Docker's default networking configuration is not ideally suited to Couchbase Server deployments.

There are several deployment scenarios which this Docker image can easily support. These will be detailed below, along with recommended network arrangements for each.

## Volumes

A Couchbase Server Docker container will write all persistent and node-specific data under the directory `/opt/couchbase/var`. We recommend mapping this directory to a directory on the host filesystem (using the `-v` option to `docker run`) for the following reasons:

* Persistence. Storing `/opt/couchbase/var` outside the container allows you to delete the container and re-create it later. You can even update to a container running a later point release of Couchbase Server without losing your data.
* Performance. In a standard Docker environment using a union filesystem, leaving `/opt/couchbase/var` "inside" the container will result in some amount of performance degradation.

# Common Deployment Scenarios

## Single container on single host (easy)

This is a quick way to try out Couchbase Server on your own machine with no installation overhead - download and run! In this case, any networking configuration will work; the only real requirement is that port 8091 be exposed so that you can access the Couchbase Admin Console.

    docker run -d -v ~/couchbase:/opt/couchbase/var -p8091:8091 couchbase/server
    
                                                                                      
       ┌───────────────────────┐                                                      
       │  Host OS (Ubuntu...)  │                                                      
       │  ┌─────────────────┐  │                                                      
       │  │  Container OS   │  │                                                      
       │  │    (CentOS)     │  │                                                      
       │  │  ┌───────────┐  │  │                                                      
       │  │  │ Couchbase │  │  │                                                      
       │  │  │  Server   │  │  │                                                      
       │  │  └───────────┘  │  │                                                      
       │  │                 │  │                                                      
       │  └─────────────────┘  │                                                      
       │                       │                                                      
       └───────────────────────┘                                                      


## Multiple hosts in a single datacenter, single container on each host (easy)

This is a "true" Couchbase Server cluster, where each node runs on a dedicated host. We assume that the datacenter LAN configuration allows each host in the cluster to see each other host via known IPs.

In this case, the most efficient way to run your cluster in Docker is to use the host's own networking stack, by running each container with the `--net=host` option. There is no need to use `-p` to "expose" any ports. Each container will use the IP address(es) of its host.

    docker run -d -v ~/couchbase:/opt/couchbase/var --net=host couchbase/server

You can access the Couchbase Server Admin Console via port 8091 on any of the hosts.

In addition to being easy to set up, this is also likely to be the most performant way to deploy a Docker-based cluster as there will be no Docker-imposed networking overhead.

       ┌───────────────────────┐  ┌───────────────────────┐  ┌───────────────────────┐
       │  Host OS (Ubuntu...)  │  │  Host OS (Ubuntu...)  │  │  Host OS (Ubuntu...)  │
       │  ┌─────────────────┐  │  │  ┌─────────────────┐  │  │  ┌─────────────────┐  │
       │  │  Container OS   │  │  │  │  Container OS   │  │  │  │  Container OS   │  │
       │  │    (CentOS)     │  │  │  │    (CentOS)     │  │  │  │    (CentOS)     │  │
       │  │  ┌───────────┐  │  │  │  │  ┌───────────┐  │  │  │  │  ┌───────────┐  │  │
       │  │  │ Couchbase │  │  │  │  │  │ Couchbase │  │  │  │  │  │ Couchbase │  │  │
       │  │  │  Server   │  │  │  │  │  │  Server   │  │  │  │  │  │  Server   │  │  │
       │  │  └───────────┘  │  │  │  │  └───────────┘  │  │  │  │  └───────────┘  │  │
       │  │                 │  │  │  │                 │  │  │  │                 │  │
       │  └─────────────────┘  │  │  └─────────────────┘  │  │  └─────────────────┘  │
       │                       │  │                       │  │                       │
       └───────────────────────┘  └───────────────────────┘  └───────────────────────┘


## Running in environments with SDN (easy)

Some cloud providers, such as Amazon ECS and Joyent's Triton Container Cloud, provide Software Defined Networking (SDN) which simplifies the networking setup required to run Couchbase Server. We have experimented with Couchbase Server deployments on Joyent's Triton offering and have been very pleased with the performance and ease of use, so this section will be based on those experiences.

Within Joyent, a container is itself a first-class citizen; there is no "host" for the container. This is how they achieve bare-metal speeds while keeping the advantages of containerization. Each container is given an IP on an account-wide LAN. Every container can see every other container on these internal IP addresses, so when configuring the cluster, these are the IPs you should use. The network infrastructure between containers is handled automatically and efficiently.

In addition, by specifying the `-P` option to `docker run`, you can request that a container be given a public IP that is visible from the internet. You should specify this option for at least one node in your cluster so that you can access the Admin Console, Client API ports, and so on. It is not necessary or desirable to specify this for every container.

The Docker Volume story is also different for Joyent. As mentioned, there is no "host" for a container in Joyent. Therefore the `-v` option is not used. All storage must be inside a container. Fortunately Joyent does not use a union filesystem for its Docker layer, but rather a highly efficient ZFS implementation. Therefore there is no performance penalty to using in-container storage.

As for data persistence and keeping data around while upgrading the version of Couchbase Server in your container, Joyent does support volume links between containers. You could therefore launch two containers per node in your cluster - one simply to host the storage, and the other running Couchbase Server. This will significantly increase your Joyent cost, however. A better solution is to stick with a single container per node, and use [rolling upgrades](http://blog.couchbase.com/Couchbase-rolling-upgrades "Couchbase blog on rolling upgrades") when you wish to upgrade to a newer Couchbase Server version.

So the `docker run` command for nodes in Joyent becomes very easy:

    docker run -d couchbase/server

Just remember to also specify `-P` for one or two nodes so you can connect to port 8091 for the Admin Console.

                                                                                      
       ┌─────────────────────────────────────────────────────┐                        
       │                     Environment                     │                        
       │                                                     │                        
       │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │                        
       │  │  Couchbase  │  │  Couchbase  │  │  Couchbase  │  │                        
       │  │   Server    │  │   Server    │  │   Server    │  │                        
       │  │             │  │             │  │             │  │                        
       │  │ private ip: │  │ private ip: │  │ private ip: │  │                        
       │  │ 10.20.21.1  │  │ 10.20.21.2  │  │ 10.20.21.3  │  │                        
       │  │             │  │             │  │             │  │                        
       │  │ public ip:  │  │             │  │             │  │                        
       │  │ 62.87.22.8  │  │             │  │             │  │                        
       │  └─────────────┘  └─────────────┘  └─────────────┘  │                        
       └─────────────────────────────────────────────────────┘


## Multiple containers on a single host (medium)

This is not a very useful deployment scenario unless you simply want to test out a multi-node cluster on your local workstation. We would not recommend this for a production environment. Again, the norm for a production cluster is that each node runs on dedicated hardware.

Still, if you want to play around with a local cluster to watch how rebalancing, failover, and so on work, this is probably the easiest way to achieve that. Network-wise, this is effectively the same as described the Software-Defined Network section: each container is given an internal IP address by Docker, and each of these IPs is visible to all other containers running on the same host. As above, these internal IPs should be used in the Admin Console when adding new nodes to the cluster. Likewise, for external access to the admin console, you should expose port 8091 of exactly one of the containers when you start it.

You can choose to mount `/opt/couchbase/var` from the host as you like. If you do so, though, remember to give each container a separate host directory!

    docker run -d -v ~/couchbase/node1:/opt/couchbase/var couchbase/server
    docker run -d -v ~/couchbase/node2:/opt/couchbase/var couchbase/server
    docker run -d -v ~/couchbase/node3:/opt/couchbase/var -p 8091:8091 couchbase/server

                                                                                      
       ┌──────────────────────────────────────────────────────────┐                   
       │                    Host OS (Ubuntu...)                   │                   
       │                                                          │                   
       │  ┌───────────────┐ ┌───────────────┐  ┌───────────────┐  │                   
       │  │ Container OS  │ │ Container OS  │  │ Container OS  │  │                   
       │  │   (CentOS)    │ │   (CentOS)    │  │   (CentOS)    │  │                   
       │  │ ┌───────────┐ │ │ ┌───────────┐ │  │ ┌───────────┐ │  │                   
       │  │ │ Couchbase │ │ │ │ Couchbase │ │  │ │ Couchbase │ │  │                   
       │  │ │  Server   │ │ │ │  Server   │ │  │ │  Server   │ │  │                   
       │  │ └───────────┘ │ │ └───────────┘ │  │ └───────────┘ │  │                   
       │  │               │ │               │  │               │  │                   
       │  └───────────────┘ └───────────────┘  └───────────────┘  │                   
       │                                                          │                   
       └──────────────────────────────────────────────────────────┘                   


## Multiple hosts, multiple containers per host (hard)

This is very difficult to achieve with Docker, because there is no native way to allow each container unrestricted access to the internal IPs of containers running on other hosts. There are software networking layers such as [Flannel](https://github.com/coreos/flannel "Flannel") and [Weave](https://github.com/weaveworks/weave "Weave"), but it is beyond the scope of this README to cover how those might be configured. This is not a particularly useful deployment scenario for either testing or production use, so we will simply suggest that you not try this.
