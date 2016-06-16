
This README will guide you through running Couchbase Server with Docker Containers.

[Couchbase Server](http://www.couchbase.com/nosql-databases/couchbase-server) is a NoSQL document database with a distributed architecture for performance, scalability, and availability. It enables developers to build applications easier and faster by leveraging the power of SQL with the flexibility of JSON.

For additional questions and feedback, please visit the [Couchbase Forums](https://forums.couchbase.com/) or [Stack Overflow] (http://stackoverflow.com/questions/tagged/couchbase).

# QuickStart with Couchbase Server and Docker #

**Step - 1 :** Run Couchbase Server docker container

```docker run -d --name db -p 8091-8094:8091-8094 -p 11210:11210 couchbase```

**Step - 2 :** Next, visit `http://localhost:8091` on the host machine to see the Web Console to start Couchbase Server setup.

TODO: Image : link to the setup wizard fist page.  

Walk through the Setup wizard and accept the default values. 
- Note: You may need to lower the RAM allocated to various services to fit within the bounds of the resource of the containers.
- Enable the beer-sample bucket to load some sample data.

**Note :** For detailed information on configuring the Server, see [Initial Couchbase Server Setup](http://developer.couchbase.com/documentation/server/4.5/install/init-setup.html).

## Running A N1QL Query on the Couchbase Server Cluster ##
N1QL is the SQL based query language for Couchbase Server. Simply switch to the Query tab on the Web Console at `http://localhost:8091` and run the following N1QL Query in the query window:

```SELECT name FROM `beer-sample` WHERE  brewery_id ="mishawaka_brewing";```

You can also execute N1QL queries from the commandline. To run a query from command line query tool, run the interactive shell on the container:

`bash -c "clear && docker exec -it db sh"`

Then, navigate to the `bin` directory under Couchbase Server installation and run cbq command line tool and execute the N1QL Query on `beer-sample` bucket

`/opt/couchbase/bin/cbq`

```cbq> SELECT name FROM `beer-sample` WHERE  brewery_id ="mishawaka_brewing";```

For more query samples, refer to the [Running your first N1QL query](http://developer.couchbase.com/documentation/server/4.5/getting-started/first-n1ql-query.html) guide.   

## Connect to the Couchbase Server Cluster via Applications and SDKs ##
Couchbase Server SDKs comes in many languages:  C SDK 2.4/2.5 Go, Java, .NET, Node.js, PHP, Python. Simply run your application through the Couchbase Server SDK of your choice on the host, and point it to http://localhost:8091/pools to connect to the container.

For running a sample application, refer to the [Running a sample Web app](http://developer.couchbase.com/documentation/server/4.5/travel-app/index.html) guide.

# Requirements and Best Practices #
## Container Requirements ##
Official Couchbase Server containers on Docker Hub are based on Ubuntu 14.04. 

**Docker Container Resource Requirements :** For minimum container requirements, you can follow [Couchbase Server minimum HW recommendations](http://developer.couchbase.com/documentation/server/current/install/pre-install.html) for development, test and production environments. 

## Best Practices ##
** Avoid a Single Point of Failure : ** Couchbase Server's resilience and high-availability are achieved through creating a cluster of independent nodes and replicating data between them so that any individual node failure doesn't lead to loss of access to your data. In a containerized environment, if you were to run multiple nodes on the same piece of physical hardware, you can inadvertently re-introduce a single point of failure. In environments where you control VM placement, we advise ensuring each Couchbase Server node runs on a different piece of physical hardware.

** Sizing your containers : ** Physical hardware performance characteristics are well understood. Even though containers insert a lightweight layer between Couchbase Server and the underlying OS, there is still a small overhead in running Couchbase Server in containers. For stability and better performance predictability, It is recommended to have at least 2 cores dedicated to the container in development environments and 4 cores dedicated to the container rather than shared across multiple containers for Couchbase Server instances running in production. 
With an over-committed environment you can end up with containers competing with each other causing unpredictable performance and sometimes stability issues. 

** Map Couchbase Node Specific Data to a Local Folder : ** A Couchbase Server Docker container will write all persistent and node-specific data under the directory /opt/couchbase/var by default. It is recommended to map this directory to a directory on the host file system using the `-v` option to `docker run` to get persistence and performance.
* Persistence: Storing `/opt/couchbase/var` outside the container with the `-v` option, allows you to delete the container and recreate it later without loosing the data in Couchbase Server. You can even update to a container running a later release/version of Couchbase Server without losing your data.
* Performance: In a standard Docker environment using a union file system, leaving /opt/couchbase/var inside the container results in some amount of performance degradation.

> NOTE for SELinux : If you have SELinux enabled, mounting the host volumes in a container requires an extra step. Assuming you are mounting the `~/couchbase` directory on the host file system, you need to run the following command once before running your first container on that host:

`mkdir ~/couchbase && chcon -Rt svirt_sandbox_file_t ~/couchbase`

** Increase ULIMIT in Production Deployments :** Couchbase Server normally expects the following changes to ulimits: 
`ulimit -n 40960        # nofile: max number of open files`

`ulimit -c unlimited    # core: max core file size`

`ulimit -l unlimited    # memlock: maximum locked-in-memory address space`

These ulimit settings are necessary when running under heavy load. If you are just doing light testing and development, you can omit these settings, and everything will still work. 

To set the ulimits in your container, you will need to run Couchbase Docker containers with the following additional --ulimit flags: 

`docker run -d --ulimit nofile=40960:40960 --ulimit core=100000000:100000000 --ulimit memlock=100000000:100000000 --name db -p 8091-8094:8091-8094 -p 11210:11210 couchbase``

Since "unlimited" is not supported as a value, it sets the core and memlock values to 100 GB. If your system has more than 100 GB RAM, you will want to increase this value to match the available RAM on the system. 

Note:The --ulimit flags only work on Docker 1.6 or later.

** Networking :** Couchbase Server communicates on many different ports (see the [Couchbase Server documentation](http://docs.couchbase.com/admin/admin/Install/install-networkPorts.html "Network ports page on Couchbase Server documentation")). Also, it is generally not supported that the cluster nodes be placed behind any NAT. For these reasons, Docker's default networking configuration is not ideally suited to Couchbase Server deployments. For production deployments it is recomended to use --net=host setting to avoid performance and reliability issues. 


# Multi Node Couchbase Server Cluster Deployment Topologies #
With multi node Couchbase Server clusters, there are 2 popular topologies. 
* All Couchbase Server containers on one physical machine: This model is commonly used for scale-minimized deployments simulating production deployments for development and test purposes. 
* Each Couchbase Server container on its own machine: This model is commonly used for production deployments. It prevents Couchbase Server nodes from stepping over each other and gives you better performance predictability.  This is the supported topology in production with Couchbase Server 4.5 and higher. 

## Development and Test Cluster Deployment with Multiple Couchbase Server Containers on One Physical Machine
In this deployment model all containers are placed on the same physical node. Placing all containers on a single physical machine means all containers will compete for the same resources. That is fine for test systems, however it isn’t recommended for applications sensitive to performance in production.

```
┌──────────────────────────────────────────────────────────┐
│                     Host OS (Linux)                      │
│                                                          │
│  ┌───────────────┐ ┌───────────────┐  ┌───────────────┐  │
│  │ Container OS  │ │ Container OS  │  │ Container OS  │  │
│  │   (Ubuntu)    │ │   (Ubuntu)    │  │   (Ubuntu)    │  │
│  │ ┌───────────┐ │ │ ┌───────────┐ │  │ ┌───────────┐ │  │
│  │ │ Couchbase │ │ │ │ Couchbase │ │  │ │ Couchbase │ │  │
│  │ │  Server   │ │ │ │  Server   │ │  │ │  Server   │ │  │
│  │ └───────────┘ │ │ └───────────┘ │  │ └───────────┘ │  │
│  └───────────────┘ └───────────────┘  └───────────────┘  │
└──────────────────────────────────────────────────────────┘
```




## Multiple hosts, single container on each host

```
┌───────────────────────┐  ┌───────────────────────┐  ┌───────────────────────┐
│   Host OS (Linux)     │  │   Host OS (Linux)     │  │   Host OS (Linux)     │
│  ┌─────────────────┐  │  │  ┌─────────────────┐  │  │  ┌─────────────────┐  │
│  │  Container OS   │  │  │  │  Container OS   │  │  │  │  Container OS   │  │
│  │    (Ubuntu)     │  │  │  │    (Ubuntu)     │  │  │  │    (Ubuntu)     │  │
│  │  ┌───────────┐  │  │  │  │  ┌───────────┐  │  │  │  │  ┌───────────┐  │  │
│  │  │ Couchbase │  │  │  │  │  │ Couchbase │  │  │  │  │  │ Couchbase │  │  │
│  │  │  Server   │  │  │  │  │  │  Server   │  │  │  │  │  │  Server   │  │  │
│  │  └───────────┘  │  │  │  │  └───────────┘  │  │  │  │  └───────────┘  │  │
│  └─────────────────┘  │  │  └─────────────────┘  │  │  └─────────────────┘  │
└───────────────────────┘  └───────────────────────┘  └───────────────────────┘
```

This deployment scenario represents a typical Couchbase Server cluster, where each node runs on a dedicated host. We assume that the nodes are located in the same datacenter with high-speed network links between them. We also assume that the datacenter LAN configuration allows each host in the cluster to see other hosts via the known IPs.

Currently, the only supported approach for Couchbase Server on this deployment architecture is to use the `--net=host` flag.

Using the `--net=host` flag will have the following effects:

* The container will use the host's own networking stack, and bind directly to ports on the host.
* Removes networking complications with Couchbase Server being behind a NAT.
* From a networking perspective, it is effectively the same as running Couchbase Server directly on the host.
* There is no need to use `-p` to "expose" any ports. Each container will use the IP address(es) of its host.
* Increased efficiency, as there will be no Docker-imposed networking overhead.

Start a container on *each host* via:

```
docker run -d -v ~/couchbase:/opt/couchbase/var --net=host couchbase/server
```

To configure Couchbase Server:

* Access the Couchbase Server Web Console via port 8091 on any of the hosts.
* Follow the same steps from the *Multiple containers on single host* section. However, use the use the host IP address itself rather than using `docker inspect` to discover the IP address.


## Multiple hosts, multiple containers per host

```
┌─────────────────────────────────────────┐  ┌─────────────────────────────────────────┐
│            Host OS (Linux)              │  │            Host OS (Linux)              │
│ ┌─────────────────┐ ┌─────────────────┐ │  │ ┌─────────────────┐ ┌─────────────────┐ │
│ │  Container OS   │ │  Container OS   │ │  │ │  Container OS   │ │  Container OS   │ │
│ │    (Ubuntu)     │ │    (Ubuntu)     │ │  │ │    (Ubuntu)     │ │    (Ubuntu)     │ │
│ │  ┌───────────┐  │ │  ┌───────────┐  │ │  │ │  ┌───────────┐  │ │  ┌───────────┐  │ │
│ │  │ Couchbase │  │ │  │ Couchbase │  │ │  │ │  │ Couchbase │  │ │  │ Couchbase │  │ │
│ │  │  Server   │  │ │  │  Server   │  │ │  │ │  │  Server   │  │ │  │  Server   │  │ │
│ │  └───────────┘  │ │  └───────────┘  │ │  │ │  └───────────┘  │ │  └───────────┘  │ │
│ └─────────────────┘ └─────────────────┘ │  │ └─────────────────┘ └─────────────────┘ │
└─────────────────────────────────────────┘  └─────────────────────────────────────────┘
```

* This deployment scenario is difficult to achieve with plain vanilla Docker, as there is no native way to allow to each container unrestricted access to the internal IPs of containers running on other hosts.
* There are software networking layers such as [Flannel](https://github.com/coreos/flannel "Flannel") and [Weave](https://github.com/weaveworks/weave "Weave"), but it is beyond the scope of this README to explain how those might be configured.
* This deployment scenario is not particularly useful either for testing or production. You will be better off checking out the various available [cloud hosting scenarios](https://github.com/couchbase/docker/wiki#container-specific-cloud-hosting-platforms).

## Cloud environments

Although it is beyond the scope of this README, there is a [github wiki](https://github.com/couchbase/docker/wiki#container-specific-cloud-hosting-platforms) that contains guidance and instructions on how to run Couchbase Server Docker containers in various cloud environments.


# Licensing

Couchbase Server comes in two editions:

* [Community Edition](http://www.couchbase.com/community) -- free for unrestricted use.

* [Enterprise Edition](http://www.couchbase.com/agreement/subscription) -- free for development, paid subscription required for production deployment.

By default, the `latest` Docker tag points to the latest Enterprise Edition.  If you want the Community Edition instead, you should add the appropriate tag:

```
docker run couchbase/server:community-3.0.1
```
