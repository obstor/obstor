# Deploy Obstor on Docker Compose

Docker Compose allows defining and running single host, multi-container Docker applications.

With Compose, you use a Compose file to configure Obstor services. Then, using a single command, you can create and launch all the Distributed Obstor instances from your configuration. Distributed Obstor instances will be deployed in multiple containers on the same host. This is a great way to set up development, testing, and staging environments, based on Distributed Obstor.

## 1. Prerequisites

* Familiarity with [Docker Compose](https://docs.docker.com/compose/overview/).
* Docker installed on your machine. Download the relevant installer from [here](https://www.docker.com/community-edition#/download).

## 2. Run Distributed Obstor on Docker Compose

To deploy Distributed Obstor on Docker Compose, please download [docker-compose.yaml](https://github.com/obstor/obstor/blob/main/docs/orchestration/docker-compose/docker-compose.yaml?raw=true) and [nginx.conf](https://github.com/obstor/obstor/blob/main/docs/orchestration/docker-compose/nginx.conf?raw=true) to your current working directory. Note that Docker Compose pulls the Obstor Docker image, so there is no need to explicitly download Obstor binary. Then run one of the below commands

### GNU/Linux and macOS

```sh
docker-compose pull
docker-compose up
```

### Windows

```sh
docker-compose.exe pull
docker-compose.exe up
```

Distributed instances are now accessible on the host at ports 9000, proceed to access the Web browser at http://127.0.0.1:9000/. Here 4 Obstor server instances are reverse proxied through Nginx load balancing.

### Notes

* By default the Docker Compose file uses the Docker image for latest Obstor server release. You can change the image tag to pull a specific [Obstor Docker image](https://ghcr.io/cloudment/obstor).

* There are 4 obstor distributed instances created by default. You can add more Obstor services (up to total 16) to your Obstor Compose deployment. To add a service
  * Replicate a service definition and change the name of the new service appropriately.
  * Update the command section in each service.
  * Add a new Obstor server instance to the upstream directive in the Nginx configuration file.

  Read more about distributed Obstor [here](/docs/distributed).

### Explore Further
- [Overview of Docker Compose](https://docs.docker.com/compose/overview/)
- [Obstor Docker Quickstart Guide](/docs/docker)
- [Deploy Obstor on Docker Swarm](/docs/orchestration/docker-swarm)
- [Obstor Erasure Code QuickStart Guide](/docs/erasure)
