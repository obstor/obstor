# Obstor Docker Quickstart Guide

## Prerequisites
Docker installed on your machine. Download the relevant installer from [here](https://www.docker.com/community-edition#/download).

## Run Standalone Obstor on Docker.
Obstor needs a persistent volume to store configuration and application data. However, for testing purposes, you can launch Obstor by simply passing a directory (`/data` in the example below). This directory gets created in the container filesystem at the time of container start. But all the data is lost after container exits.

```bash
docker run -p 9000:9000 \
  -e "OBSTOR_ROOT_USER=AKIAIOSFODNN7EXAMPLE" \
  -e "OBSTOR_ROOT_PASSWORD=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" \
  ghcr.io/obstor/obstor server /data
```

To create a Obstor container with persistent storage, you need to map local persistent directories from the host OS to virtual config `~/.obstor` and export `/data` directories. To do this, run the below commands

#### GNU/Linux and macOS
```bash
docker run -p 9000:9000 \
  --name obstor1 \
  -v /mnt/data:/data \
  -e "OBSTOR_ROOT_USER=AKIAIOSFODNN7EXAMPLE" \
  -e "OBSTOR_ROOT_PASSWORD=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" \
  ghcr.io/obstor/obstor server /data
```

#### Windows
```bash
docker run -p 9000:9000 \
  --name obstor1 \
  -v D:\data:/data \
  -e "OBSTOR_ROOT_USER=AKIAIOSFODNN7EXAMPLE" \
  -e "OBSTOR_ROOT_PASSWORD=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" \
  ghcr.io/obstor/obstor server /data
```

## Run Distributed Obstor on Docker
Distributed Obstor can be deployed via [Docker Compose](/docs/orchestration/docker-compose) or [Swarm mode](/docs/orchestration/docker-swarm). The major difference between these two being, Docker Compose creates a single host, multi-container deployment, while Swarm mode creates a multi-host, multi-container deployment.

This means Docker Compose lets you quickly get started with Distributed Obstor on your computer - ideal for development, testing, staging environments. While deploying Distributed Obstor on Swarm offers a more robust, production level deployment.

## Obstor Docker Tips

### Obstor Custom Access and Secret Keys
To override Obstor's auto-generated keys, you may pass secret and access keys explicitly as environment variables. Obstor server also allows regular strings as access and secret keys.

#### GNU/Linux and macOS
```bash
docker run -p 9000:9000 --name obstor1 \
  -e "OBSTOR_ROOT_USER=AKIAIOSFODNN7EXAMPLE" \
  -e "OBSTOR_ROOT_PASSWORD=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" \
  -v /mnt/data:/data \
  ghcr.io/obstor/obstor server /data
```

#### Windows
```bash
docker run -p 9000:9000 --name obstor1 \
  -e "OBSTOR_ROOT_USER=AKIAIOSFODNN7EXAMPLE" \
  -e "OBSTOR_ROOT_PASSWORD=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" \
  -v D:\data:/data \
  ghcr.io/obstor/obstor server /data
```

### Run Obstor Docker as a regular user
Docker provides standardized mechanisms to run docker containers as non-root users.

#### GNU/Linux and macOS
On Linux and macOS you can use `--user` to run the container as regular user.

> NOTE: make sure --user has write permission to *${HOME}/data* prior to using `--user`.
```bash
mkdir -p ${HOME}/data
docker run -p 9000:9000 \
  --user $(id -u):$(id -g) \
  --name obstor1 \
  -e "OBSTOR_ROOT_USER=AKIAIOSFODNN7EXAMPLE" \
  -e "OBSTOR_ROOT_PASSWORD=wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY" \
  -v ${HOME}/data:/data \
  ghcr.io/obstor/obstor server /data
```

#### Windows
On windows you would need to use Docker integrated windows authentication and [Create a container with Active Directory Support](https://blogs.msdn.microsoft.com/containerstuff/2017/01/30/create-a-container-with-active-directory-support/)

> NOTE: make sure your AD/Windows user has write permissions to *D:\data* prior to using `credentialspec=`.

```bash
docker run -p 9000:9000 \
  --name obstor1 \
  --security-opt "credentialspec=file://myuser.json"
  -e "OBSTOR_ROOT_USER=AKIAIOSFODNN7EXAMPLE" \
  -e "OBSTOR_ROOT_PASSWORD=wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY" \
  -v D:\data:/data \
  ghcr.io/obstor/obstor server /data
```

### Obstor Custom Access and Secret Keys using Docker secrets
To override Obstor's auto-generated keys, you may pass secret and access keys explicitly by creating access and secret keys as [Docker secrets](https://docs.docker.com/engine/swarm/secrets/). Obstor server also allows regular strings as access and secret keys.

```bash
echo "AKIAIOSFODNN7EXAMPLE" | docker secret create access_key -
echo "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" | docker secret create secret_key -
```

Create a Obstor service using `docker service` to read from Docker secrets.
```bash
docker service create --name="obstor-service" --secret="access_key" --secret="secret_key" ghcr.io/obstor/obstor server /data
```

Read more about `docker service` [here](https://docs.docker.com/engine/swarm/how-swarm-mode-works/services/)

#### Obstor Custom Access and Secret Key files
To use other secret names follow the instructions above and replace `access_key` and `secret_key` with your custom names (e.g. `my_secret_key`,`my_custom_key`). Run your service with
```bash
docker service create --name="obstor-service" \
  --secret="my_access_key" \
  --secret="my_secret_key" \
  --env="OBSTOR_ROOT_USER_FILE=my_access_key" \
  --env="OBSTOR_ROOT_PASSWORD_FILE=my_secret_key" \
  ghcr.io/obstor/obstor server /data
```
`OBSTOR_ROOT_USER_FILE` and `OBSTOR_ROOT_PASSWORD_FILE` also support custom absolute paths, in case Docker secrets are mounted to custom locations or other tools are used to mount secrets into the container. For example, HashiCorp Vault injects secrets to `/vault/secrets`. With the custom names above, set the environment variables to
```bash
OBSTOR_ROOT_USER_FILE=/vault/secrets/my_access_key
OBSTOR_ROOT_PASSWORD_FILE=/vault/secrets/my_secret_key
```

### Retrieving Container ID
To use Docker commands on a specific container, you need to know the `Container ID` for that container. To get the `Container ID`, run

```bash
docker ps -a
```

`-a` flag makes sure you get all the containers (Created, Running, Exited). Then identify the `Container ID` from the output.

### Starting and Stopping Containers
To start a stopped container, you can use the [`docker start`](https://docs.docker.com/engine/reference/commandline/start/) command.

```bash
docker start <container_id>
```

To stop a running container, you can use the [`docker stop`](https://docs.docker.com/engine/reference/commandline/stop/) command.
```bash
docker stop <container_id>
```

### Obstor container logs
To access Obstor logs, you can use the [`docker logs`](https://docs.docker.com/engine/reference/commandline/logs/) command.

```bash
docker logs <container_id>
```

### Monitor Obstor Docker Container
To monitor the resources used by Obstor container, you can use the [`docker stats`](https://docs.docker.com/engine/reference/commandline/stats/) command.

```bash
docker stats <container_id>
```

## Explore Further

* [Deploy Obstor on Docker Compose](/docs/orchestration/docker-compose)
* [Deploy Obstor on Docker Swarm](/docs/orchestration/docker-swarm)
* [Distributed Obstor Quickstart Guide](/docs/distributed)
* [Obstor Erasure Code QuickStart Guide](/docs/erasure)
