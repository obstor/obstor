# Bucket Lifecycle Configuration Quickstart Guide

Enable object lifecycle configuration on buckets to setup automatic deletion of objects after a specified number of days or a specified date.

## 1. Prerequisites
- Install Obstor - Obstor Quickstart Guide.
- Install the AWS CLI - [Installing AWS Command Line Interface](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html)

## 2. Enable bucket lifecycle configuration

- Create a bucket lifecycle configuration which expires the objects under the prefix `old/` on `2026-01-01T00:00:00.000Z` date and the objects under `temp/` after 7 days.
- Save the configuration to a file `lifecycle.json`:

```json
{
  "Rules": [
    {
      "Expiration": {
        "Date": "2026-01-01T00:00:00.000Z"
      },
      "ID": "OldPictures",
      "Filter": {
        "Prefix": "old/"
      },
      "Status": "Enabled"
    },
    {
      "Expiration": {
        "Days": 7
      },
      "ID": "TempUploads",
      "Filter": {
        "Prefix": "temp/"
      },
      "Status": "Enabled"
    }
  ]
}
```

- Apply the bucket lifecycle configuration using the AWS CLI:

```bash
$ aws --endpoint-url http://localhost:9000 s3api put-bucket-lifecycle-configuration \
  --bucket testbucket \
  --lifecycle-configuration file://lifecycle.json
```

- List the current settings
```bash
$ aws --endpoint-url http://localhost:9000 s3api get-bucket-lifecycle-configuration \
  --bucket testbucket
```

## 3. Activate ILM versioning features

This will only work with a versioned bucket, take a look at [Bucket Versioning Guide](/docs/bucket/versioning) for more understanding.

### 3.1 Automatic removal of non current objects versions

A non-current object version is a version which is not the latest for a given object. It is possible to set up an automatic removal of non-current versions when a version becomes older than a given number of days.

e.g., To scan objects stored under `user-uploads/` prefix and remove versions older than one year.
```json
{
  "Rules": [
    {
      "ID": "Removing all old versions",
      "Filter": {
        "Prefix": "users-uploads/"
      },
      "NoncurrentVersionExpiration": {
        "NoncurrentDays": 365
      },
      "Status": "Enabled"
    }
  ]
}
```

### 3.2 Automatic removal of delete markers with no other versions

When an object has only one version as a delete marker, the latter can be automatically removed after a certain number of days using the following configuration:

```json
{
  "Rules": [
    {
      "ID": "Removing all delete markers",
      "Expiration": {
        "DeleteMarker": true
      },
      "Status": "Enabled"
    }
  ]
}
```

## Explore Further
- [Obstor | Golang Client API Reference](/docs/bucket/lifecycle)
- [Object Lifecycle Management](https://docs.aws.amazon.com/AmazonS3/latest/dev/object-lifecycle-mgmt.html)
