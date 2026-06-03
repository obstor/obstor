# Bucket Quota Configuration Quickstart Guide

![quota](https://raw.githubusercontent.com/cloudment/obstor/main/docs/bucket/quota/bucket-quota.png)

Buckets can be configured to have one of two types of quota configuration - FIFO and Hard quota.

- `Hard` quota disallows writes to the bucket after configured quota limit is reached.
- `FIFO` quota automatically deletes oldest content until bucket usage falls within configured limit while permitting writes.

> NOTE: Bucket quotas are not supported under backend or standalone single disk deployments.

## Prerequisites
- Install Obstor - Obstor Quickstart Guide.
- Use `mc` with Obstor Server

## Set bucket quota configuration

### Set a hard quota of 1GB for a bucket `mybucket` on Obstor object storage:

```sh
$ mc admin bucket quota myobstor/mybucket --hard 1gb
```

### Set FIFO quota of 5GB for a bucket "mybucket" on Obstor to allow automatic deletion of older content to ensure bucket usage remains within 5GB

```sh
$ mc admin bucket quota myobstor/mybucket --fifo 5gb
```

### Verify the quota configured on `mybucket` on Obstor

```sh
$ mc admin bucket quota myobstor/mybucket
```

### Clear bucket quota configuration for `mybucket` on Obstor

```sh
$ mc admin bucket quota myobstor/mybucket --clear
```
