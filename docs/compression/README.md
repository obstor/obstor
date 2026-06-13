# Compression Guide

Obstor server allows streaming compression to ensure efficient disk space usage.
Compression happens inflight, i.e objects are compressed before being written to disk(s).
Obstor uses [`klauspost/compress/s2`](https://github.com/klauspost/compress/tree/master/s2)
streaming compression due to its stability and performance.

This algorithm is specifically optimized for machine generated content.
Write throughput is typically at least 500MB/s per CPU core,
and scales with the number of available CPU cores.
Decompression speed is typically at least 1GB/s.

This means that in cases where raw IO is below these numbers
compression will not only reduce disk usage but also help increase system throughput.
Typically, enabling compression on spinning disk systems
will increase speed when the content can be compressed.

## Get Started

### 1. Prerequisites

Install Obstor - Obstor Quickstart Guide.

### 2. Run Obstor with compression

Compression is enabled and tuned through environment variables read by the Obstor server.
The `compress` settings take extensions and mime-types to be compressed.

The default configuration includes most common compressed file extensions and mime-types:

```bash
export OBSTOR_COMPRESS_EXTENSIONS=".txt,.log,.csv,.json,.tar,.xml,.bin"
export OBSTOR_COMPRESS_MIME_TYPES="text/*,application/json,application/xml"
```

To compress a specific set of types, set the extensions and mime-types you want.

```bash
export OBSTOR_COMPRESS_EXTENSIONS=".pdf"
export OBSTOR_COMPRESS_MIME_TYPES="application/pdf"
```

To enable compression for all content, no matter the extension and content type
(except for the default excluded types) set BOTH extensions and mime types to empty.

```bash
export OBSTOR_COMPRESS="on"
export OBSTOR_COMPRESS_EXTENSIONS=""
export OBSTOR_COMPRESS_MIME_TYPES=""
```

Set these environment variables before starting the Obstor server. Restart the server for changes to take effect.

```bash
export OBSTOR_COMPRESS="on"
export OBSTOR_COMPRESS_EXTENSIONS=".txt,.log,.csv,.json,.tar,.xml,.bin"
export OBSTOR_COMPRESS_MIME_TYPES="text/*,application/json,application/xml"
```

### 3. Compression + Encryption

Combining encryption and compression is not safe in all setups.
This is particularly so if the compression ratio of your content reveals information about it.
See [CRIME TLS](https://en.wikipedia.org/wiki/CRIME) as an example of this.

Therefore, compression is disabled when encrypting by default, and must be enabled separately.

Consult our security experts on [SUBNET](https://pgg.net/pricing) to help you evaluate if
your setup can use this feature combination safely.

To enable compression+encryption set the environment variable:

```bash
export OBSTOR_COMPRESS_ALLOW_ENCRYPTION=on
```

### 4. Excluded Types

- Already compressed objects are not fit for compression since they do not have compressible patterns.
Such objects do not produce efficient [`LZ compression`](https://en.wikipedia.org/wiki/LZ77_and_LZ78)
which is a fitness factor for a lossless data compression.

Pre-compressed input typically compresses in excess of 2GiB/s per core,
so performance impact should be minimal even if precompressed data is re-compressed.
Decompressing incompressible data has no significant performance impact.

Below is a list of common files and content-types which are typically not suitable for compression.

    - Extensions

      | `gz` | (GZIP)
      | `bz2` | (BZIP2)
      | `rar` | (WinRAR)
      | `zip` | (ZIP)
      | `7z` | (7-Zip)
      | `xz` | (LZMA)
      | `mp4` | (MP4)
      | `mkv` | (MKV media)
      | `mov` | (MOV)

    - Content-Types

      | `video/*` |
      | `audio/*` |
      | `application/zip` |
      | `application/x-gzip` |
      | `application/zip` |
      | `application/x-bz2` |
      | `application/x-compress` |
      | `application/x-xz` |

All files with these extensions and mime types are excluded from compression,
even if compression is enabled for all types.

### 5. Notes

- Obstor does not support compression for Backend (Azure/GCS/NAS) implementations.

## To test the setup

To test this setup, upload objects with an S3 client such as `rclone copy` or `aws s3 cp`,
then inspect the data directory on disk to view the stored size of the object.

## Explore Further

- Use `rclone` with Obstor Server
- Use `aws-cli` with Obstor Server
- Use `s3cmd` with Obstor Server
- [The Obstor documentation website](/docs)
