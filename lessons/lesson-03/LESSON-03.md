# Урок 3: Kubernetes -- развертывание приложения

**Цель урока:** Понять основные абстракции Kubernetes и научиться развертывать приложение в кластере двумя способами: обычными манифестами и через Helm.

**Длительность:** 2 часа

---

## Содержание

1. [Теория: основные абстракции K8s](#1-теория-основные-абстракции-k8s)
2. [Установка окружения](#2-установка-окружения)
3. [Установка зависимостей кластера](#3-установка-зависимостей-кластера)
4. [Деплой обычными манифестами](#4-деплой-обычными-манифестами)
5. [Деплой через Helm](#5-деплой-через-helm)
6. [Проверка работоспособности](#6-проверка-работоспособности)
7. [Полезные команды для отладки](#7-полезные-команды-для-отладки)
8. [Очистка ресурсов](#8-очистка-ресурсов)

---

## 1. Теория: основные абстракции K8s

### Зачем нужен Kubernetes?

Представьте: у вас есть Docker-образ приложения. Вы умеете его запускать через `docker run`. Но что если:
- Контейнер упал -- кто его перезапустит?
- Нужно 3 копии приложения для нагрузки -- как распределить трафик?
- Нужно обновить версию без даунтайма?
- Нужно откатиться к предыдущей версии?

Kubernetes решает все эти задачи. Он берет на себя управление контейнерами: запускает, следит, перезапускает, масштабирует, обновляет.

### Pod

**Pod** -- минимальная единица развертывания в Kubernetes. Это обертка вокруг одного или нескольких контейнеров, которые разделяют сеть и хранилище.

#### Аналогия: Pod как предмет мебели в комнате

Представьте, что вы заказали диван из IKEA. Приезжает несколько коробок: сам диван, подушки, инструменты для сборки. Все они -- части одного заказа, и нет смысла ставить подушки в одну комнату, а диван в другую.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: диван
  namespace: гостиная
  labels:
    тип: угловой
    цвет: синий
spec:
  containers:
    - name: диван
      image: ikea/sofa:latest
      commands: ["поставить"]
    - name: коробка
      image: ikea/sofa:latest
      commands: ["открыть"]
    - name: инструменты
      image: ikea/tools:latest
      commands: ["собрать"]
```

В реальном Kubernetes:
- **namespace** -- это "комната" (логическое разделение ресурсов)
- **labels** -- это "ярлыки" на коробке (по ним можно фильтровать и искать)
- **containers** -- это составные части одного Pod (работают вместе, видят друг друга по localhost)

В большинстве случаев Pod содержит **один** контейнер. Несколько контейнеров в Pod -- это паттерн sidecar (логирование, прокси и т.д.).

Напрямую Pod создают редко -- обычно ими управляют контроллеры (Deployment, StatefulSet). Но знать, что такое Pod, важно: все остальные абстракции строятся поверх него.

### Deployment

**Deployment** -- контроллер, который управляет набором одинаковых подов через ReplicaSet.

Он обеспечивает:
- **Желаемое количество реплик** -- указали `replicas: 3`, Kubernetes гарантирует 3 работающих пода. Если один упал -- автоматически создаст новый.
- **Rolling update** -- обновление без простоя. Новые поды создаются один за другим, старые удаляются по мере готовности новых.
- **Rollback** -- откат к предыдущей версии одной командой.

```
Deployment (replicas: 3)
   └── ReplicaSet
         ├── Pod-1 (go-simple-api:v1.0)
         ├── Pod-2 (go-simple-api:v1.0)
         └── Pod-3 (go-simple-api:v1.0)
```

При обновлении образа с v1.0 на v1.1:
```
Deployment
   ├── ReplicaSet-old (масштабируется до 0)
   │     ├── Pod-1 (v1.0) -- удаляется
   │     └── Pod-2 (v1.0) -- удаляется
   └── ReplicaSet-new (масштабируется до 3)
         ├── Pod-1 (v1.1) -- создается
         ├── Pod-2 (v1.1) -- создается
         └── Pod-3 (v1.1) -- создается
```

### Service

**Service** -- стабильная сетевая точка доступа к набору подов.

Проблема: поды создаются и уничтожаются динамически, их IP-адреса постоянно меняются. Service дает постоянный IP и DNS-имя внутри кластера.

```
Клиент  -->  Service (go-simple-api:80)
                  ├── Pod-1:8080
                  ├── Pod-2:8080
                  └── Pod-3:8080
```

Service автоматически балансирует трафик между подами, у которых совпадают labels из `selector`.

Основные типы:
- **ClusterIP** (по умолчанию) -- доступен только внутри кластера. Один под обращается к другому по имени сервиса.
- **NodePort** -- открывает порт (30000-32767) на каждом узле кластера. Можно обратиться снаружи по `<NodeIP>:<NodePort>`.
- **LoadBalancer** -- создает внешний балансировщик нагрузки (работает в облаке: AWS, GCP, Azure).

### Ingress

**Ingress** -- правила маршрутизации внешнего HTTP/HTTPS-трафика к сервисам внутри кластера.

```
Интернет --> Ingress Controller --> Ingress Rules --> Service --> Pods
                                     |
                                     ├── wit2.local/       --> go-simple-api
                                     ├── wit2.local/api    --> api-service
                                     └── other.local/      --> other-service
```

Ingress позволяет:
- Маршрутизировать по доменному имени (`wit2.local` -> `go-simple-api`)
- Маршрутизировать по URL-пути (`/api` -> один сервис, `/web` -> другой)
- Терминировать TLS (HTTPS)

Для работы Ingress нужен **Ingress Controller** (например, ingress-nginx). Без него Ingress-ресурсы ничего не делают.

### Namespace

**Namespace** -- логическое разделение ресурсов в кластере. Как папки в файловой системе.

```
Кластер
  ├── namespace: default          # Для экспериментов
  ├── namespace: kube-system      # Системные компоненты K8s
  ├── namespace: wit2-devops      # Наше приложение
  ├── namespace: ingress-nginx    # Ingress controller
  └── namespace: argocd           # ArgoCD (следующий урок)
```

Зачем:
- Изоляция ресурсов (dev/staging/prod могут быть в разных namespace)
- Управление доступом (RBAC по namespace)
- Resource quotas (лимиты на namespace)

### Helm

**Helm** -- пакетный менеджер для Kubernetes. Как `apt` для Linux или `brew` для macOS, только для K8s.

**Проблема:** У нас 5+ YAML-файлов для одного приложения. Если нужно развернуть в dev и prod -- придется копировать и менять значения вручную.

**Решение:** Helm Chart -- шаблонизированные манифесты с параметрами.

```
Helm Chart
├── Chart.yaml       # Метаданные (имя, версия)
├── values.yaml      # Параметры по умолчанию
└── templates/       # Шаблоны манифестов
    ├── deployment.yaml   # {{ .Values.replicaCount }}
    ├── service.yaml
    └── ingress.yaml
```

Один chart, разные values:
```bash
# Dev: 1 реплика, debug логи
helm install app ./chart --set replicaCount=1 --set env.LOG_LEVEL=debug

# Prod: 3 реплики, info логи
helm install app ./chart --set replicaCount=3 --set env.LOG_LEVEL=info
```

Ключевые команды Helm:

| Команда | Что делает |
|---------|-----------|
| `helm install` | Установить chart (создать релиз) |
| `helm upgrade` | Обновить релиз (новые values или chart) |
| `helm rollback` | Откатить к предыдущей ревизии |
| `helm uninstall` | Удалить релиз |
| `helm list` | Показать установленные релизы |
| `helm template` | Отрендерить шаблоны без установки |
| `helm history` | История ревизий релиза |

---

## 2. Установка окружения

### macOS (рекомендуется Colima)

**Colima** -- легковесная альтернатива Docker Desktop с поддержкой Kubernetes.

```bash
# Установка инструментов
brew install colima kubectl helm

# Запуск Colima с Kubernetes
colima start --kubernetes

# Проверка
kubectl cluster-info
kubectl get nodes
```

> **Примечание:** если Colima уже запущен без K8s, остановите и перезапустите:
> ```bash
> colima stop
> colima start --kubernetes
> ```

> **Важно:** убедитесь, что kubectl-контекст указывает на Colima:
> ```bash
> kubectl config current-context
> # Должно быть: colima
>
> # Если нет -- переключите:
> kubectl config use-context colima
> ```

### Windows

**Вариант 1: Docker Desktop**

1. Установите [Docker Desktop](https://www.docker.com/products/docker-desktop/)
2. Откройте Settings -> Kubernetes -> Enable Kubernetes -> Apply & Restart
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
# Kubectl подключен к кластеру
kubectl cluster-info

# Helm установлен
helm version

# Ноды кластера доступны
kubectl get nodes
# NAME     STATUS   ROLES                  AGE   VERSION
# colima   Ready    control-plane,master   1m    v1.33.4+k3s1
```

---

## 3. Установка зависимостей кластера

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

Дождитесь, пока под `ingress-nginx-controller-*` перейдет в статус `Running` (1-2 минуты).

---

## 4. Деплой обычными манифестами

### Структура файлов

```
k8s/manifests/
├── namespace.yaml    # Namespace wit2-devops
├── pod.yaml          # Одиночный Pod (для демонстрации)
├── deployment.yaml   # Deployment с 2 репликами
├── service.yaml      # ClusterIP Service (80 -> 8080)
└── ingress.yaml      # Ingress для wit2.local
```

### Шаг 1: Создаем namespace и секрет для доступа к GHCR

Образ `ghcr.io/akulovnv/wit2-devops-project` хранится в приватном реестре. Kubernetes нужен секрет с учетными данными, чтобы скачать образ.

```bash
# Создаем namespace
kubectl apply -f k8s/manifests/namespace.yaml

# Создаем секрет для доступа к GitHub Container Registry
# Замените YOUR_GITHUB_USERNAME и YOUR_GITHUB_TOKEN на свои значения
kubectl create secret docker-registry ghcr-credentials \
  --namespace=wit2-devops \
  --docker-server=ghcr.io \
  --docker-username=YOUR_GITHUB_USERNAME \
  --docker-password=YOUR_GITHUB_TOKEN
```

> **Как получить GitHub Token (PAT):**
> 1. GitHub -> Settings -> Developer settings -> Personal access tokens -> Tokens (classic)
> 2. Generate new token (classic)
> 3. Выберите scope: `read:packages`
> 4. Скопируйте токен и используйте как `--docker-password`

### Шаг 2: Разбираем каждый манифест

#### namespace.yaml -- создание "комнаты" для наших ресурсов

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: wit2-devops
```

#### deployment.yaml -- описание приложения

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: go-simple-api
  namespace: wit2-devops
spec:
  replicas: 2                    # Хотим 2 копии приложения
  selector:
    matchLabels:
      app: go-simple-api         # Deployment управляет подами с этим label
  template:
    metadata:
      labels:
        app: go-simple-api       # Label пода -- должен совпадать с selector
    spec:
      imagePullSecrets:
        - name: ghcr-credentials # Секрет для скачивания образа из GHCR
      containers:
        - name: go-simple-api
          image: ghcr.io/akulovnv/wit2-devops-project:latest
          ports:
            - containerPort: 8080
          env:
            - name: PORT
              value: "8080"
          livenessProbe:          # Kubernetes перезапустит под, если проверка провалится
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
          readinessProbe:         # Kubernetes не будет отправлять трафик, пока под не готов
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 3
            periodSeconds: 5
          resources:
            requests:             # Минимальные гарантированные ресурсы
              cpu: 50m
              memory: 64Mi
            limits:               # Максимальные допустимые ресурсы
              cpu: 200m
              memory: 128Mi
```

Ключевые поля:
- **replicas** -- сколько подов создать
- **selector.matchLabels** -- как Deployment находит "свои" поды
- **livenessProbe** -- жив ли контейнер? Если нет -- перезапуск
- **readinessProbe** -- готов ли контейнер принимать трафик? Если нет -- Service не направляет на него запросы
- **resources.requests** -- сколько ресурсов "забронировать" (влияет на планирование)
- **resources.limits** -- максимум (при превышении memory -- OOMKill, при превышении CPU -- throttling)

#### service.yaml -- стабильная точка входа

```yaml
apiVersion: v1
kind: Service
metadata:
  name: go-simple-api
  namespace: wit2-devops
spec:
  type: ClusterIP
  selector:
    app: go-simple-api   # Направлять трафик на поды с этим label
  ports:
    - port: 80           # Порт Service (на него обращаются клиенты)
      targetPort: 8080   # Порт контейнера (куда перенаправляется)
```

#### ingress.yaml -- внешний доступ

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: go-simple-api
  namespace: wit2-devops
spec:
  ingressClassName: nginx
  rules:
    - host: wit2.local          # Доменное имя
      http:
        paths:
          - path: /             # Все запросы на wit2.local/
            pathType: Prefix
            backend:
              service:
                name: go-simple-api  # Направлять на Service
                port:
                  number: 80
```

### Шаг 3: Применяем манифесты

```bash
# Применяем все манифесты (кроме pod.yaml -- он для демонстрации)
kubectl apply -f k8s/manifests/deployment.yaml
kubectl apply -f k8s/manifests/service.yaml
kubectl apply -f k8s/manifests/ingress.yaml
```

> **Примечание:** файл `pod.yaml` создан для демонстрации. На практике поды создаются через Deployment, а не напрямую. Если хотите попробовать:
> ```bash
> kubectl apply -f k8s/manifests/pod.yaml
> kubectl get pod go-simple-api -n wit2-devops
> kubectl delete -f k8s/manifests/pod.yaml
> ```

### Шаг 4: Проверяем

```bash
# Проверяем поды
kubectl get pods -n wit2-devops
# NAME                             READY   STATUS    RESTARTS   AGE
# go-simple-api-678fb688c6-tdg42   1/1     Running   0          30s
# go-simple-api-678fb688c6-xhgbl   1/1     Running   0          30s

# Проверяем сервис
kubectl get svc -n wit2-devops

# Проверяем ingress
kubectl get ingress -n wit2-devops

# Логи подов
kubectl logs -n wit2-devops -l app=go-simple-api --tail=20
```

### Шаг 5: Тестируем доступ

```bash
# Через port-forward (самый надежный способ на локальном кластере)
kubectl port-forward svc/go-simple-api 8080:80 -n wit2-devops &
curl http://localhost:8080/health
curl http://localhost:8080/metrics

# Остановить port-forward
kill %1
```

---

## 5. Деплой через Helm

### Зачем Helm, если манифесты работают?

С 5 файлами манифестов можно справиться. Но когда:
- Нужно развернуть в dev с 1 репликой и debug-логами, а в prod с 3 репликами и info-логами
- Нужно обновить версию образа в 3 файлах одновременно
- Нужно откатить ВСЕ ресурсы к предыдущей версии одной командой

-- Helm становится незаменим.

### Структура chart

```
k8s/helm/go-simple-api/
├── Chart.yaml           # Метаданные chart (имя, версия)
├── values.yaml          # Параметры по умолчанию
└── templates/
    ├── _helpers.tpl     # Хелпер-функции (имена, лейблы)
    ├── deployment.yaml  # Шаблон Deployment
    ├── service.yaml     # Шаблон Service
    ├── ingress.yaml     # Шаблон Ingress (условный)
    └── NOTES.txt        # Сообщение после установки
```

### Как работают шаблоны

В `values.yaml` определяются параметры:
```yaml
replicaCount: 2
image:
  repository: ghcr.io/akulovnv/wit2-devops-project
  tag: "latest"
```

В `templates/deployment.yaml` используются через `{{ .Values.xxx }}`:
```yaml
spec:
  replicas: {{ .Values.replicaCount }}
  containers:
    - image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
```

При установке Helm подставляет значения и применяет результат.

### Шаг 1: Проверяем шаблоны (dry-run)

```bash
# Рендерим шаблоны без установки -- можно посмотреть итоговый YAML
helm template go-api k8s/helm/go-simple-api -n wit2-devops
```

### Шаг 2: Устанавливаем chart

```bash
# Создаем namespace (если еще не создан)
kubectl create namespace wit2-devops --dry-run=client -o yaml | kubectl apply -f -

# Создаем секрет для GHCR (если еще не создан -- команда вернет ошибку, это ОК)
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
# Обновить количество реплик и тег образа
helm upgrade go-api k8s/helm/go-simple-api -n wit2-devops \
  --set replicaCount=3 \
  --set image.tag=v1.0.1

# Посмотреть историю ревизий
helm history go-api -n wit2-devops

# Откатить к предыдущей версии
helm rollback go-api 1 -n wit2-devops
```

### Удаление Helm-релиза

```bash
helm uninstall go-api -n wit2-devops
```

---

## 6. Проверка работоспособности

### Настройка /etc/hosts

Для доступа через Ingress по домену `wit2.local`:

```bash
# macOS / Linux
echo "127.0.0.1 wit2.local" | sudo tee -a /etc/hosts
```

Windows (PowerShell от администратора):
```powershell
Add-Content C:\Windows\System32\drivers\etc\hosts "127.0.0.1 wit2.local"
```

### Тестирование

```bash
# Через port-forward (надежный способ)
kubectl port-forward svc/go-simple-api 8080:80 -n wit2-devops &

# Health check
curl http://localhost:8080/health
# {"status":"ok","timestamp":"...","version":"1.0.0"}

# Метрики Prometheus
curl http://localhost:8080/metrics

# Остановить port-forward
kill %1
```

---

## 7. Полезные команды для отладки

```bash
# Описание пода (события, статус, причины ошибок)
kubectl describe pod -n wit2-devops -l app=go-simple-api

# Логи пода
kubectl logs -n wit2-devops -l app=go-simple-api --tail=50

# Логи предыдущего контейнера (если под перезапускался)
kubectl logs -n wit2-devops -l app=go-simple-api --previous

# Проверить endpoints сервиса (к каким подам подключен)
kubectl get endpoints go-simple-api -n wit2-devops

# Войти в контейнер (для отладки)
kubectl exec -it -n wit2-devops deploy/go-simple-api -- sh

# Посмотреть события в namespace (полезно при ошибках)
kubectl get events -n wit2-devops --sort-by='.lastTimestamp'

# Проверить ресурсы подов
kubectl top pods -n wit2-devops
```

### Типичные проблемы и решения

| Проблема | Как диагностировать | Решение |
|----------|---------------------|---------|
| Pod в `ImagePullBackOff` | `kubectl describe pod ...` | Проверить секрет `ghcr-credentials`, имя образа |
| Pod в `CrashLoopBackOff` | `kubectl logs ...` | Ошибка в приложении, проверить логи |
| Service не отвечает | `kubectl get endpoints ...` | Проверить совпадение labels в selector |
| Ingress 404 | `kubectl describe ingress ...` | Проверить ingressClassName, backend |

---

## 8. Очистка ресурсов

### Удаление манифестов

```bash
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

## Чеклист выполнения урока

- [ ] Понимаю что такое Pod, Deployment, Service, Ingress
- [ ] Могу объяснить разницу между ClusterIP, NodePort, LoadBalancer
- [ ] Умею применять K8s манифесты через kubectl apply
- [ ] Умею диагностировать проблемы (describe, logs, get events)
- [ ] Понимаю зачем нужен Helm и как он работает
- [ ] Умею устанавливать, обновлять и откатывать Helm-релизы
- [ ] Развернул приложение в локальном кластере

---

## Что дальше

В следующем уроке мы настроим **ArgoCD** для GitOps-деплоя: приложение будет автоматически обновляться в кластере при изменении манифестов в Git-репозитории.
