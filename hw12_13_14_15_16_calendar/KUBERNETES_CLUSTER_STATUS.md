# Статус Kubernetes кластера

## Проверка кластера

### 1. Статус узлов кластера
```bash
kubectl get nodes
```

**Результат:**
- Кластер: minikube
- Статус: Ready
- Роль: control-plane
- Версия Kubernetes: v1.31.0

### 2. Статус Minikube
```bash
minikube status
```

**Результат:**
- host: Running
- kubelet: Running
- apiserver: Running
- kubeconfig: Configured

### 3. Все ресурсы в namespace calendar
```bash
kubectl get all -n calendar
```

**Результат:**
```
NAME                            READY   STATUS    RESTARTS   AGE
pod/postgres-5f687656c9-6mkkg   1/1     Running   0          33m
pod/rabbitmq-7cdd6bd97f-9sjwt   1/1     Running   0          32m

NAME               TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)              AGE
service/postgres   ClusterIP   10.107.171.44   <none>        5432/TCP             33m
service/rabbitmq   ClusterIP   10.97.28.49     <none>        5672/TCP,15672/TCP   32m

NAME                       READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/postgres   1/1     1            1           33m
deployment.apps/rabbitmq   1/1     1            1           32m
```

### 4. Версии инструментов
- **Helm:** v3.19.2
- **Kubernetes:** v1.31.0
- **Minikube:** v1.34.0

## Вывод

✅ Kubernetes кластер развернут и работает
✅ Все компоненты (host, kubelet, apiserver) в состоянии Running
✅ Зависимости приложения (PostgreSQL, RabbitMQ) успешно развернуты
