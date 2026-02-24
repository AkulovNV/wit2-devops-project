# Урок 3: Kubernetes — развёртывание приложения

## Содержание

1. [Теория](#теория)
2. [Установка окружения](#установка-окружения)
3. [Установка зависимостей кластера](#установка-зависимостей-кластера)
4. [Деплой обычными манифестами](#деплой-обычными-манифестами)
5. [Деплой через Helm](#деплой-через-helm)
6. [Проверка работоспособности](#проверка-работоспособности)
7. [Очистка ресурсов](#очистка-ресурсов)

---

## Теория

### Pod

**Pod** — минимальная единица развёртывания в Kubernetes. Это обёртка вокруг одного или нескольких контейнеров, которые разделяют сеть и хранилище. В большинстве случаев Pod содержит ровно один контейнер.

Напрямую Pod создают редко — обычно ими управляют контроллеры (Deployment, StatefulSet и т.д.). Но знать, что такое Pod, важно: все остальные абстракции строятся поверх него.

### Deployment

**Deployment** — контроллер, который управляет набором одинаковых подов (ReplicaSet). Он обеспечивает:

- **Желаемое количество реплик** — если под упал, Deployment создаст новый
- **Rolling update** — обновление без простоя (поды обновляются по одному)
- **Rollback** — откат к предыдущей версии одной командой

### Service

**Service** — стабильная сетевая точка доступа к набору подов. Поды создаются и уничтожаются динамически, а Service предоставляет постоянный IP и DNS-имя внутри кластера.

Основные типы:
- **ClusterIP** (по умолчанию) — доступен только внутри кластера
- **NodePort** — открывает порт на каждом узле кластера
- **LoadBalancer** — создаёт внешний балансировщик (в облаке)

### Ingress

**Ingress** — правила маршрутизации внешнего HTTP/HTTPS-трафика к сервисам внутри кластера. Ingress позволяет:

- Маршрутизировать по доменному имени (`wit2.local → go-simple-api`)
- Маршрутизировать по URL-пути
- Терминировать TLS

Для работы Ingress нужен **Ingress Controller** (например, ingress-nginx).

### Helm

**Helm** — пакетный менеджер для Kubernetes. Helm chart — это набор шаблонизированных манифестов с параметрами (values). Преимущества:

- **Параметризация** — один chart, разные значения для dev/staging/prod
- **Версионирование** — каждый релиз имеет версию, легко откатить
- **Переиспользование** — готовые charts для популярного софта (nginx, PostgreSQL, Prometheus)

---

## Установка окружения

### macOS (рекомендуется Colima)

**Colima** — легковесная альтернатива Docker Desktop с поддержкой Kubernetes.

```bash
# Установка инструментов
brew install colima kubectl helm

# Запуск Colima с Kubernetes
colima kubernetes start

# Проверка
kubectl cluster-info
kubectl get nodes
```

> **Примечание:** если Colima уже запущен без K8s, остановите и перезапустите:
> ```bash
> colima stop
> colima kubernetes start
> ```

### Windows

**Вариант 1: Docker Desktop**

1. Установите [Docker Desktop](https://www.docker.com/products/docker-desktop/)
2. Откройте Settings → Kubernetes → Enable Kubernetes → Apply & Restart
3. Установите kubectl и helm:
   ```powershell
   # Через chocolatey
   choco install kubernetes-cli kubernetes-helm

   # Или через scoop
   scoop install kubectl helm
   ```

**Вариант 2: minikube**

```powershell
choco install minikube
minikube start --driver=docker
```

### Проверка установки

```bash
# Kubectl подключён к кластеру
kubectl cluster-info

# Helm установлен
helm version

# Ноды кластера доступны
kubectl get nodes
```

---

## Установка зависимостей кластера

Для работы Ingress нужен ingress-nginx controller:

```bash
# Добавляем Helm-репозиторий
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

# Устанавливаем ingress-nginx
helm install ingress-nginx ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.service.type=NodePort

# Проверяем, что контроллер запущен
kubectl get pods -n ingress-nginx
```

Дождитесь, пока под `ingress-nginx-controller-*` перейдёт в статус `Running`.

---

## Деплой обычными манифестами

### Структура файлов

```
k8s/manifests/
├── namespace.yaml    # Namespace wit2-devops
├── pod.yaml          # Одиночный Pod (для демонстрации)
├── deployment.yaml   # Deployment с 2 репликами
├── service.yaml      # ClusterIP Service (80 → 8080)
└── ingress.yaml      # Ingress для wit2.local
```

### Шаг 1: Создаём namespace и секрет для доступа к GHCR

Образ `ghcr.io/akulovnv/wit2-devops-project` хранится в приватном реестре. Kubernetes нужен секрет с учётными данными, чтобы скачать образ.

```bash
# Создаём namespace
kubectl apply -f k8s/manifests/namespace.yaml

# Создаём секрет для доступа к GitHub Container Registry
# Замените YOUR_GITHUB_USERNAME и YOUR_GITHUB_TOKEN на свои значения
# Token: GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
# Нужен scope: read:packages
kubectl create secret docker-registry ghcr-credentials \
  --namespace=wit2-devops \
  --docker-server=ghcr.io \
  --docker-username=YOUR_GITHUB_USERNAME \
  --docker-password=YOUR_GITHUB_TOKEN
```

> **Как получить GitHub Token (PAT):**
> 1. Откройте GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
> 2. Generate new token (classic)
> 3. Выберите scope: `read:packages`
> 4. Скопируйте токен и используйте как `--docker-password`

### Шаг 2: Применяем манифесты

```bash
# Применяем все манифесты (кроме pod.yaml — он для демонстрации)
kubectl apply -f k8s/manifests/deployment.yaml
kubectl apply -f k8s/manifests/service.yaml
kubectl apply -f k8s/manifests/ingress.yaml
```

> **Примечание:** файл `pod.yaml` создан для демонстрации того, что Pod — базовый ресурс K8s. На практике поды создаются через Deployment, а не напрямую. Если хотите попробовать:
> ```bash
> kubectl apply -f k8s/manifests/pod.yaml
> kubectl get pod go-simple-api -n wit2-devops
> kubectl delete -f k8s/manifests/pod.yaml
> ```

### Шаг 3: Проверяем

```bash
# Проверяем поды
kubectl get pods -n wit2-devops

# Проверяем сервис
kubectl get svc -n wit2-devops

# Проверяем ingress
kubectl get ingress -n wit2-devops

# Логи одного из подов
kubectl logs -n wit2-devops -l app=go-simple-api --tail=20
```

### Шаг 4: Тестируем доступ

```bash
# Через port-forward (без Ingress)
kubectl port-forward svc/go-simple-api 8080:80 -n wit2-devops &
curl http://localhost:8080/health
curl http://localhost:8080/metrics

# Остановить port-forward
kill %1
```

---

## Деплой через Helm

### Структура chart

```
k8s/helm/go-simple-api/
├── Chart.yaml           # Метаданные chart
├── values.yaml          # Параметры по умолчанию
├── .helmignore          # Исключения при упаковке
└── templates/
    ├── _helpers.tpl     # Хелпер-функции (имена, лейблы)
    ├── deployment.yaml  # Шаблон Deployment
    ├── service.yaml     # Шаблон Service
    ├── ingress.yaml     # Шаблон Ingress (условный)
    └── NOTES.txt        # Сообщение после установки
```

### Шаг 1: Проверяем шаблоны (dry-run)

```bash
# Рендерим шаблоны без установки
helm template go-api k8s/helm/go-simple-api -n wit2-devops
```

### Шаг 2: Устанавливаем chart

```bash
# Создаём namespace (если ещё не создан)
kubectl create namespace wit2-devops --dry-run=client -o yaml | kubectl apply -f -

# Создаём секрет для GHCR (если ещё не создан)
kubectl create secret docker-registry ghcr-credentials \
  --namespace=wit2-devops \
  --docker-server=ghcr.io \
  --docker-username=YOUR_GITHUB_USERNAME \
  --docker-password=YOUR_GITHUB_TOKEN

# Устанавливаем
helm install go-api k8s/helm/go-simple-api -n wit2-devops
```

### Шаг 3: Проверяем установку

```bash
# Статус релиза
helm list -n wit2-devops

# Поды
kubectl get pods -n wit2-devops

# Все ресурсы
kubectl get all -n wit2-devops
```

### Обновление и переопределение параметров

```bash
# Обновить с другими параметрами
helm upgrade go-api k8s/helm/go-simple-api -n wit2-devops \
  --set replicaCount=3 \
  --set image.tag=v1.0.1

# Посмотреть историю
helm history go-api -n wit2-devops

# Откатить к предыдущей версии
helm rollback go-api 1 -n wit2-devops
```

### Удаление Helm-релиза

```bash
helm uninstall go-api -n wit2-devops
```

---

## Проверка работоспособности

### Настройка /etc/hosts

Для доступа через Ingress по домену `wit2.local` добавьте запись в файл hosts:

```bash
# macOS / Linux
echo "127.0.0.1 wit2.local" | sudo tee -a /etc/hosts
```

На Windows (запустите PowerShell от администратора):
```powershell
Add-Content C:\Windows\System32\drivers\etc\hosts "127.0.0.1 wit2.local"
```

### Тестирование через Ingress

```bash
# Health check
curl http://wit2.local/health
# Ожидаемый ответ: {"status":"ok"}

# Метрики
curl http://wit2.local/metrics
```

> **Примечание:** если Ingress не работает (например, Colima с NodePort), используйте port-forward:
> ```bash
> kubectl port-forward svc/go-simple-api 8080:80 -n wit2-devops
> curl http://localhost:8080/health
> ```

### Полезные команды для отладки

```bash
# Описание пода (события, статус)
kubectl describe pod -n wit2-devops -l app=go-simple-api

# Логи пода
kubectl logs -n wit2-devops -l app=go-simple-api --tail=50

# Проверить endpoints сервиса
kubectl get endpoints go-simple-api -n wit2-devops

# Войти в контейнер
kubectl exec -it -n wit2-devops deploy/go-simple-api -- sh
```

---

## Очистка ресурсов

### Удаление манифестов

```bash
# Удалить все ресурсы из манифестов
kubectl delete -f k8s/manifests/ingress.yaml
kubectl delete -f k8s/manifests/service.yaml
kubectl delete -f k8s/manifests/deployment.yaml
kubectl delete -f k8s/manifests/namespace.yaml
```

### Удаление Helm-релиза

```bash
helm uninstall go-api -n wit2-devops
kubectl delete namespace wit2-devops
```

### Удаление ingress-nginx

```bash
helm uninstall ingress-nginx -n ingress-nginx
kubectl delete namespace ingress-nginx
```

### Остановка кластера

```bash
# macOS (Colima)
colima stop

# Windows (minikube)
minikube stop
```

---

## Что дальше

В следующем уроке мы настроим **ArgoCD** для GitOps-деплоя: приложение будет автоматически обновляться в кластере при изменении манифестов в Git-репозитории.
