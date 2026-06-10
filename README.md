# :lobster: Obstor Quickstart Guide
[![Discord](https://pgg.net/discord?type=svg)](https://pgg.net/discord)

[![Obstor](https://raw.githubusercontent.com/cloudment/obstor/main/.github/logo.svg?sanitize=true)](https://obster.net)

Obstor is a high-performance object storage system supporting popular transfer protocols like S3 and SFTP, making it suitable for building high-performance infrastructure for machine learning, analytics, and application data workloads. Obstor is based on the 2021 Apache-licensed release of MinIO, prior to the project's transition to AGPL and later archival.

This README provides quickstart instructions on running Obstor on baremetal hardware, including Docker-based installations. For Kubernetes environments,
use the [Obstor Kubernetes Operator](https://github.com/minio/operator/blob/master/README.md).

# Docker Installation

Use the following commands to run a standalone Obstor server on a Docker container.

Standalone Obstor servers are best suited for early development and evaluation. Certain features such as versioning, object locking, and bucket replication
require distributed deploying Obstor with Erasure Coding. For extended development and production, deploy Obstor with Erasure Coding enabled - specifically,
with a *minimum* of 4 drives per Obstor server. See [Obstor Erasure Code Quickstart Guide](https://obstor.net/docs/obstor-erasure-code-quickstart-guide)
for more complete documentation.

## Stable

Run the following command to run the latest stable image of Obstor on a Docker container using an ephemeral data volume:

```sh
docker run -p 9000:9000 ghcr.io/cloudment/obstor server /data
```

The Obstor deployment starts using default root credentials `obstoradmin:obstoradmin`. You can test the deployment using the Obstor Browser, an embedded
web-based object browser built into Obstor Server. Point a web browser running on the host machine to http://127.0.0.1:9000 and log in with the
root credentials. You can use the Browser to create buckets, upload objects, and browse the contents of the Obstor server.

You can also connect using any S3-compatible tool, such as the Obstor Client `mc` commandline tool. See
[Test using Obstor Client `mc`](#test-using-obstor-client-mc) for more information on using the `mc` commandline tool. For application developers,
see https://obstor.net/docs/ and click **OBSTOR SDKS** in the navigation to view Obstor SDKs for supported languages.


> NOTE: To deploy Obstor on Docker with persistent storage, you must map local persistent directories from the host OS to the container using the
  `docker -v` option. For example, `-v /mnt/data:/data` maps the host OS drive at `/mnt/data` to `/data` on the Docker container.

# macOS

Use the following commands to run a standalone Obstor server on macOS.

Standalone Obstor servers are best suited for early development and evaluation. Certain features such as versioning, object locking, and bucket replication
require distributed deploying Obstor with Erasure Coding. For extended development and production, deploy Obstor with Erasure Coding enabled - specifically,
with a *minimum* of 4 drives per Obstor server. See [Obstor Erasure Code Quickstart Guide](https://obstor.net/docs/obstor-erasure-code-quickstart-guide)
for more complete documentation.

# GNU/Linux

Use the following command to run a standalone Obstor server on Linux hosts running 64-bit Intel/AMD architectures. Replace ``/data`` with the path to the drive or directory in which you want Obstor to store data.

```sh
wget https://dl.pgg.net/packages/obstor/release/linux-amd64/obstor
chmod +x obstor
./obstor server /data
```

Replace ``/data`` with the path to the drive or directory in which you want Obstor to store data.

The following table lists supported architectures. Replace the `wget` URL with the architecture for your Linux host.

| Architecture                   | URL                                                        |
| --------                       | ------                                                     |
| 64-bit Intel/AMD               | https://dl.pgg.net/packages/obstor/release/linux-amd64/obstor   |
| 64-bit ARM                     | https://dl.pgg.net/packages/obstor/release/linux-arm64/obstor   |
| 64-bit PowerPC LE (ppc64le)    | https://dl.pgg.net/packages/obstor/release/linux-ppc64le/obstor |
| IBM Z-Series (S390X)           | https://dl.pgg.net/packages/obstor/release/linux-s390x/obstor   |

The Obstor deployment starts using default root credentials `obstoradmin:obstoradmin`. You can test the deployment using the Obstor Browser, an embedded
web-based object browser built into Obstor Server. Point a web browser running on the host machine to http://127.0.0.1:9000 and log in with the
root credentials. You can use the Browser to create buckets, upload objects, and browse the contents of the Obstor server.

You can also connect using any S3-compatible tool, such as the Obstor Client `mc` commandline tool. See
[Test using Obstor Client `mc`](#test-using-obstor-client-mc) for more information on using the `mc` commandline tool. For application developers,
see https://obstor.net/docs/ and click **OBSTOR SDKS** in the navigation to view Obstor SDKs for supported languages.


> NOTE: Standalone Obstor servers are best suited for early development and evaluation. Certain features such as versioning, object locking, and bucket replication
require distributed deploying Obstor with Erasure Coding. For extended development and production, deploy Obstor with Erasure Coding enabled - specifically,
with a *minimum* of 4 drives per Obstor server. See [Obstor Erasure Code Quickstart Guide](https://obstor.net/docs/obstor-erasure-code-quickstart-guide)
for more complete documentation.

# Microsoft Windows

To run Obstor on 64-bit Windows hosts, download the Obstor executable from the following URL:

```sh
https://dl.pgg.net/packages/obstor/release/windows-amd64/obstor.exe
```

Use the following command to run a standalone Obstor server on the Windows host. Replace ``D:\`` with the path to the drive or directory in which you want Obstor to store data. You must change the terminal or powershell directory to the location of the ``obstor.exe`` executable, *or* add the path to that directory to the system ``$PATH``:

```sh
obstor.exe server D:\
```

The Obstor deployment starts using default root credentials `obstoradmin:obstoradmin`. You can test the deployment using the Obstor Browser, an embedded
web-based object browser built into Obstor Server. Point a web browser running on the host machine to http://127.0.0.1:9000 and log in with the
root credentials. You can use the Browser to create buckets, upload objects, and browse the contents of the Obstor server.

You can also connect using any S3-compatible tool, such as the Obstor Client `mc` commandline tool. See
[Test using Obstor Client `mc`](#test-using-obstor-client-mc) for more information on using the `mc` commandline tool. For application developers,
see https://obstor.net/docs/ and click **OBSTOR SDKS** in the navigation to view Obstor SDKs for supported languages.

> NOTE: Standalone Obstor servers are best suited for early development and evaluation. Certain features such as versioning, object locking, and bucket replication
require distributed deploying Obstor with Erasure Coding. For extended development and production, deploy Obstor with Erasure Coding enabled - specifically,
with a *minimum* of 4 drives per Obstor server. See [Obstor Erasure Code Quickstart Guide](https://obstor.net/docs/obstor-erasure-code-quickstart-guide)
for more complete documentation.

# Install from Source

Use the following commands to compile and run a standalone Obstor server from source. Source installation is only intended for developers and advanced users. If you do not have a working Golang environment, please follow [How to install Golang](https://golang.org/doc/install). Recommended version is [go1.26](https://golang.org/dl/#stable) or newer.

```sh
GO111MODULE=on go get github.com/obstor/obstor
```

The Obstor deployment starts using default root credentials `obstoradmin:obstoradmin`. You can test the deployment using the Obstor Browser, an embedded
web-based object browser built into Obstor Server. Point a web browser running on the host machine to http://127.0.0.1:9000 and log in with the
root credentials. You can use the Browser to create buckets, upload objects, and browse the contents of the Obstor server.

You can also connect using any S3-compatible tool, such as the Obstor Client `mc` commandline tool. See
[Test using Obstor Client `mc`](#test-using-obstor-client-mc) for more information on using the `mc` commandline tool. For application developers,
see https://obstor.net/docs/ and click **OBSTOR SDKS** in the navigation to view Obstor SDKs for supported languages.


> NOTE: Standalone Obstor servers are best suited for early development and evaluation. Certain features such as versioning, object locking, and bucket replication
require distributed deploying Obstor with Erasure Coding. For extended development and production, deploy Obstor with Erasure Coding enabled - specifically,
with a *minimum* of 4 drives per Obstor server. See [Obstor Erasure Code Quickstart Guide](https://obstor.net/docs/obstor-erasure-code-quickstart-guide)
for more complete documentation.

Obstor strongly recommends *against* using compiled-from-source Obstor servers for production environments.

# Deployment Recommendations

## Allow port access for Firewalls

By default Obstor uses the port 9000 to listen for incoming connections. If your platform blocks the port by default, you may need to enable access to the port.

### ufw

For hosts with ufw enabled (Debian based distros), you can use `ufw` command to allow traffic to specific ports. Use below command to allow access to port 9000

```sh
ufw allow 9000
```

Below command enables all incoming traffic to ports ranging from 9000 to 9010.

```sh
ufw allow 9000:9010/tcp
```

### firewall-cmd

For hosts with firewall-cmd enabled (CentOS), you can use `firewall-cmd` command to allow traffic to specific ports. Use below commands to allow access to port 9000

```sh
firewall-cmd --get-active-zones
```

This command gets the active zone(s). Now, apply port rules to the relevant zones returned above. For example if the zone is `public`, use

```sh
firewall-cmd --zone=public --add-port=9000/tcp --permanent
```

Note that `permanent` makes sure the rules are persistent across firewall start, restart or reload. Finally reload the firewall for changes to take effect.

```sh
firewall-cmd --reload
```

### iptables

For hosts with iptables enabled (RHEL, CentOS, etc), you can use `iptables` command to enable all traffic coming to specific ports. Use below command to allow
access to port 9000

```sh
iptables -A INPUT -p tcp --dport 9000 -j ACCEPT
service iptables restart
```

Below command enables all incoming traffic to ports ranging from 9000 to 9010.

```sh
iptables -A INPUT -p tcp --dport 9000:9010 -j ACCEPT
service iptables restart
```

## Pre-existing data
When deployed on a single drive, Obstor server lets clients access any pre-existing data in the data directory. For example, if Obstor is started with the command  `obstor server /mnt/data`, any pre-existing data in the `/mnt/data` directory would be accessible to the clients.

The above statement is also valid for all gateway backends.

# Test Obstor Connectivity

## Test using Browser Dashboard
Obstor Server comes with an embedded web based object browser. Point your web browser to http://127.0.0.1:9000 to ensure your server has started successfully.

![Dashboard](https://raw.githubusercontent.com/cloudment/obstor/main/docs/screenshots/dashboard.png)

![Object Browser](https://raw.githubusercontent.com/cloudment/obstor/main/docs/screenshots/browser.png)

## Test using Obstor Client `mc`
`mc` provides a modern alternative to UNIX commands like ls, cat, cp, mirror, diff etc. It supports filesystems and Amazon S3-compatible cloud storage services. Follow the Obstor Client [Quickstart Guide](https://obstor.net/docs/obstor-client-quickstart-guide) for further instructions.

# Upgrading Obstor
Obstor server supports rolling upgrades, i.e. you can update one Obstor instance at a time in a distributed cluster. This allows upgrades with no downtime. Upgrades can be done manually by replacing the binary with the latest release and restarting all servers in a rolling fashion. However, we recommend all our users to use [`mc admin update`](https://obstor.net/docs/obstor-admin-complete-guide#update) from the client. This will update all the nodes in the cluster simultaneously and restart them, as shown in the following command from the Obstor client (mc):

```
mc admin update <obstor alias, e.g., myobstor>
```

> NOTE: some releases might not allow rolling upgrades, this is always called out in the release notes and it is generally advised to read release notes before upgrading. In such a situation `mc admin update` is the recommended upgrading mechanism to upgrade all servers at once.

## Important things to remember during Obstor upgrades

- `mc admin update` will only work if the user running Obstor has write access to the parent directory where the binary is located, for example if the current binary is at `/usr/local/bin/obstor`, you would need write access to `/usr/local/bin`.
- `mc admin update` updates and restarts all servers simultaneously, applications would retry and continue their respective operations upon upgrade.
- `mc admin update` is disabled in kubernetes/container environments, container environments provide their own mechanisms to rollout of updates.
- In the case of federated setups `mc admin update` should be run against each cluster individually. Avoid updating `mc` to any new releases until all clusters have been successfully updated.
- If using `kes` as KMS with Obstor, just replace the binary and restart `kes` more information about `kes` can be found [here](https://github.com/minio/kes/wiki)
- If using Vault as KMS with Obstor, ensure you have followed the Vault upgrade procedure outlined here: https://www.vaultproject.io/docs/upgrading/index.html
- If using etcd with Obstor for the federation, ensure you have followed the etcd upgrade procedure outlined here: https://github.com/etcd-io/etcd/blob/master/Documentation/upgrades/upgrading-etcd.md

# Explore Further
- [Obstor Erasure Code QuickStart Guide](https://obstor.net/docs/obstor-erasure-code-quickstart-guide)
- [Use `mc` with Obstor Server](https://obstor.net/docs/obstor-client-quickstart-guide)
- [Use `aws-cli` with Obstor Server](https://obstor.net/docs/aws-cli-with-obstor)
- [Use `s3cmd` with Obstor Server](https://obstor.net/docs/s3cmd-with-obstor)
- [Use `minio-go` SDK with Obstor Server](https://obstor.net/docs/golang-client-quickstart-guide)
- [The Obstor documentation website](https://obstor.net/docs/obstor)

# Contribute to Obstor Project
Please follow Obstor [Contributor's Guide](https://github.com/obstor/obstor/blob/main/CONTRIBUTING.md)

# License
Use of Obstor is governed by the Apache 2.0 License found at [LICENSE](https://github.com/obstor/obstor/blob/main/LICENSE).
