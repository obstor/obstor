# Deploy Obstor on Kubernetes

Obstor is a high performance distributed object storage server, designed for large-scale private cloud infrastructure. Obstor is designed in a cloud-native manner to scale sustainably in multi-tenant environments. Orchestration platforms like Kubernetes provide perfect cloud-native environment to deploy and scale Obstor.

## Obstor Deployment on Kubernetes

There are multiple options to deploy Obstor on Kubernetes:

- Obstor-Operator: Operator offers seamless way to create and update highly available distributed Obstor clusters. Refer [Obstor Operator documentation](https://github.com/obstor/obstor-operator/blob/master/README.md) for more details.

- Helm Chart: Obstor Helm Chart offers customizable and easy Obstor deployment with a single command.

## Monitoring Obstor in Kubernetes

Obstor server exposes un-authenticated liveness endpoints so Kubernetes can natively identify unhealthy Obstor containers. Obstor also exposes Prometheus compatible data on a different endpoint to enable Prometheus users to natively monitor their Obstor deployments.

## Explore Further

- [Obstor Erasure Code QuickStart Guide](/docs/erasure)
- [Kubernetes Documentation](https://kubernetes.io/docs/home/)
- [Helm package manager for kubernetes](https://helm.sh/)
