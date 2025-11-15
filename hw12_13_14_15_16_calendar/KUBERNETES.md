# Развертывание в Kubernetes

## Структура Helm Chart

```
.
├── Chart.yaml                    # Метаданные chart
├── values.yaml                   # Дефолтные значения
└── templates/                    # Шаблоны Kubernetes манифестов
    ├── _helpers.tpl             # Вспомогательные шаблоны
    ├── calendar-deployment.yaml  # Deployment для API сервиса
    ├── calendar-service.yaml     # Service для API сервиса
    ├── scheduler-deployment.yaml # Deployment для Scheduler
    ├── sender-deployment.yaml    # Deployment для Sender
    ├── ingress.yaml             # Ingress для API сервиса
    └── configmap.yaml           # ConfigMap для конфигураций
```

## Развертывание кластера Kubernetes

### Вариант 1: Minikube

```bash
# Установка minikube
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube

# Запуск кластера
minikube start

# Проверка
kubectl get nodes
```

### Вариант 2: k3s

```bash
# Установка k3s
curl -sfL https://get.k3s.io | sh -

# Проверка
kubectl get nodes
```

### Вариант 3: MicroK8s

```bash
# Установка microk8s
sudo snap install microk8s --classic

# Добавление пользователя в группу
sudo usermod -a -G microk8s $USER
newgrp microk8s

# Включение необходимых аддонов
microk8s enable dns storage ingress

# Проверка
microk8s kubectl get nodes
```

## Подготовка образов

1. Соберите Docker образы:
```bash
make build
```

2. Загрузите образы в registry или используйте локальный registry:
```bash
# Для minikube
eval $(minikube docker-env)
docker build -f build/Dockerfile.calendar -t calendar-app:latest .
docker build -f build/Dockerfile.scheduler -t calendar-scheduler:latest .
docker build -f build/Dockerfile.sender -t calendar-sender:latest
```

## Развертывание зависимостей

Перед развертыванием приложения необходимо развернуть PostgreSQL и RabbitMQ:

```bash
# PostgreSQL
kubectl create namespace calendar
kubectl apply -f - <<EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-pvc
  namespace: calendar
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
  namespace: calendar
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:15
        env:
        - name: POSTGRES_USER
          value: calendar
        - name: POSTGRES_PASSWORD
          value: calendar
        - name: POSTGRES_DB
          value: calendar
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: postgres-storage
          mountPath: /var/lib/postgresql/data
      volumes:
      - name: postgres-storage
        persistentVolumeClaim:
          claimName: postgres-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: postgres
  namespace: calendar
spec:
  selector:
    app: postgres
  ports:
  - port: 5432
    targetPort: 5432
EOF

# RabbitMQ
kubectl apply -f - <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rabbitmq
  namespace: calendar
spec:
  replicas: 1
  selector:
    matchLabels:
      app: rabbitmq
  template:
    metadata:
      labels:
        app: rabbitmq
    spec:
      containers:
      - name: rabbitmq
        image: rabbitmq:3-management
        env:
        - name: RABBITMQ_DEFAULT_USER
          value: calendar
        - name: RABBITMQ_DEFAULT_PASS
          value: calendar
        ports:
        - containerPort: 5672
        - containerPort: 15672
---
apiVersion: v1
kind: Service
metadata:
  name: rabbitmq
  namespace: calendar
spec:
  selector:
    app: rabbitmq
  ports:
  - port: 5672
    targetPort: 5672
    name: amqp
  - port: 15672
    targetPort: 15672
    name: management
EOF
```

## Развертывание приложения через Helm

1. Обновите `values.yaml` с правильными именами образов:
```yaml
calendar:
  image:
    repository: calendar-app
    tag: "latest"
    pullPolicy: IfNotPresent  # или Never для локальных образов

scheduler:
  image:
    repository: calendar-scheduler
    tag: "latest"
    pullPolicy: IfNotPresent

sender:
  image:
    repository: calendar-sender
    tag: "latest"
    pullPolicy: IfNotPresent
```

2. Установите chart:
```bash
helm install calendar-release ./ --namespace calendar
```

3. Проверьте статус:
```bash
kubectl get pods -n calendar
kubectl get services -n calendar
kubectl get ingress -n calendar
```

4. Просмотр логов:
```bash
kubectl logs -n calendar -l app.kubernetes.io/component=calendar-api
kubectl logs -n calendar -l app.kubernetes.io/component=scheduler
kubectl logs -n calendar -l app.kubernetes.io/component=sender
```

## Включение Ingress

Для доступа к API через Ingress:

1. Обновите `values.yaml`:
```yaml
calendar:
  ingress:
    enabled: true
    className: "nginx"  # или другой ingress controller
    hosts:
      - host: calendar.local
        paths:
          - path: /
            pathType: Prefix
```

2. Обновите release:
```bash
helm upgrade calendar-release ./ --namespace calendar
```

3. Добавьте в `/etc/hosts`:
```
<INGRESS_IP> calendar.local
```

## Удаление

```bash
helm uninstall calendar-release --namespace calendar
kubectl delete namespace calendar
```

