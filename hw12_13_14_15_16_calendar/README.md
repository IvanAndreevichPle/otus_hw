# Calendar Application

Calendar application with API, Scheduler, and Sender services.

## Kubernetes Deployment

This project includes Helm charts for deploying the Calendar application to Kubernetes.

### Prerequisites

- Kubernetes cluster (minikube/k3s/microk8s)
- Helm 3.x installed
- Docker images built and pushed to registry

### Installation

1. Build Docker images:
```bash
make build
docker tag calendar-app:latest <registry>/calendar-app:latest
docker tag calendar-scheduler:latest <registry>/calendar-scheduler:latest
docker tag calendar-sender:latest <registry>/calendar-sender:latest
docker push <registry>/calendar-app:latest
docker push <registry>/calendar-scheduler:latest
docker push <registry>/calendar-sender:latest
```

2. Update `values.yaml` with your image registry:
```yaml
global:
  imageRegistry: "<registry>"

calendar:
  image:
    repository: calendar-app
    tag: "latest"

scheduler:
  image:
    repository: calendar-scheduler
    tag: "latest"

sender:
  image:
    repository: calendar-sender
    tag: "latest"
```

3. Install the chart:
```bash
helm install calendar-release ./
```

### Configuration

Edit `values.yaml` to customize:
- Replica counts
- Resource limits
- Database and RabbitMQ connection settings
- Ingress configuration

### Services

The chart deploys:
- **Calendar API**: HTTP API service (port 8080)
- **Scheduler**: Event notification scheduler
- **Sender**: Notification sender service

### Ingress

To enable Ingress, set in `values.yaml`:
```yaml
calendar:
  ingress:
    enabled: true
    className: "nginx"
    hosts:
      - host: calendar.local
        paths:
          - path: /
            pathType: Prefix
```

### Dependencies

The application requires:
- PostgreSQL database
- RabbitMQ message queue

These should be deployed separately or as Helm chart dependencies.
