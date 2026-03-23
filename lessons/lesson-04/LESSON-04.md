# Урок 4: ArgoCD & GitOps -- автоматический деплой из Git

**Цель урока:** Понять подход GitOps, установить и настроить ArgoCD, связать Git-репозиторий с Kubernetes-кластером так, чтобы изменения в Git автоматически применялись в кластере.

**Длительность:** 2 часа

---

## Содержание

1. [Теория: что такое GitOps](#1-теория-что-такое-gitops)
2. [Теория: ArgoCD](#2-теория-argocd)
3. [Установка ArgoCD](#3-установка-argocd)
4. [Первое знакомство с ArgoCD UI](#4-первое-знакомство-с-argocd-ui)
5. [Создание Application](#5-создание-application)
6. [Синхронизация и автосинк](#6-синхронизация-и-автосинк)
7. [Обновление приложения через Git](#7-обновление-приложения-через-git)
8. [Откат (Rollback)](#8-откат-rollback)
9. [ArgoCD CLI](#9-argocd-cli)
10. [Полная картина CI/CD + GitOps](#10-полная-картина-cicd--gitops)

---

## 1. Теория: что такое GitOps

### Проблема: как деплоить в Kubernetes?

На уроке 3 мы деплоили через `kubectl apply` и `helm install`. Это работает, но:

- **Кто последний деплоил и что именно?** -- непонятно
- **Как откатиться?** -- нужно помнить что было до
- **Как гарантировать что в кластере то же самое что в Git?** -- никак
- **Кто имеет доступ к kubectl?** -- проблема безопасности

### Решение: GitOps

**GitOps** -- подход, при котором Git-репозиторий является единственным источником правды (single source of truth) для состояния инфраструктуры и приложений.

Принципы GitOps:

1. **Декларативность** -- желаемое состояние описано в Git (YAML-манифесты, Helm charts)
2. **Версионирование** -- вся история изменений в Git (кто, когда, что изменил)
3. **Автоматическое применение** -- агент (ArgoCD) следит за Git и применяет изменения
4. **Самовосстановление** -- если кто-то вручную изменил кластер, агент вернет к состоянию из Git

### Push vs Pull модель

**Push-модель (традиционная):**
```
CI Pipeline  --push-->  Kubernetes
(kubectl apply)
```
- CI-система имеет credentials к кластеру
- Кластер не знает о "желаемом состоянии"
- Ручные изменения через kubectl никто не отследит

**Pull-модель (GitOps):**
```
Git Repo  <--pull--  ArgoCD  --apply-->  Kubernetes
```
- ArgoCD живет внутри кластера и сам тянет изменения из Git
- CI-системе не нужны credentials к кластеру (только к Git)
- ArgoCD постоянно сравнивает Git и кластер -- любой дрифт виден

### Workflow с GitOps

```
Developer
    |
    v
Git Push (код) --> GitHub Actions (CI)
    |                    |
    |                    v
    |              Build Docker Image --> Push to Registry
    |                    |
    |                    v
    |              Update image tag in Git (k8s/manifests)
    |                    |
    v                    v
Git Repo (манифесты)  <----
    |
    v
ArgoCD (следит за Git)
    |
    v
Kubernetes (применяет изменения)
```

**Что меняется по сравнению с уроком 2-3:**
- CI (GitHub Actions) больше НЕ делает `kubectl apply`
- CI обновляет YAML/values в Git (например, меняет `image.tag`)
- ArgoCD замечает изменение в Git и синхронизирует кластер

---

## 2. Теория: ArgoCD

### Что такое ArgoCD?

**ArgoCD** -- Kubernetes-контроллер, который непрерывно следит за Git-репозиторием и синхронизирует состояние кластера с тем, что описано в Git.

### Ключевые понятия

| Понятие | Описание |
|---------|----------|
| **Application** | Связь между Git-репозиторием (source) и кластером (destination) |
| **Source** | Откуда брать манифесты: Git-репозиторий + путь + ветка |
| **Destination** | Куда деплоить: K8s-кластер + namespace |
| **Sync** | Процесс приведения кластера к состоянию из Git |
| **Sync Status** | Synced (совпадает с Git) / OutOfSync (расходится) |
| **Health Status** | Healthy / Degraded / Progressing / Missing |
| **Refresh** | Проверить Git на наличие изменений |

### Sync Strategies

| Стратегия | Описание | Когда использовать |
|-----------|----------|-------------------|
| **Manual** | Нужно нажать "Sync" вручную | Production (нужен approval) |
| **Auto Sync** | ArgoCD синхронизирует автоматически при изменении в Git | Dev/Staging |
| **Self Heal** | ArgoCD восстанавливает состояние если кто-то изменил кластер вручную | Всегда рекомендуется |
| **Prune** | ArgoCD удалит ресурсы, которых больше нет в Git | С осторожностью |

### Архитектура ArgoCD

```
                    ArgoCD (namespace: argocd)
                    |
      +-------------+-------------+
      |             |             |
API Server     Repo Server   Application Controller
(UI + API)   (клонирует Git)  (следит за состоянием)
      |             |             |
      v             v             v
  Пользователь   Git Repo     K8s API
  (UI/CLI)      (source)     (destination)
```

- **API Server** -- UI и REST API, через него управляем ArgoCD
- **Repo Server** -- клонирует Git-репозитории, рендерит Helm/Kustomize
- **Application Controller** -- следит за Application ресурсами, выполняет sync

---

## 3. Установка ArgoCD

### Через Helm (рекомендуется)

```bash
# Добавляем Helm-репозиторий ArgoCD
helm repo add argo https://argoproj.github.io/argo-helm
helm repo update

# Создаем namespace
kubectl create namespace argocd

# Устанавливаем ArgoCD
helm install argocd argo/argo-cd \
  --namespace argocd \
  --set 'server.service.type=NodePort' \
  --set 'configs.params.server\.insecure=true'
```

Параметры:
- `server.service.type=NodePort` -- доступ к UI через NodePort (на локальном кластере)
- `configs.params.server.insecure=true` -- отключаем HTTPS (для локальной разработки)

### Проверяем установку

```bash
# Все поды должны быть Running
kubectl get pods -n argocd
# NAME                                               READY   STATUS
# argocd-application-controller-0                    1/1     Running
# argocd-applicationset-controller-...               1/1     Running
# argocd-dex-server-...                              1/1     Running
# argocd-notifications-controller-...                1/1     Running
# argocd-redis-...                                   1/1     Running
# argocd-repo-server-...                             1/1     Running
# argocd-server-...                                  1/1     Running
```

### Получаем пароль администратора

```bash
kubectl -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath="{.data.password}" | base64 -d && echo
```

Запишите пароль -- он понадобится для входа.

- **Логин:** `admin`
- **Пароль:** результат команды выше

---

## 4. Первое знакомство с ArgoCD UI

### Открываем UI

```bash
# Port-forward для доступа к UI
kubectl port-forward svc/argocd-server -n argocd 8443:80 &
```

Откройте в браузере: **http://localhost:8443**

Войдите:
- Username: `admin`
- Password: (из предыдущего шага)

### Что видим

После входа видим пустой дашборд -- приложений пока нет. Основные элементы UI:

- **Applications** -- список всех приложений (пока пуст)
- **Settings** -- настройки: репозитории, кластеры, проекты
- **User Info** -- информация о текущем пользователе

---

## 5. Создание Application

Application -- основной ресурс ArgoCD. Он связывает Git (source) с кластером (destination).

### Способ 1: Через UI

1. Нажмите **"+ NEW APP"**
2. Заполните:
   - **Application Name:** `go-simple-api`
   - **Project Name:** `default`
   - **Sync Policy:** Manual (пока без автосинка)
3. Source:
   - **Repository URL:** `https://github.com/akulovnv/wit2-devops-project`
   - **Revision:** `main`
   - **Path:** `k8s/helm/go-simple-api`
4. Destination:
   - **Cluster URL:** `https://kubernetes.default.svc` (локальный кластер)
   - **Namespace:** `wit2-devops`
5. Helm:
   - ArgoCD автоматически определит что это Helm chart
   - Можно переопределить values
6. Нажмите **"CREATE"**

### Способ 2: Через YAML-манифест (рекомендуется)

```yaml
# k8s/argocd/application.yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: go-simple-api
  namespace: argocd
spec:
  project: default

  source:
    repoURL: https://github.com/akulovnv/wit2-devops-project
    targetRevision: main
    path: k8s/helm/go-simple-api
    helm:
      valueFiles:
        - values.yaml

  destination:
    server: https://kubernetes.default.svc
    namespace: wit2-devops

  syncPolicy:
    automated:          # Автоматическая синхронизация
      prune: true       # Удалять ресурсы, которых нет в Git
      selfHeal: true    # Восстанавливать если кто-то изменил вручную
    syncOptions:
      - CreateNamespace=true  # Создать namespace если не существует
```

Применяем:
```bash
kubectl apply -f k8s/argocd/application.yaml
```

### Способ 3: Через ArgoCD CLI

```bash
argocd app create go-simple-api \
  --repo https://github.com/akulovnv/wit2-devops-project \
  --path k8s/helm/go-simple-api \
  --dest-server https://kubernetes.default.svc \
  --dest-namespace wit2-devops \
  --sync-policy automated \
  --auto-prune \
  --self-heal
```

### После создания Application

В UI появится приложение. Его статус:
- **Sync Status: OutOfSync** -- ресурсы еще не созданы в кластере
- **Health: Missing** -- ресурсов нет

---

## 6. Синхронизация и автосинк

### Ручная синхронизация

Если sync policy = Manual:

**Через UI:**
1. Нажмите на приложение
2. Нажмите **"SYNC"**
3. Нажмите **"SYNCHRONIZE"**

**Через CLI:**
```bash
argocd app sync go-simple-api
```

**Через kubectl:**
```bash
# Принудительный refresh
kubectl -n argocd patch application go-simple-api \
  --type merge -p '{"operation":{"sync":{"syncStrategy":{"apply":{}}}}}'
```

### Автоматическая синхронизация

Если в Application указано:
```yaml
syncPolicy:
  automated:
    prune: true
    selfHeal: true
```

-- ArgoCD будет автоматически:
1. Проверять Git каждые 3 минуты (по умолчанию)
2. Если есть разница -- синхронизировать
3. Если кто-то изменил кластер вручную (selfHeal) -- вернуть к состоянию из Git

### Проверяем статус после синхронизации

```bash
# Через CLI
argocd app get go-simple-api

# Или через kubectl
kubectl get application go-simple-api -n argocd

# Проверяем что поды создались
kubectl get pods -n wit2-devops
```

В UI приложение должно стать:
- **Sync Status: Synced** (зеленый)
- **Health: Healthy** (зеленое сердечко)

### Визуализация в UI

ArgoCD UI показывает дерево ресурсов:
```
Application: go-simple-api
  ├── Service/go-simple-api        (Healthy)
  ├── Deployment/go-simple-api     (Healthy)
  │     └── ReplicaSet/go-simple-api-xxx
  │           ├── Pod/go-simple-api-xxx-abc  (Running)
  │           └── Pod/go-simple-api-xxx-def  (Running)
  └── Ingress/go-simple-api        (Healthy)
```

Можно кликнуть на любой ресурс, посмотреть его YAML, логи, события.

---

## 7. Обновление приложения через Git

Это ключевой момент GitOps: изменения в Git автоматически отражаются в кластере.

### Сценарий: обновить число реплик

1. Откройте `k8s/helm/go-simple-api/values.yaml`
2. Измените `replicaCount: 2` на `replicaCount: 3`
3. Закоммитьте и запушьте:
   ```bash
   git add k8s/helm/go-simple-api/values.yaml
   git commit -m "scale: increase replicas to 3"
   git push origin main
   ```
4. Подождите 1-3 минуты (или нажмите Refresh в UI)
5. ArgoCD заметит изменение и синхронизирует:
   ```bash
   kubectl get pods -n wit2-devops
   # Теперь 3 пода
   ```

### Сценарий: обновить версию образа

Это то, что делает CI-pipeline в реальном проекте:

1. CI собирает новый Docker-образ с тегом `v1.1.0`
2. CI обновляет `values.yaml`:
   ```yaml
   image:
     tag: "v1.1.0"  # было "latest"
   ```
3. CI коммитит и пушит изменение
4. ArgoCD подхватывает и деплоит новую версию

### Сценарий: Self-Heal

```bash
# Попробуем изменить кластер вручную
kubectl scale deployment go-api-go-simple-api --replicas=5 -n wit2-devops

# Подождите ~30 секунд
kubectl get pods -n wit2-devops
# ArgoCD вернет к 3 репликам (из Git)
```

Это и есть **self-heal** -- ArgoCD гарантирует что кластер соответствует Git.

---

## 8. Откат (Rollback)

### Через UI

1. Откройте приложение в ArgoCD UI
2. Нажмите **"HISTORY AND ROLLBACK"**
3. Выберите предыдущую ревизию
4. Нажмите **"Rollback"**

### Через CLI

```bash
# Посмотреть историю
argocd app history go-simple-api

# Откатить к конкретной ревизии
argocd app rollback go-simple-api <REVISION_ID>
```

### Через Git (рекомендуемый способ)

В GitOps правильный способ отката -- это `git revert`:

```bash
# Откатить последний коммит
git revert HEAD
git push origin main
# ArgoCD подхватит revert и применит предыдущее состояние
```

Почему через Git лучше:
- Остается история (кто откатил, когда, почему)
- Git остается source of truth
- Rollback через ArgoCD UI -- это "ручное вмешательство", которое selfHeal потом перезапишет

---

## 9. ArgoCD CLI

### Установка

```bash
# macOS
brew install argocd

# Проверка
argocd version --client
```

### Авторизация

```bash
# Логин (используем port-forward)
argocd login localhost:8443 --insecure --username admin --password <PASSWORD>
```

### Основные команды

```bash
# Список приложений
argocd app list

# Подробная информация о приложении
argocd app get go-simple-api

# Синхронизировать
argocd app sync go-simple-api

# Обновить информацию из Git (hard refresh)
argocd app get go-simple-api --hard-refresh

# История синхронизаций
argocd app history go-simple-api

# Логи приложения
argocd app logs go-simple-api

# Удалить приложение (но не ресурсы в кластере)
argocd app delete go-simple-api

# Удалить приложение И ресурсы в кластере
argocd app delete go-simple-api --cascade
```

### Управление репозиториями

```bash
# Добавить приватный репозиторий
argocd repo add https://github.com/user/repo \
  --username git \
  --password <GITHUB_TOKEN>

# Список репозиториев
argocd repo list
```

---

## 10. Полная картина CI/CD + GitOps

### До GitOps (уроки 2-3)

```
Developer --> Git Push --> GitHub Actions --> kubectl apply --> K8s
                              |
                              v
                         Docker Build --> Push to Registry
```

Проблемы:
- CI имеет доступ к кластеру (security risk)
- Нет единого source of truth
- Ручные kubectl-изменения не отслеживаются

### С GitOps (урок 4)

```
Developer --> Git Push --> GitHub Actions (CI) --> Docker Build --> Push to Registry
                              |
                              v
                    Update image tag in Git
                              |
                              v
                         Git Repo (source of truth)
                              |
                              v
                    ArgoCD (pull changes)
                              |
                              v
                    Kubernetes (apply changes)
```

Преимущества:
- CI не имеет доступа к кластеру
- Git = единственный источник правды
- Полный аудит (git log)
- Самовосстановление (self-heal)
- Легкий откат (git revert)

### Пример GitHub Actions + ArgoCD

```yaml
# В CI pipeline после сборки Docker образа:
- name: Update image tag in manifests
  run: |
    cd k8s/helm/go-simple-api
    sed -i "s/tag: .*/tag: \"${{ github.sha }}\"/" values.yaml

- name: Commit and push
  run: |
    git config user.name "github-actions"
    git config user.email "actions@github.com"
    git add k8s/
    git commit -m "deploy: update image to ${{ github.sha }}"
    git push
```

ArgoCD заметит изменение в values.yaml и автоматически задеплоит новую версию.

---

## Чеклист выполнения урока

- [ ] Понимаю концепцию GitOps и отличие от традиционного деплоя
- [ ] Понимаю Pull vs Push модель
- [ ] Установил ArgoCD в кластер
- [ ] Умею открывать ArgoCD UI
- [ ] Создал Application (через UI или YAML)
- [ ] Выполнил ручную синхронизацию
- [ ] Настроил автоматическую синхронизацию
- [ ] Обновил приложение через Git и увидел как ArgoCD подхватил
- [ ] Проверил self-heal (ручное изменение кластера откатилось)
- [ ] Выполнил rollback
- [ ] Знаю основные команды ArgoCD CLI
- [ ] Понимаю полную картину CI/CD + GitOps

---

## Дополнительные ресурсы

- [ArgoCD Documentation](https://argo-cd.readthedocs.io/)
- [GitOps Principles (OpenGitOps)](https://opengitops.dev/)
- [ArgoCD Best Practices](https://argo-cd.readthedocs.io/en/stable/user-guide/best_practices/)
- [ArgoCD Declarative Setup](https://argo-cd.readthedocs.io/en/stable/operator-manual/declarative-setup/)
