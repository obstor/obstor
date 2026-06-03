*Federation feature is deprecated and should be avoided for future deployments*

# Federation Quickstart Guide
This document explains how to configure Obstor with `Bucket lookup from DNS` style federation.

## Get started

### 1. Prerequisites
Install Obstor - Obstor Quickstart Guide.

### 2. Run Obstor in federated mode
Bucket lookup from DNS federation requires two dependencies

- etcd (for bucket DNS service records)
- CoreDNS (for DNS management based on populated bucket DNS service records, optional)

## Architecture

![bucket-lookup](https://raw.githubusercontent.com/cloudment/obstor/main/docs/federation/lookup/bucket-lookup.png)

### Environment variables

#### OBSTOR_ETCD_ENDPOINTS

This is comma separated list of etcd servers that you want to use as the Obstor federation back-end. This should
be same across the federated deployment, i.e. all the Obstor instances within a federated deployment should use same
etcd back-end.

#### OBSTOR_DOMAIN

This is the top level domain name used for the federated setup. This domain name should ideally resolve to a load-balancer
running in front of all the federated Obstor instances. The domain name is used to create sub domain entries to etcd. For
example, if the domain is set to `domain.com`, the buckets `bucket1`, `bucket2` will be accessible as `bucket1.domain.com`
and `bucket2.domain.com`.

#### OBSTOR_PUBLIC_IPS

This is comma separated list of IP addresses to which buckets created on this Obstor instance will resolve to. For example,
a bucket `bucket1` created on current Obstor instance will be accessible as `bucket1.domain.com`, and the DNS entry for
`bucket1.domain.com` will point to IP address set in `OBSTOR_PUBLIC_IPS`.

*Note*

- This field is mandatory for standalone and erasure code Obstor server deployments, to enable federated mode.
- This field is optional for distributed deployments. If you don't set this field in a federated setup, we use the IP addresses of
hosts passed to the Obstor server startup and use them for DNS entries.

### Run Multiple Clusters

> cluster1

```sh
export OBSTOR_ETCD_ENDPOINTS="http://remote-etcd1:2379,http://remote-etcd2:4001"
export OBSTOR_DOMAIN=domain.com
export OBSTOR_PUBLIC_IPS=44.35.2.1,44.35.2.2,44.35.2.3,44.35.2.4
obstor server http://rack{1...4}.host{1...4}.domain.com/mnt/export{1...32}
```

> cluster2

```sh
export OBSTOR_ETCD_ENDPOINTS="http://remote-etcd1:2379,http://remote-etcd2:4001"
export OBSTOR_DOMAIN=domain.com
export OBSTOR_PUBLIC_IPS=44.35.1.1,44.35.1.2,44.35.1.3,44.35.1.4
obstor server http://rack{5...8}.host{5...8}.domain.com/mnt/export{1...32}
```

In this configuration you can see `OBSTOR_ETCD_ENDPOINTS` points to the etcd backend which manages Obstor's
`config.json` and bucket DNS SRV records. `OBSTOR_DOMAIN` indicates the domain suffix for the bucket which
will be used to resolve bucket through DNS. For example if you have a bucket such as `mybucket`, the
client can use now `mybucket.domain.com` to directly resolve itself to the right cluster. `OBSTOR_PUBLIC_IPS`
points to the public IP address where each cluster might be accessible, this is unique for each cluster.

NOTE: `mybucket` only exists on one cluster either `cluster1` or `cluster2` this is random and
is decided by how `domain.com` gets resolved, if there is a round-robin DNS on `domain.com` then
it is randomized which cluster might provision the bucket.

### 3. Upgrading to `etcdv3` API

Users running Obstor federation from release `RELEASE.2026-06-09T03-43-35Z` to `RELEASE.2026-07-10T01-42-11Z`, should migrate the existing bucket data on etcd server to `etcdv3` API, and update CoreDNS version to `1.2.0` before updating their Obstor server to the latest version.

Here is some background on why this is needed - Obstor server release `RELEASE.2026-06-09T03-43-35Z` to `RELEASE.2026-07-10T01-42-11Z` used etcdv2 API to store bucket data to etcd server. This was due to `etcdv3` support not available for CoreDNS server. So, even if Obstor used `etcdv3` API to store bucket data, CoreDNS wouldn't be able to read and serve it as DNS records.

Now that CoreDNS [supports etcdv3](https://coredns.io/2018/07/11/coredns-1.2.0-release/), Obstor server uses `etcdv3` API to store bucket data to etcd server. As `etcdv2` and `etcdv3` APIs are not compatible, data stored using `etcdv2` API is not visible to the `etcdv3` API. So, bucket data stored by previous Obstor version will not be visible to current Obstor version, until a migration is done.

CoreOS team has documented the steps required to migrate existing data from `etcdv2` to `etcdv3` in [this blog post](https://coreos.com/blog/migrating-applications-etcd-v3.html). Please refer the post and migrate etcd data to `etcdv3` API.

### 4. Test your setup

To test this setup, access the Obstor server via browser or `mc`. You’ll see the uploaded files are accessible from the all the Obstor endpoints.

# Explore Further

- Use `mc` with Obstor Server
- Use `aws-cli` with Obstor Server
- Use `s3cmd` with Obstor Server
- Use `minio-go` SDK with Obstor Server
- [The Obstor documentation website](/docs)
