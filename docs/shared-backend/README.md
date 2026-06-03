# Shared Backend Obstor Quickstart Guide

Obstor shared mode lets you use single [NAS](https://en.wikipedia.org/wiki/Network-attached_storage) (like NFS, GlusterFS, and other
distributed filesystems) as the storage backend for multiple Obstor servers. Synchronization among Obstor servers is taken care by design.
Read more about the Obstor shared mode design [here](/docs/shared-backend/DESIGN).

Obstor shared mode is developed to solve several real world use cases, without any special configuration changes. Some of these are

- You have already invested in NAS and would like to use Obstor to add S3 and [other protocol](/docs/protocols) compatibility to your storage tier.
- You need to use NAS with an S3 interface due to your application architecture requirements.
- You expect huge traffic and need a load balanced S3-compatible server, serving files from a single NAS backend.

With a proxy running in front of multiple, shared mode Obstor servers, it is very easy to create a Highly Available, load balanced, AWS S3-compatible storage system.

# Get started

If you're aware of stand-alone Obstor set up, the installation and running remains the same.

## 1. Prerequisites

Install Obstor - Obstor Quickstart Guide.

## 2. Run Obstor on Shared Backend

To run Obstor shared backend instances, you need to start multiple Obstor servers pointing to the same backend storage. We'll see examples on how to do this in the following sections.

*Note*

- All the nodes running shared Obstor need to have same access key and secret key. To achieve this, we export access key and secret key as environment variables on all the nodes before executing Obstor server command.
- The drive paths below are for demonstration purposes only, you need to replace these with the actual drive paths/folders.

#### Obstor shared mode on Ubuntu 26.04 LTS.

You'll need the path to the shared volume, e.g. `/path/to/nfs-volume`. Then run the following commands on all the nodes you'd like to launch Obstor.

```sh
export OBSTOR_ROOT_USER=<ACCESS_KEY>
export OBSTOR_ROOT_PASSWORD=<SECRET_KEY>
obstor backend nas /path/to/nfs-volume
```

#### Obstor shared mode on Windows Server 2025

You'll need the path to the shared volume, e.g. `\\remote-server\smb`. Then run the following commands on all the nodes you'd like to launch Obstor.

```bash
set OBSTOR_ROOT_USER=my-username
set OBSTOR_ROOT_PASSWORD=my-password
obstor.exe backend nas \\remote-server\smb\export
```

*Windows Tip*

If a remote volume, e.g. `\\remote-server\smb` is mounted as a drive, e.g. `M:\`. You can use [`net use`](https://technet.microsoft.com/en-us/library/bb490717.aspx) command to map the drive to a folder.

```bash
set OBSTOR_ROOT_USER=my-username
set OBSTOR_ROOT_PASSWORD=my-password
net use m: \\remote-server\smb\export /P:Yes
obstor.exe backend nas M:\export
```

## 3. Test your setup

To test this setup, access the Obstor server via browser or `mc`. You’ll see the uploaded files are accessible from the all the Obstor shared backend endpoints.

## Explore Further
- Use `mc` with Obstor Server
- Use `aws-cli` with Obstor Server
- Use `s3cmd` with Obstor Server
- Use `minio-go` SDK with Obstor Server
- [The Obstor documentation website](/docs)
