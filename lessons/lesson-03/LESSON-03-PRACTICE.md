# Практические задания -- Урок 3: Kubernetes

Пошаговые инструкции для самостоятельного выполнения.

---

## Задание 1: Деплой манифестами (20 минут)

### Цель
Развернуть приложение в Kubernetes через обычные YAML-манифесты.

### Шаги

#### 1. Проверяем кластер

```bash
kubectl config current-context
# Должно быть: colima

kubectl get nodes
# Нода в статусе Ready
```

#### 2. Создаем namespace

```bash
kubectl apply -f k8s/manifests/namespace.yaml
kubectl get ns | grep wit2
```

#### 3. Создаем секрет для GHCR

```bash
kubectl create secret docker-registry ghcr-credentials \
  --namespace=wit2-devops \
  --docker-server=ghcr.io \
  --docker-username=YOUR_GITHUB_USERNAME \
  --docker-password=YOUR_GITHUB_TOKEN
```

#### 4. Применяем манифесты

```bash
kubectl apply -f k8s/manifests/deployment.yaml
kubectl apply -f k8s/manifests/service.yaml
kubectl apply -f k8s/manifests/ingress.yaml
```

#### 5. Ждем готовности и проверяем

```bash
# Ждем пока поды запустятся
kubectl get pods -n wit2-devops -w
# Ctrl+C когда оба в Running

# Проверяем все ресурсы
kubectl get all -n wit2-devops
```

#### 6. Тестируем

```bash
kubectl port-forward svc/go-simple-api 8080:80 -n wit2-devops &
curl http://localhost:8080/health
curl http://localhost:8080/api/version
curl http://localhost:8080/metrics
kill %1
```

### Проверка знаний

1. Что произойдет если удалить один из подов?
   <details>
   <summary>Ответ</summary>
   Deployment заметит что реплик меньше чем указано в spec.replicas и автоматически создаст новый под.
   </details>

2. Зачем нужен imagePullSecrets в deployment?
   <details>
   <summary>Ответ</summary>
   Образ находится в приватном GitHub Container Registry. Без секрета Kubernetes не сможет скачать образ и под будет в статусе ImagePullBackOff.
   </details>

3. В чем разница между port и targetPort в Service?
   <details>
   <summary>Ответ</summary>
   port (80) -- это порт самого Service, на который приходят запросы. targetPort (8080) -- это порт контейнера, куда Service перенаправляет трафик.
   </details>

---

## Задание 2: Эксперименты с Deployment (15 минут)

### Цель
Понять как Deployment управляет подами.

### Эксперимент 1: Самовосстановление

```bash
# Смотрим текущие поды
kubectl get pods -n wit2-devops

# Удаляем один из подов (подставьте реальное имя)
kubectl delete pod <ИМЯ_ПОДА> -n wit2-devops

# Сразу смотрим -- новый под создается
kubectl get pods -n wit2-devops
```

### Эксперимент 2: Масштабирование

```bash
# Увеличиваем до 4 реплик
kubectl scale deployment go-simple-api --replicas=4 -n wit2-devops
kubectl get pods -n wit2-devops
# Должно быть 4 пода

# Уменьшаем до 1
kubectl scale deployment go-simple-api --replicas=1 -n wit2-devops
kubectl get pods -n wit2-devops
# Должен остаться 1 под

# Возвращаем к 2
kubectl scale deployment go-simple-api --replicas=2 -n wit2-devops
```

### Эксперимент 3: Отладка

```bash
# Подробная информация о поде
kubectl describe pod -n wit2-devops -l app=go-simple-api

# Обратите внимание на секции:
# - Events (события: Scheduled, Pulling, Pulled, Created, Started)
# - Conditions (Ready, ContainersReady)
# - Liveness/Readiness probes

# Логи приложения
kubectl logs -n wit2-devops -l app=go-simple-api --tail=20

# Войти внутрь контейнера
kubectl exec -it -n wit2-devops deploy/go-simple-api -- sh
# Внутри:
#   wget -qO- http://localhost:8080/health
#   exit
```

### Эксперимент 4: События namespace

```bash
# Все события -- полезно при отладке проблем
kubectl get events -n wit2-devops --sort-by='.lastTimestamp'
```

---

## Задание 3: Деплой через Helm (20 минут)

### Цель
Развернуть приложение через Helm chart и научиться управлять релизами.

### Подготовка

Сначала удалите ресурсы, созданные манифестами (Helm создаст свои):

```bash
kubectl delete -f k8s/manifests/ingress.yaml
kubectl delete -f k8s/manifests/service.yaml
kubectl delete -f k8s/manifests/deployment.yaml
```

### Шаг 1: Изучаем chart

```bash
# Посмотрим параметры
cat k8s/helm/go-simple-api/values.yaml

# Отрендерим шаблоны (без установки)
helm template go-api k8s/helm/go-simple-api -n wit2-devops
```

Сравните вывод `helm template` с оригинальными манифестами -- они должны быть похожи.

### Шаг 2: Устанавливаем

```bash
helm install go-api k8s/helm/go-simple-api -n wit2-devops

# Проверяем
helm list -n wit2-devops
kubectl get all -n wit2-devops
```

### Шаг 3: Обновляем (upgrade)

```bash
# Увеличиваем реплики до 3
helm upgrade go-api k8s/helm/go-simple-api -n wit2-devops \
  --set replicaCount=3

# Проверяем
kubectl get pods -n wit2-devops
# Должно быть 3 пода

# Смотрим историю
helm history go-api -n wit2-devops
# REVISION  STATUS      DESCRIPTION
# 1         superseded  Install complete
# 2         deployed    Upgrade complete
```

### Шаг 4: Откатываем (rollback)

```bash
# Откат к ревизии 1
helm rollback go-api 1 -n wit2-devops

# Проверяем
kubectl get pods -n wit2-devops
# Обратно 2 пода

helm history go-api -n wit2-devops
# Ревизия 3 -- rollback
```

### Шаг 5: Тестируем

```bash
kubectl port-forward svc/go-api-go-simple-api 8080:80 -n wit2-devops &
curl http://localhost:8080/health
kill %1
```

### Шаг 6: Удаляем

```bash
helm uninstall go-api -n wit2-devops
kubectl get all -n wit2-devops
# Все ресурсы удалены
```

### Проверка знаний

1. В чем преимущество Helm перед обычными манифестами?
   <details>
   <summary>Ответ</summary>
   Параметризация (разные values для dev/prod), версионирование (rollback), управление жизненным циклом (install/upgrade/uninstall как единая операция).
   </details>

2. Что делает `helm template`?
   <details>
   <summary>Ответ</summary>
   Рендерит шаблоны с подставленными значениями из values.yaml, но НЕ применяет результат в кластер. Полезно для проверки и отладки.
   </details>

3. Чем `helm upgrade` отличается от `kubectl apply`?
   <details>
   <summary>Ответ</summary>
   helm upgrade обновляет все ресурсы релиза атомарно, ведет историю ревизий и позволяет откат. kubectl apply обновляет каждый ресурс независимо, без общей истории.
   </details>

---

## Задание 4: Pod -- базовый ресурс (10 минут)

### Цель
Понять разницу между Pod и Deployment на практике.

### Шаги

```bash
# Создаем одиночный Pod
kubectl apply -f k8s/manifests/pod.yaml

# Проверяем
kubectl get pods -n wit2-devops
# Виден отдельный под "go-simple-api" (не управляемый Deployment)

# Удаляем под
kubectl delete -f k8s/manifests/pod.yaml

# Проверяем
kubectl get pods -n wit2-devops
# Под исчез и НЕ пересоздался -- в отличие от Deployment!
```

**Вывод:** Pod -- низкоуровневый ресурс. Без контроллера (Deployment) он не восстанавливается при сбоях. Поэтому на практике всегда используют Deployment.

---

## Итоговое задание: Полный цикл (15 минут)

### Цель
Выполнить полный цикл: deploy -> verify -> update -> rollback -> cleanup

```bash
# 1. Деплой через Helm
helm install go-api k8s/helm/go-simple-api -n wit2-devops

# 2. Проверка
kubectl get all -n wit2-devops
kubectl port-forward svc/go-api-go-simple-api 8080:80 -n wit2-devops &
curl http://localhost:8080/health
kill %1

# 3. Обновление -- 3 реплики
helm upgrade go-api k8s/helm/go-simple-api -n wit2-devops --set replicaCount=3
kubectl get pods -n wit2-devops

# 4. Откат
helm rollback go-api 1 -n wit2-devops
kubectl get pods -n wit2-devops

# 5. Очистка
helm uninstall go-api -n wit2-devops
```

---

## Troubleshooting

### Pod в ImagePullBackOff

```bash
kubectl describe pod <pod-name> -n wit2-devops
# В Events будет: Failed to pull image
# Причина: неверный секрет или имя образа
```

Решение: проверить секрет
```bash
kubectl get secret ghcr-credentials -n wit2-devops -o yaml
```

### Pod в CrashLoopBackOff

```bash
kubectl logs <pod-name> -n wit2-devops
# Посмотрите ошибку запуска приложения
```

### Service не отвечает через port-forward

```bash
# Проверьте endpoints -- есть ли привязанные поды
kubectl get endpoints go-simple-api -n wit2-devops
# Если ENDPOINTS пустые -- labels не совпадают
```

### Helm install ошибка "already exists"

```bash
# Если ресурсы уже созданы манифестами
kubectl delete -f k8s/manifests/deployment.yaml
kubectl delete -f k8s/manifests/service.yaml
kubectl delete -f k8s/manifests/ingress.yaml
# Потом повторить helm install
```
