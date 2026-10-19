# Operational Insights UDF Executor

This image is the sandbox that runs Python user-defined functions (UDFs) for [Couchbase Operational Insights](https://hub.docker.com/r/couchbase/operational-insights). Python UDFs let you extend SQL++ with code written in Python, for example to bring a trained ML model into a query, or to implement logic that is difficult to express in SQL++.

**You do not need to run this image yourself.** It is pulled and started on each node by the `operational-insights-udf` add-on package, which is how Python UDFs are enabled. See [Python User-Defined Functions](https://docs.couchbase.com/operational-insights/current/sqlpp/9b_python_udf.html).

## How it is used

Python UDFs require the `operational-insights-udf` add-on package, installed on every Operational Insights node. Without it, no node can evaluate a Python UDF.

When the package is installed, it:

1. Installs the gVisor `runsc` runtime and registers it with Docker.
1. Pulls this image.
1. Runs one executor container per node, shared by every UDF library on that node, under the gVisor runtime and as the node's `couchbase` user.

The node and the executor communicate over a Unix domain socket. The executor mounts the node's applications directory read-only, so that it can import the libraries you upload, and it has no network access.

## Requirements

* An Operational Insights node installed from the `.deb` or `.rpm` package, with the matching version of the `operational-insights-udf` add-on package.
* Docker installed and running on the host (`docker.service`), with access to this image. Docker is not a declared dependency of the add-on package, so the install fails if Docker is missing or not running.

## Using a mirrored or private registry

By default the add-on package pulls this image from Docker Hub. On hosts that cannot reach Docker Hub, push the image to a registry the host can reach, and set `UDF_EXECUTOR_IMAGE` to its full name and tag when installing the add-on package, for example:

```bash
sudo UDF_EXECUTOR_IMAGE=registry.example.com/couchbase/operational-insights-udf:3.0.0 \
  apt install ./operational-insights-udf_<version>-linux_amd64.deb
```

Set `UDF_EXECUTOR_IMAGE` again whenever you upgrade the add-on package; otherwise the upgrade pulls from Docker Hub.

## Security

Python UDFs run in an isolated sandbox, but this is not a security boundary. Vet any code that you deploy as a UDF library carefully, and assume that anyone who can create a function can run arbitrary code in the sandbox as part of a query. See [Security Considerations](https://docs.couchbase.com/operational-insights/current/install/security-considerations.html).

## Tags

Each tag matches an Operational Insights release, for example `3.0.0`. Use the tag that matches the Operational Insights version of your nodes.

The image is published for both `linux/amd64` and `linux/arm64`.

## Additional References

- [Python User-Defined Functions](https://docs.couchbase.com/operational-insights/current/sqlpp/9b_python_udf.html)
- [Creating an External Function](https://docs.couchbase.com/operational-insights/current/sqlpp/9_udf.html)
- [Operational Insights documentation](https://docs.couchbase.com/operational-insights/current/intro/intro.html)

## License

Operational Insights is licensed under the Couchbase Enterprise License Agreement, available on the [Legal Agreements](https://www.couchbase.com/legal/agreements) page.
