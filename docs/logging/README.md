# Obstor Logging Quickstart Guide
This document explains how to configure Obstor server to log to different logging targets.

## Log Targets
Obstor supports currently two target types

- console
- http

### Console Target
Console target is on always and cannot be disabled.

### HTTP Target
HTTP target logs to a generic HTTP endpoint in JSON format and is not enabled by default. To enable HTTP target logging you configure the corresponding `OBSTOR_LOGGER_WEBHOOK_*` environment variables before starting the server.

```bash
export OBSTOR_LOGGER_WEBHOOK_ENABLE_target1="on"
export OBSTOR_LOGGER_WEBHOOK_AUTH_TOKEN_target1="token"
export OBSTOR_LOGGER_WEBHOOK_ENDPOINT_target1=http://localhost:8080/obstor/logs
obstor server /mnt/data
```

NOTE: `http://localhost:8080/obstor/logs` is a placeholder value to indicate the URL format, please change this accordingly as per your configuration.

## Audit Targets
To enable audit logging to an HTTP target you configure the corresponding `OBSTOR_AUDIT_WEBHOOK_*` environment variables before starting the server.

```bash
export OBSTOR_AUDIT_WEBHOOK_ENABLE_target1="on"
export OBSTOR_AUDIT_WEBHOOK_AUTH_TOKEN_target1="token"
export OBSTOR_AUDIT_WEBHOOK_ENDPOINT_target1=http://localhost:8080/obstor/logs
export OBSTOR_AUDIT_WEBHOOK_CLIENT_CERT="/tmp/cert.pem"
export OBSTOR_AUDIT_WEBHOOK_CLIENT_KEY=="/tmp/key.pem"
obstor server /mnt/data
```

Setting this environment variable automatically enables audit logging to the HTTP target. The audit logging is in JSON format as described below.

NOTE:
- `timeToFirstByte` and `timeToResponse` will be expressed in Nanoseconds.
- Additionally in the case of the erasure coded setup `tags.objectErasureMap` provides per object details about
   - Pool number the object operation was performed on.
   - Set number the object operation was performed on.
   - The list of disks participating in this operation belong to the set.

```json
{
  "version": "1",
  "deploymentid": "bc0e4d1e-bacc-42eb-91ad-2d7f3eacfa8d",
  "time": "2026-08-12T21:34:37.187817748Z",
  "api": {
    "name": "PutObject",
    "bucket": "testbucket",
    "object": "hosts",
    "status": "OK",
    "statusCode": 200,
    "timeToFirstByte": "366333ns",
    "timeToResponse": "16438202ns"
  },
  "remotehost": "127.0.0.1",
  "requestID": "15BA4A72C0C70AFC",
  "userAgent": "rclone/v1.74.3",
  "requestHeader": {
    "Authorization": "AWS4-HMAC-SHA256 Credential=obstor/20260812/us-east-1/s3/aws4_request,SignedHeaders=host;x-amz-content-sha256;x-amz-date;x-amz-decoded-content-length,Signature=d3f02a6aeddeb29b06e1773b6a8422112890981269f2463a26f307b60423177c",
    "Content-Length": "686",
    "Content-Type": "application/octet-stream",
    "User-Agent": "rclone/v1.74.3",
    "X-Amz-Content-Sha256": "STREAMING-AWS4-HMAC-SHA256-PAYLOAD",
    "X-Amz-Date": "20260812T213437Z",
    "X-Amz-Decoded-Content-Length": "512"
  },
  "responseHeader": {
    "Accept-Ranges": "bytes",
    "Content-Length": "0",
    "Content-Security-Policy": "block-all-mixed-content",
    "ETag": "a414c889dc276457bd7175f974332cb0-1",
    "Server": "Obstor/DEVELOPMENT.2026-08-12T21-28-07Z",
    "Vary": "Origin",
    "X-Amz-Request-Id": "15BA4A72C0C70AFC",
    "X-Xss-Protection": "1; mode=block"
  },
  "tags": {
    "objectErasureMap": {
      "object": {
        "poolId": 1,
        "setId": 10,
        "disks": [
          "http://server01/mnt/pool1/disk01",
          "http://server02/mnt/pool1/disk02",
          "http://server03/mnt/pool1/disk03",
          "http://server04/mnt/pool1/disk04"
        ]
     }
  }
}
```

## Explore Further
* Obstor Quickstart Guide
* [Configure Obstor Server with TLS](/docs/tls)
