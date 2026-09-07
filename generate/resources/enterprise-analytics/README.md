# Introduction to Enterprise Analytics

Enterprise Analytics is a self-managed, JSON-native NoSQL analytical database. It serves to unify data from diverse sources, allowing for the execution of complex analytical queries and the extraction of timely insights.

Enterprise Analytics harnesses the power of NoSQL for analytics. It integrates seamlessly with both Couchbase Server and the Couchbase Capella cloud platform, enabling the creation of real-time, adaptive applications.

Traditionally, analyzing JSON data in NoSQL databases requires complex transformations, like flattening, to prepare it for analytics, causing delays and hindering real-time insights. Enterprise Analytics eliminates these ETL complexities by using a unifying JSON data model with schema flexibility. This allows data to fluidly evolve at its source without requiring manual schema or transformation management. This facilitates a Zero ETL environment, leading to faster time to insight, reduced costs, and increased agility.

## Features
* A column-oriented, Log-Structured Merge (LSM) tree–based storage engine delivers scalable analytic performance and capacity for customers with self-managed on-premises or cloud deployments. The LSM tree architecture provides high write throughput for fast data ingestion, while columnar storage accelerates analytical queries by accessing only the necessary columns.
* A shared-nothing compute and shared-object storage architecture that allows customers to scale compute resources independently of storage.
* An enhanced MPP-based query engine enables scalable, real-time analytical query computation.
* A cost-based optimizer improves query execution without requiring user intervention. Using a sample-based approach, it quickly estimates data statistics from a small subset of the data, enabling it to identify the lowest-cost query plan without scanning the entire dataset.
* Zero ETL for incoming data, with real-time ingestion capabilities powered by Confluent Kafka, that provide the ability to connect, capture, and extract data from nearly any database or application. One can optionally modify the target JSON structure of the incoming data while in transit, for example, to omit or modify its fields.
* Data Lakehouse capabilities that enable direct querying from supported object storage, with support for formats including JSON, Parquet, Avro, CSV, TSV, Delta tables, and Apache Iceberg tables via Iceberg catalogs, providing the ability for queries to combine external data with other data in Enterprise Analytics.
* A choice of certified object stores for the underlying storage layer: Amazon S3, S3-compatible storage, [Azure Blob Storage](https://docs.couchbase.com/enterprise-analytics/current/manage/manage-nodes/azure-blob-storage.html) (2.1 and later) and [Google Cloud Storage](https://docs.couchbase.com/enterprise-analytics/current/manage/manage-nodes/google-cloud-storage.html) (2.2 and later).
* Continuous, row-level change data capture from Oracle and SQL Server, powered by Kafka links and Debezium connectors (2.2 and later).
* A SQL++ based path for writing the results of a query back to the Couchbase Operational data service to support adaptive applications.
* A tabular view facility that provides native SQL-based support for Tableau, PowerBI and Apache Superset for building business reports, visualizations, and dashboards.
* An Index Advisor that recommends secondary indexes for your queries, plus index-only query plans that serve requests entirely from covered indexes (2.2 and later).
* JWT authentication for REST API and query requests, including tokens issued by external identity providers (2.2 and later).
* An asynchronous REST API for submitting long-running queries without blocking application workflows (2.2 and later).

This Docker image is designed to make it easy to run Enterprise Analytics for development, testing, and proof-of-concept environments.

## Quickstart with Enterprise Analytics and Docker

To quickly get started with Enterprise Analytics, you can run an instance using Docker. This is ideal for development and testing purposes.

### Prerequisites
These instructions assume the following:
1. Docker installed and running
1. A host that meets the container requirements described under "Container Requirements" below
1. No services running on ports `8091` or `8095`
1. No existing containers named `ea` (or `s3mock` if using S3Mock)
1. No existing Docker network named `ea-net`

### 1. Create a Docker network
Create a user-defined network so the container can communicate with other services if needed.

```bash
docker network create ea-net
```

### 2. Configure S3Mock (optional)
If you want to use S3Mock as the blob storage backend, you can start the S3Mock container first. Otherwise, you need to configure Enterprise Analytics to use one of the other supported blob storage backends — Amazon S3, S3-compatible storage, Azure Blob Storage, or Google Cloud Storage. See [Supported Object Storage Solutions](https://docs.couchbase.com/enterprise-analytics/current/install/supported-platform.html#supported-object-storage-solutions).

```bash
docker run -d --name s3mock --network ea-net \
  -e COM_ADOBE_TESTING_S3MOCK_STORE_INITIAL_BUCKETS=ea-storage \
  adobe/s3mock
```

### 3. Start the Enterprise Analytics container
Run the Enterprise Analytics container with host and port mappings for the Couchbase Web Console and Enterprise Analytics service, exposed on ports `8091` and `8095` on the host.

```bash
docker run -d --name ea --network ea-net -p 8091:8091 -p 8095:8095 couchbase/enterprise-analytics
```

To connect to this container from an SDK, also publish the Data Service ports, e.g. `-p 18091:18091 -p 18095:18095 -p 11207:11207 -p 11210:11210`. See "Ports" below.

### 4. Initialize the cluster

Next, visit http://localhost:8091 on the host machine to see the Web Console to start Enterprise Analytics setup.

> **Note:** The storage scheme and the number of storage partitions are fixed at cluster setup and cannot be changed afterwards. The remaining blob storage settings — bucket, region, path prefix, endpoint, authentication and TLS options — can be changed later under *Settings > Cluster*.

![Setup splash screen](https://d774lla4im6mk.cloudfront.net/ea/setup-initial.png)

Walk through the Setup wizard

![Setup wizard](https://d774lla4im6mk.cloudfront.net/ea/setup-wizard.png)

If using S3Mock, you can configure the blob storage settings to point to the S3Mock endpoint:

![Memory & Blob Storage Configuration-1](https://d774lla4im6mk.cloudfront.net/ea/blob-storage-config-1.png)
![Memory & Blob Storage Configuration-2](https://d774lla4im6mk.cloudfront.net/ea/blob-storage-config-2.png)

Otherwise, configure the blob storage settings to point to your chosen backend (e.g. AWS S3 or another S3-compatible service/appliance).

### 5. Install travel-sample dataset

After completing the setup, the console will load.

![Workbench](https://d774lla4im6mk.cloudfront.net/ea/workbench.png)
![Install Samples](https://d774lla4im6mk.cloudfront.net/ea/install-samples.png)

### 6. Execute a sample query
You can now run a sample query to verify that everything is working correctly. For example, you can run the following SQL++ query to get a count of airlines in the `travel-sample`.

![Sample query](https://d774lla4im6mk.cloudfront.net/ea/sample-query.png)

### 7. Next steps

You can now explore the features of Enterprise Analytics, such as creating views, running more complex queries, and integrating with other data sources.

See the [Enterprise Analytics documentation](https://docs.couchbase.com/enterprise-analytics/current/index.html) for more information.

Alternatively, you can follow the instructions below to set up a multi-node cluster using S3Mock as the blob storage backend.

## Running a Two-Node Enterprise Analytics Cluster with S3Mock

The following example shows how to start a two-node Enterprise Analytics cluster, using [Adobe S3Mock](https://github.com/adobe/S3Mock) as the blob storage backend.

### Prerequisites
These instructions assume the following:

1. Docker installed and running
1. A host that meets the container requirements described under "Container Requirements" below
1. No existing containers named `s3mock`, `ea1`, or `ea2`
1. No existing Docker network named `ea-net`
1. No services running on ports `8091`, `8095`, `9091`, or `9095`

### 1. Create a Docker network

Create a user-defined network so the containers can talk to each other by name.

```bash
docker network create ea-net
```

### 2. Start the Adobe S3Mock service

Start the S3Mock container with an initial bucket called `ea-storage`.

```bash
docker run -d --name s3mock --network ea-net \
  -e COM_ADOBE_TESTING_S3MOCK_STORE_INITIAL_BUCKETS=ea-storage \
  adobe/s3mock
```

### 3. Start the first Enterprise Analytics node

Run the first node (`ea1`) with host and port mappings for the Couchbase Web Console and Analytics service.

```bash
docker run -d --name ea1 --network ea-net --hostname ea1.example.com --network-alias ea1.example.com \
       -p 8091:8091 -p 8095:8095 couchbase/enterprise-analytics
```

### 4. Start the second Enterprise Analytics node

Run the second node (`ea2`) with its own mapped ports so you can access it separately from `ea1`.

```bash
docker run -d --name ea2 --network ea-net --hostname ea2.example.com --network-alias ea2.example.com \
       -p 9091:8091 -p 9095:8095 couchbase/enterprise-analytics
```

### 5. Wait for the nodes to be ready

Before proceeding, ensure the nodes are fully booted and ready for configuration. You can check the status of each node by querying the `/pools/default` API. The API should return a `404` status with the text 'unknown pool' when the server is ready to accept configuration.

e.g.
```
$ curl http://localhost:8091/pools/default
"unknown pool"
```

### 6. Initialize the nodes

Initialize `ea1` and `ea2` nodes with hostnames and admin credentials.

```bash
docker exec ea1 couchbase-cli node-init \
  --cluster http://localhost:8091 \
  --username Administrator \
  --password password \
  --node-init-hostname ea1.example.com

docker exec ea2 couchbase-cli node-init \
  --cluster http://localhost:8091 \
  --username Administrator \
  --password password \
  --node-init-hostname ea2.example.com
```

### 7. Configure blob storage to use S3Mock

* Configure Enterprise Analytics to use the S3Mock endpoint

```bash
docker exec ea1 couchbase-cli setting-enterprise-analytics --cluster http://localhost:8091 \
  --username Administrator --password password \
  --set \
  --scheme s3 \
  --bucket ea-storage \
  --region us-east-1 \
  --endpoint http://s3mock:9090 \
  --anonymous-auth 1 \
  --path-style-addressing 1 \
  --partitions 16
```

`--partitions 16` keeps the number of storage partitions low; the default of 128 adds unnecessary overhead for a small development cluster.

> **Note:** The storage scheme (`--scheme`) and the number of storage partitions (`--partitions`) are fixed once the cluster is initialized, so this step must come before `cluster-init` below. The remaining settings can be changed later with `--set`, or under *Settings > Cluster* in the Web Console.

To use a backend other than S3 or S3-compatible storage, set `--scheme azblob` for [Azure Blob Storage](https://docs.couchbase.com/enterprise-analytics/current/manage/manage-nodes/azure-blob-storage.html) or `--scheme gs` for [Google Cloud Storage](https://docs.couchbase.com/enterprise-analytics/current/manage/manage-nodes/google-cloud-storage.html).

### 8. Initialize the cluster

Initialize the Enterprise Analytics cluster.

```bash
docker exec ea1 couchbase-cli cluster-init \
  --cluster http://localhost:8091 \
  --cluster-username Administrator \
  --cluster-password password
```

### 9. Add the second node (ea2) to the cluster (ea1)

Add `ea2` to the cluster, and perform a rebalance.

```bash
docker exec ea1 couchbase-cli server-add \
  --cluster http://localhost:8091 \
  --username Administrator \
  --password password \
  --server-add ea2.example.com \
  --server-add-username Administrator \
  --server-add-password password
  
docker exec ea1 couchbase-cli rebalance \
  --cluster http://localhost:8091 \
  --username Administrator \
  --password password
```

### 10. Access the Web Console

Once rebalanced, the cluster is ready to be used. Access the UI at:

- **ea1:** [http://localhost:8091](http://localhost:8091)
- **ea2:** [http://localhost:9091](http://localhost:9091)

## Container Requirements

| Requirement | Details |
|-------------|---------|
| CPU         | 2 GHz dual core x86_64 **supporting AVX2 or later**, or 2 GHz dual core 64-bit ARM v8. The container aborts at startup if the host CPU does not meet the required microarchitecture. |
| RAM         | 4 GB minimum; 32 GB or more recommended |
| Disk        | 8 GB minimum, block-based (HDD, SSD, EBS, iSCSI). Network file systems such as CIFS and NFS are not supported. |

For development and testing you can go below these minimums — as little as 1 GB of free RAM and a single core — but production deployments must meet them. See [System Resource Requirements](https://docs.couchbase.com/enterprise-analytics/current/install/sys-resource-req.html).

The image is published for both `linux/amd64` and `linux/arm64`.

For production deployments, raise the container's ulimits:

```bash
docker run -d --name ea \
  --ulimit nofile=40960:40960 \
  --ulimit core=100000000:100000000 \
  --ulimit memlock=100000000:100000000 \
  couchbase/enterprise-analytics
```

`unlimited` is not a supported value, so `core` and `memlock` are set to 100 GB. If your host has more than 100 GB of RAM, raise these to match. See [Deployment Considerations for Virtual Machines and Containers](https://docs.couchbase.com/enterprise-analytics/current/install/vm-container-guidelines.html).

## Ports

The image exposes the following ports:

| Port  | Description                                                                 |
|-------|-----------------------------------------------------------------------------|
| 8091  | Cluster administration REST/HTTP traffic, including the Web Console         |
| 8095  | Enterprise Analytics Service REST/HTTP traffic                              |
| 9123  | Prometheus metrics                                                          |
| 11207 | Data Service (SDK cluster topology and auditing, TLS)                       |
| 11210 | Data Service (SDK cluster topology and auditing)                            |
| 11280 | Data Service Prometheus metrics                                             |
| 18091 | Cluster administration REST/HTTP traffic, including the Web Console (TLS)   |
| 18095 | Enterprise Analytics Service REST/HTTP traffic (TLS)                        |

A multi-node cluster also requires node-to-node connectivity on a number of additional ports. When each node runs in its own container on a shared Docker network, this works without further configuration. For the full list, see [Enterprise Analytics Ports](https://docs.couchbase.com/enterprise-analytics/current/install/cb-enterprise-analytics-ports.html).

## Volumes

Enterprise Analytics writes all persistent, node-specific data — including cluster configuration, local storage, and logs — under `/opt/enterprise-analytics/var`, which the image declares as a volume. For persistent deployments, mount a Docker volume or host directory to this path so you can delete and recreate the container, or upgrade to a later version, without losing data.

Example:

```bash
docker run -d --name ea1 -v ea1-data:/opt/enterprise-analytics/var couchbase/enterprise-analytics
```

Leaving `/opt/enterprise-analytics/var` inside the container's union filesystem also results in some performance degradation.

> **Note:** If SELinux is enabled and you are bind-mounting a host directory, run `chcon -Rt svirt_sandbox_file_t <host-dir>` once before starting the first container on that host.

## Additional References

- [Enterprise Analytics documentation](https://docs.couchbase.com/enterprise-analytics/current/index.html)
- [Install Enterprise Analytics Using Docker](https://docs.couchbase.com/enterprise-analytics/current/install/getting-started-docker.html)
- [Do a Quick Install](https://docs.couchbase.com/enterprise-analytics/current/intro/do-a-quick-install.html)
- [Release Notes](https://docs.couchbase.com/enterprise-analytics/current/release-notes/release-notes.html)

## License

Enterprise Analytics is licensed under the Couchbase Enterprise License Agreement, available on the [Legal Agreements](https://www.couchbase.com/legal/agreements) page. The license is free for development and testing; a paid subscription is required for production deployment. See the [pricing](https://www.couchbase.com/pricing) page for details.
