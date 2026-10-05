# Introduction to Operational Insights

Operational Insights is a self-managed, JSON-native NoSQL analytical database. It serves to unify data from diverse sources, allowing for the execution of complex analytical queries and the extraction of timely insights.

Operational Insights harnesses the power of NoSQL for analytics. It integrates seamlessly with both Couchbase Server and the Couchbase Capella cloud platform, enabling the creation of real-time, adaptive applications.

Traditionally, analyzing JSON data in NoSQL databases requires complex transformations, like flattening, to prepare it for analytics, causing delays and hindering real-time insights. Operational Insights eliminates these ETL complexities by using a unifying JSON data model with schema flexibility. This allows data to fluidly evolve at its source without requiring manual schema or transformation management. This facilitates a Zero ETL environment, leading to faster time to insight, reduced costs, and increased agility.

## Features
* A column-oriented, Log-Structured Merge (LSM) tree–based storage engine delivers scalable analytic performance and capacity for customers with self-managed on-premises or cloud deployments. The LSM tree architecture provides high write throughput for fast data ingestion, while columnar storage accelerates analytical queries by accessing only the necessary columns.
* A shared-nothing compute and shared-object storage architecture that allows customers to scale compute resources independently of storage.
* An enhanced MPP-based query engine enables scalable, real-time analytical query computation.
* A cost-based optimizer improves query execution without requiring user intervention. Using a sample-based approach, it quickly estimates data statistics from a small subset of the data, enabling it to identify the lowest-cost query plan without scanning the entire dataset.
* Zero ETL for incoming data, with real-time ingestion capabilities powered by Confluent Kafka, that provide the ability to connect, capture, and extract data from nearly any database or application. One can optionally modify the target JSON structure of the incoming data while in transit, for example, to omit or modify its fields.
* Data Lakehouse capabilities that enable direct querying from Amazon S3 and S3-compatible storage, with support for formats including JSON, Parquet, Avro, CSV, TSV, and Delta tables, providing the ability for queries to combine external data with other data in Operational Insights.
* A SQL++ based path for writing the results of a query back to the Couchbase Operational data service to support adaptive applications.
* A tabular view facility that provides native SQL-based support for Tableau, PowerBI and Apache Superset for building business reports, visualizations, and dashboards.

This Docker image is designed to make it easy to run Operational Insights for development, testing, and proof-of-concept environments.

## Quickstart with Operational Insights and Docker

To quickly get started with Operational Insights, you can run an instance using Docker. This is ideal for development and testing purposes.

### Prerequisites
These instructions assume the following:
1. Docker installed and running
1. No services running on ports `8091` or `8095`
1. No existing containers named `insights` (or `versitygw` if using Versity S3 Gateway)

### 1. Create a Docker network
Create a user-defined network so the container can communicate with other services if needed.

```bash
docker network create insights-net
```

### 2. Start Versity S3 Gateway (optional)
If you don't have an S3-compatible object store to use as the blob storage backend, you can run [Versity S3 Gateway](https://github.com/versity/versitygw) locally. Otherwise, you need to configure Operational Insights to use a different blob storage backend.

Start the gateway, then create a bucket called `insights-storage`:

```bash
docker run -d --name versitygw --network insights-net -v insights-s3-data:/data \
       versity/versitygw --access Administrator --secret password posix /data

docker exec versitygw versitygw admin --access Administrator --secret password \
       --endpoint-url http://localhost:7070 create-bucket --bucket insights-storage --owner Administrator
```

The `insights-s3-data` named volume keeps the stored data across container restarts. Use a named volume rather than a host directory: on macOS, a bind-mounted folder needs extra gateway options and treats object keys as case-insensitive. Replace `Administrator` / `password` with credentials of your own.

### 3. Start the Operational Insights container
Run the Operational Insights container with host and port mappings for the Couchbase Web Console and Operational Insights service, exposed on ports `8091` and `8095` on the host.

```bash
docker run -d --name insights --network insights-net -p 8091:8091 -p 8095:8095 couchbase/operational-insights:3.0.0
```

### 4. Initialize the cluster

Next, visit http://localhost:8091 on the host machine to see the Web Console to start Operational Insights setup.

![Setup splash screen](https://d774lla4im6mk.cloudfront.net/ea/setup-initial.png)

Walk through the Setup wizard

![Setup wizard](https://d774lla4im6mk.cloudfront.net/ea/setup-wizard.png)

If using Versity S3 Gateway, configure the blob storage settings in the wizard as follows:

| Setting                   | Value                   |
|---------------------------|-------------------------|
| Storage Scheme            | S3-Compatible Storage   |
| Storage Endpoint          | `http://versitygw:7070` |
| Bucket Name               | `insights-storage`      |
| Bucket Region             | `us-east-1`             |
| Authentication            | Static Credentials      |
| Access Key ID             | `Administrator`         |
| Secret Access Key         | `password`              |
| Use Path-Style Addressing | Enabled                 |

Otherwise, configure the blob storage settings to point to your chosen backend (e.g. AWS S3 or another S3-compatible service/appliance).

### 5. Install travel-sample dataset

After completing the setup, the console will load.

![Workbench](https://d774lla4im6mk.cloudfront.net/ea/workbench.png)
![Install Samples](https://d774lla4im6mk.cloudfront.net/ea/install-samples.png)

### 6. Execute a sample query
You can now run a sample query to verify that everything is working correctly. For example, you can run the following SQL++ query to get a count of airports in the `travel-sample`.

![Install Samples](https://d774lla4im6mk.cloudfront.net/ea/sample-query.png)

### 7. Next steps

You can now explore the features of Operational Insights, such as creating views, running more complex queries, and integrating with other data sources.

See the [Operational Insights documentation](https://docs.couchbase.com/operational-insights/current/intro/intro.html) for more information.

Alternatively, you can follow the instructions below to set up a multi-node cluster using Versity S3 Gateway as the blob storage backend.

## Running a Two-Node Operational Insights Cluster with Versity S3 Gateway

The following example shows how to start a two-node Operational Insights cluster, using [Versity S3 Gateway](https://github.com/versity/versitygw) as the blob storage backend.

### Prerequisites
These instructions assume the following:

1. Docker installed and running
1. No existing containers named `versitygw`, `insights1`, or `insights2`
1. No services running on ports `8091`, `8095`, `9091`, or `9095`

### 1. Create a Docker network

Create a user-defined network so the containers can talk to each other by name.

```bash
docker network create insights-net
```

### 2. Start Versity S3 Gateway

Start the gateway with its data on the `insights-s3-data` named volume, then create a bucket called `insights-storage`.

```bash
docker run -d --name versitygw --network insights-net -v insights-s3-data:/data \
       versity/versitygw --access Administrator --secret password posix /data

docker exec versitygw versitygw admin --access Administrator --secret password \
       --endpoint-url http://localhost:7070 create-bucket --bucket insights-storage --owner Administrator
```

### 3. Start the first Operational Insights node

Run the first node (`insights1`) with host and port mappings for the Couchbase Web Console and Analytics service.

```bash
docker run -d --name insights1 --network insights-net --hostname insights1.example.com --network-alias insights1.example.com \
       -p 8091:8091 -p 8095:8095 couchbase/operational-insights:3.0.0
```

### 4. Start the second Operational Insights node

Run the second node (`insights2`) with its own mapped ports so you can access it separately from `insights1`.

```bash
docker run -d --name insights2 --network insights-net --hostname insights2.example.com --network-alias insights2.example.com \
       -p 9091:8091 -p 9095:8095 couchbase/operational-insights:3.0.0
```

### 5. Wait for the nodes to be ready

Before proceeding, ensure the nodes are fully booted and ready for configuration. You can check the status of each node by querying the `/pools/default` API. The API should return a `404` status with the text 'unknown pool' when the server is ready to accept configuration.

e.g.
```
$ curl http://localhost:8091/pools/default
"unknown pool"
```

### 6. Initialize the nodes

Initialize `insights1` and `insights2` nodes with hostnames and admin credentials.

```bash
docker exec insights1 couchbase-cli node-init \
  --cluster http://localhost:8091 \
  --username Administrator \
  --password password \
  --node-init-hostname insights1.example.com

docker exec insights2 couchbase-cli node-init \
  --cluster http://localhost:8091 \
  --username Administrator \
  --password password \
  --node-init-hostname insights2.example.com
```

### 7. Configure blob storage to use Versity S3 Gateway

* Configure Operational Insights to use the gateway endpoint and its credentials

```bash
docker exec insights1 curl -s -X POST http://localhost:8091/settings/analytics \
  -d blobStorageScheme=s3 \
  -d blobStorageBucket=insights-storage \
  -d blobStorageRegion=us-east-1 \
  -d blobStorageEndpoint=http://versitygw:7070 \
  -d blobStoragePathStyleAddressing=true \
  -d blobStorageAccessKeyId=Administrator \
  -d blobStorageSecretAccessKey=password
```

The response echoes the settings back, with the secret masked, together with a warning that the HTTP endpoint is insecure. That is expected for a local gateway; use an HTTPS endpoint for anything beyond development.

### 8. Initialize the cluster

Initialize the Operational Insights cluster.

```bash
docker exec insights1 couchbase-cli cluster-init \
  --cluster http://localhost:8091 \
  --cluster-username Administrator \
  --cluster-password password
```

### 9. Add the second node (insights2) to the cluster (insights1)

Add `insights2` to the cluster, and perform a rebalance.

```bash
docker exec insights1 couchbase-cli server-add \
  --cluster http://localhost:8091 \
  --username Administrator \
  --password password \
  --server-add insights2.example.com \
  --server-add-username Administrator \
  --server-add-password password
  
docker exec insights1 couchbase-cli rebalance \
  --cluster http://localhost:8091 \
  --username Administrator \
  --password password
```

### 10. Access the Web Console

Once rebalanced, the cluster is ready to be used. Access the UI at:

- **insights1:** [http://localhost:8091](http://localhost:8091)
- **insights2:** [http://localhost:9091](http://localhost:9091)

## Ports

| Port  | Description                   |
|-------|-------------------------------|
| 8091  | Web console / REST API (HTTP) |
| 8095  | Analytics HTTP API            |
| 18091 | Web console / REST API (HTTPS)|
| 18095 | Analytics HTTPS API           |

## Volumes

Data in Operational Insights is stored under `/opt/couchbase/var/lib/couchbase/data`. For persistent deployments, mount a Docker volume or host directory to this path.

Example:

```bash
docker run -d --name insights1 -v insights1-data:/opt/couchbase/var/lib/couchbase/data   couchbase/operational-insights:3.0.0
```

## License

Operational Insights is licensed under the Couchbase Enterprise License Agreement.
