# KMS Guide

Obstor uses a key-management-system (KMS) to support SSE-S3. If a client requests SSE-S3, or auto-encryption is enabled, the Obstor server encrypts each object with an unique object key which is protected by a master key managed by the KMS.

## Quick Start

Obstor supports multiple KMS implementations via our [KES](https://github.com/minio/kes#kes) project. We run a KES instance at `https://play.obstor.net:7373` for you to experiment and quickly get started. To run Obstor with a KMS just fetch the root identity, set the following environment variables and then start your Obstor server. If you havn't installed Obstor, yet, then follow the Obstor install instructions first.

#### 1. Fetch the root identity
As the initial step, fetch the private key and certificate of the root identity:

```bash
curl -sSL --tlsv1.2 \
     -O 'https://raw.githubusercontent.com/minio/kes/master/root.key' \
     -O 'https://raw.githubusercontent.com/minio/kes/master/root.cert'
```

#### 2. Set the Obstor-KES configuration

```bash
export OBSTOR_KMS_KES_ENDPOINT=https://play.obstor.net:7373
export OBSTOR_KMS_KES_KEY_FILE=root.key
export OBSTOR_KMS_KES_CERT_FILE=root.cert
export OBSTOR_KMS_KES_KEY_NAME=my-obstor-key
```

#### 3. Start the Obstor Server

```bash
export OBSTOR_ROOT_USER=obstor
export OBSTOR_ROOT_PASSWORD=obstor123
obstor server ~/export
```

> The KES instance at `https://play.obstor.net:7373` is meant to experiment and provides a way to get started quickly.
> Note that anyone can access or delete master keys at `https://play.obstor.net:7373`. You should run your own KES
> instance in production.

## Configuration Guides

A typical Obstor deployment that uses a KMS for SSE-S3 looks like this:
```
    ┌────────────┐
    │ ┌──────────┴─┬─────╮          ┌────────────┐
    └─┤ ┌──────────┴─┬───┴──────────┤ ┌──────────┴─┬─────────────────╮
      └─┤ ┌──────────┴─┬─────┬──────┴─┤ KES Server ├─────────────────┤
        └─┤   Obstor   ├─────╯        └────────────┘            ┌────┴────┐
          └────────────┘                                        │   KMS   │
                                                                └─────────┘
```

In a given setup, there are `n` Obstor instances talking to `m` KES servers but only `1` central KMS. The most simple setup consists of `1` Obstor server or cluster talking to `1` KMS via `1` KES server.

The main difference between various Obstor-KMS deployments is the KMS implementation. The following table helps you select the right option for your use case:

| KMS                                                                                          | Purpose                                                           |
|:---------------------------------------------------------------------------------------------|:------------------------------------------------------------------|
| [Hashicorp Vault](https://github.com/minio/kes/wiki/Hashicorp-Vault-Keystore)                | Local KMS. Obstor and KMS on-prem (**Recommended**)                |
| [AWS-KMS + SecretsManager](https://github.com/minio/kes/wiki/AWS-SecretsManager)             | Cloud KMS. Obstor in combination with a managed KMS installation   |
| [Gemalto KeySecure /Thales CipherTrust](https://github.com/minio/kes/wiki/Gemalto-KeySecure) | Local KMS. Obstor and KMS On-Premise.                             |
| [Google Cloud Platform SecretManager](https://github.com/minio/kes/wiki/GCP-SecretManager)   | Cloud KMS. Obstor in combination with a managed KMS installation   |
| [FS](https://github.com/minio/kes/wiki/Filesystem-Keystore)                                  | Local testing or development (**Not recommended for production**) |


The Obstor-KES configuration is always the same - regardless of the underlying KMS implementation. Checkout the Obstor-KES [configuration example](https://github.com/minio/kes/wiki/Obstor-Object-Storage).

### Further references

- [Run Obstor with TLS / HTTPS](/docs/tls)
- [Tweak the KES server configuration](https://github.com/minio/kes/wiki/Configuration)
- [Run a load balancer infront of KES](https://github.com/minio/kes/wiki/TLS-Proxy)
- [Understand the KES server concepts](https://github.com/minio/kes/wiki/Concepts)

## Auto Encryption
Auto-Encryption is useful when Obstor administrator wants to ensure that all data stored on Obstor is encrypted at rest.

### Using `mc encrypt` (recommended)
Obstor automatically encrypts all objects on buckets if KMS is successfully configured and bucket encryption configuration is enabled for each bucket as shown below:
```bash
mc encrypt set sse-s3 myobstor/bucket/
```

Verify if Obstor has `sse-s3` enabled
```bash
mc encrypt info myobstor/bucket/
Auto encryption 'sse-s3' is enabled
```

### Using environment (deprecated)
> NOTE: The following ENV might be removed in future, you are advised to move to the previously recommended approach using `mc encrypt`. S3 backend supports encryption at backend layer which may  be dropped in favor of simplicity at a later time. It is advised that S3 backend users migrate to Obstor server mode or enable encryption at REST at the backend.

Obstor automatically encrypts all objects on buckets if KMS is successfully configured and following ENV is enabled:
```bash
export OBSTOR_KMS_AUTO_ENCRYPTION=on
```

### Verify auto-encryption
> Note that auto-encryption only affects requests without S3 encryption headers. So, if a S3 client sends
> e.g. SSE-C headers, Obstor will encrypt the object with the key sent by the client and won't reach out to
> the configured KMS.

To verify auto-encryption, use the following `mc` command:

```
mc cp test.file myobstor/bucket/
test.file:   5 B / 5 B  ┃███████████████████████████████████┃  100.00% 337 B/s 0s
```

```bash
mc stat myobstor/bucket/test.file
Name      : test.file
...
Encrypted :
  X-Amz-Server-Side-Encryption: AES256
```

## Explore Further

- Use `mc` with Obstor Server
- Use `aws-cli` with Obstor Server
- Use `s3cmd` with Obstor Server
- Use `minio-go` SDK with Obstor Server
- [The Obstor documentation website](/docs)
