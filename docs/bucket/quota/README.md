# Bucket Quota Configuration Quickstart Guide

![quota](https://raw.githubusercontent.com/obstor/obstor/main/docs/bucket/quota/bucket-quota.png)

Buckets can be configured to have one of two types of quota configuration - FIFO and Hard quota.

- `Hard` quota disallows writes to the bucket after configured quota limit is reached.
- `FIFO` quota automatically deletes oldest content until bucket usage falls within configured limit while permitting writes.

> NOTE: Bucket quotas are not supported under backend or standalone single disk deployments.

## Prerequisites
- Install Obstor - Obstor Quickstart Guide.

## Set bucket quota configuration

Bucket quotas are administrative settings with no S3 API option.

### Set a hard quota of 1GB for a bucket `mybucket` on Obstor object storage

Use Obstor's API or the dashboard to set a `Hard` quota of `1gb` on `mybucket`. Writes are rejected once usage reaches the limit.

### Set FIFO quota of 5GB for a bucket `mybucket` on Obstor to allow automatic deletion of older content to ensure bucket usage remains within 5GB

Use Obstor's API or the dashboard to set a `FIFO` quota of `5gb` on `mybucket`. The oldest content is deleted automatically so usage stays within the limit.

### Verify the quota configured on `mybucket` on Obstor

Use Obstor's dashboard or API to view the quota configured on `mybucket`.

### Clear bucket quota configuration for `mybucket` on Obstor

Use Obstor's API or the dashboard to clear the quota configuration on `mybucket`.
