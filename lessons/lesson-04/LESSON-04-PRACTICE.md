# Практические задания -- Урок 4: ArgoCD & GitOps

Пошаговые инструкции для самостоятельного выполнения.

---

## Задание 1: Установка и доступ к ArgoCD (10 минут)

### Цель
Убедиться что ArgoCD работает и получить доступ к UI.

### Шаги

#### 1. Проверяем что ArgoCD установлен

```bash
kubectl get pods -n argocd
```

Все поды должны быть в `Running`. Если ArgoCD не установлен:

```bash
helm repo add argo https://argoproj.github.io/argo-helm
helm repo update
kubectl create namespace argocd
helm install argocd argo/argo-cd \
  --namespace argocd \
  --set 'server.service.type=NodePort' \
  --set 'configs.params.server\.insecure=true'
```

#### 2. Получаем пароль

```bash
kubectl -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath="{.data.password}" | base64 -d && echo
```

Запишите пароль.

#### 3. Открываем UI

```bash
kubectl port-forward svc/argocd-server -n argocd 8443:80 &
```

Откройте http://localhost:8443 и войдите:
- **Username:** admin
- **Password:** (из шага 2)

#### 4. Авторизуемся в CLI

```bash
argocd login localhost:8443 --insecure --username admin --password <PASSWORD>
argocd app list
# Пока пусто
```

### Проверка

- [ ] ArgoCD UI открывается
- [ ] Могу войти как admin
- [ ] CLI подключен (`argocd app list` работает)

---

## Задание 2: Создание Application через UI (15 минут)

### Цель
Создать ArgoCD Application и задеплоить приложение через GitOps.

### Подготовка

Убедитесь что предыдущий деплой не конфликтует:

```bash
# Удалить helm-релиз если был
helm uninstall go-api -n wit2-devops 2>/dev/null

# Удалить ресурсы от kubectl apply если были
kubectl delete deployment go-simple-api -n wit2-devops 2>/dev/null
kubectl delete svc go-simple-api -n wit2-devops 2>/dev/null
kubectl delete ingress go-simple-api -n wit2-devops 2>/dev/null
```

### Шаги

#### 1. Создаем Application в UI

1. Нажмите **"+ NEW APP"** (или **"CREATE APPLICATION"**)
2. Заполните **General**:
   - Application Name: `go-simple-api`
   - Project Name: `default`
   - Sync Policy: `Manual`
3. Заполните **Source**:
   - Repository URL: `https://github.com/akulovnv/wit2-devops-project`
   - Revision: `main`
   - Path: `k8s/helm/go-simple-api`
4. Заполните **Destination**:
   - Cluster URL: `https://kubernetes.default.svc`
   - Namespace: `wit2-devops`
5. Нажмите **"CREATE"**

#### 2. Наблюдаем статус

После создания приложение будет в статусе:
- Sync: **OutOfSync** (желтый) -- ресурсы из Git не созданы в кластере
- Health: **Missing** -- ресурсов нет

#### 3. Первая синхронизация

1. Кликните на приложение
2. Нажмите **"SYNC"**
3. В окне нажмите **"SYNCHRONIZE"**
4. Наблюдайте за процессом:
   - Ресурсы появляются в дереве
   - Поды переходят в Running
   - Статус меняется на **Synced** + **Healthy**

#### 4. Проверяем в кластере

```bash
kubectl get all -n wit2-devops
# Должны быть: Deployment, ReplicaSet, Pods, Service

kubectl port-forward svc/go-api-go-simple-api 8080:80 -n wit2-devops &
curl http://localhost:8080/health
kill %1
```

### Проверка знаний

1. Что означает статус OutOfSync?
   <details>
   <summary>Ответ</summary>
   Состояние кластера не соответствует тому, что описано в Git-репозитории. ArgoCD обнаружил разницу.
   </details>

2. Почему мы выбрали Manual sync?
   <details>
   <summary>Ответ</summary>
   Для обучения -- чтобы видеть процесс по шагам. В production часто используют manual для контроля, auto -- для dev/staging.
   </details>

---

## Задание 3: Application как YAML-манифест (10 минут)

### Цель
Научиться создавать Application декларативно (через YAML).

### Шаги

#### 1. Удаляем приложение из UI

В ArgoCD UI:
1. Кликните на приложение
2. Нажмите **"DELETE"**
3. Введите имя приложения для подтверждения
4. Отметьте "Cascade" (удалить и ресурсы в кластере)
5. Подтвердите удаление

#### 2. Смотрим YAML-манифест

```bash
cat k8s/argocd/application.yaml
```

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: go-simple-api
  namespace: argocd            # Application ВСЕГДА в namespace argocd
spec:
  project: default
  source:
    repoURL: https://github.com/akulovnv/wit2-devops-project
    targetRevision: main       # Ветка Git
    path: k8s/helm/go-simple-api  # Путь к chart/manifests
    helm:
      valueFiles:
        - values.yaml
  destination:
    server: https://kubernetes.default.svc  # Локальный кластер
    namespace: wit2-devops
  syncPolicy:
    automated:                 # Автоматическая синхронизация
      prune: true              # Удалять ресурсы, которых нет в Git
      selfHeal: true           # Восстанавливать при ручных изменениях
    syncOptions:
      - CreateNamespace=true
```

#### 3. Применяем

```bash
kubectl apply -f k8s/argocd/application.yaml
```

#### 4. Наблюдаем автосинк

Поскольку `syncPolicy.automated` включен, ArgoCD автоматически синхронизирует:

```bash
# Ждем ~30 секунд и проверяем
kubectl get pods -n wit2-devops
# Поды должны появиться автоматически без ручного Sync!
```

В UI статус сразу станет **Synced** + **Healthy**.

### Проверка знаний

1. В каком namespace создается Application?
   <details>
   <summary>Ответ</summary>
   Всегда в namespace argocd, где установлен ArgoCD. Destination namespace -- это где создаются ресурсы приложения.
   </details>

2. Что делает prune: true?
   <details>
   <summary>Ответ</summary>
   Если в Git удалить ресурс (например, Ingress), ArgoCD удалит его из кластера. Без prune -- ресурс останется как "сирота".
   </details>

---

## Задание 4: GitOps в действии (20 минут)

### Цель
Увидеть как изменения в Git автоматически применяются в кластере.

### Эксперимент 1: Изменение числа реплик

```bash
# Текущее состояние
kubectl get pods -n wit2-devops
# 2 пода

# Редактируем values.yaml
# Измените replicaCount: 2 -> replicaCount: 3
```

Вариант через sed:
```bash
cd wit2-devops-project
sed -i '' 's/replicaCount: 2/replicaCount: 3/' k8s/helm/go-simple-api/values.yaml
```

Коммитим и пушим:
```bash
git add k8s/helm/go-simple-api/values.yaml
git commit -m "scale: increase replicas to 3"
git push origin main
```

Ждем синхронизации (1-3 минуты) или нажимаем Refresh в UI:
```bash
kubectl get pods -n wit2-devops
# Теперь 3 пода!
```

### Эксперимент 2: Self-Heal

```bash
# Масштабируем вручную (мимо Git)
kubectl scale deployment -n wit2-devops --all --replicas=5

# Проверяем
kubectl get pods -n wit2-devops
# 5 подов -- но ненадолго!

# Подождите 30-60 секунд...
kubectl get pods -n wit2-devops
# Снова 3 пода -- ArgoCD вернул состояние из Git!
```

**Это и есть self-heal** -- никто не может изменить кластер мимо Git.

### Эксперимент 3: Добавление нового ресурса

Добавим ConfigMap через Git:

```bash
cat > k8s/helm/go-simple-api/templates/configmap.yaml << 'EOF'
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "go-simple-api.fullname" . }}-config
  labels:
    {{- include "go-simple-api.labels" . | nindent 4 }}
data:
  APP_NAME: "Go Simple API"
  ENVIRONMENT: "development"
EOF

git add k8s/helm/go-simple-api/templates/configmap.yaml
git commit -m "feat: add configmap"
git push origin main
```

В ArgoCD UI нажмите Refresh -- увидите новый ConfigMap в дереве ресурсов.

### Эксперимент 4: Diff перед синхронизацией

Если sync policy = manual, можно посмотреть diff:

```bash
argocd app diff go-simple-api
```

Или в UI: нажмите "APP DIFF" -- покажет разницу между Git и кластером.

---

## Задание 5: Откат (Rollback) (10 минут)

### Цель
Научиться откатывать изменения двумя способами.

### Способ 1: Через ArgoCD UI

1. Откройте приложение в UI
2. Нажмите **"HISTORY AND ROLLBACK"**
3. Найдите предыдущую ревизию (2 реплики)
4. Нажмите **"Rollback"**
5. Проверьте:
   ```bash
   kubectl get pods -n wit2-devops
   # 2 пода
   ```

### Способ 2: Через Git (правильный способ)

```bash
# Откатить последний коммит
git revert HEAD --no-edit
git push origin main

# ArgoCD подхватит revert
# Проверяем
kubectl get pods -n wit2-devops
```

### Проверка знаний

1. Почему git revert лучше чем rollback через UI?
   <details>
   <summary>Ответ</summary>
   При git revert остается полная история в Git (кто, когда, почему откатил). Git остается source of truth. При rollback через UI -- self-heal может перезаписать откат обратно, т.к. Git не изменился.
   </details>

2. Что произойдет после rollback через UI если включен selfHeal?
   <details>
   <summary>Ответ</summary>
   ArgoCD заметит что кластер отличается от Git и вернет к состоянию из Git. Т.е. rollback через UI будет отменен! Поэтому правильный способ -- git revert.
   </details>

---

## Задание 6: ArgoCD CLI (10 минут)

### Цель
Познакомиться с основными командами CLI.

### Шаги

```bash
# Список приложений
argocd app list

# Подробная информация
argocd app get go-simple-api

# История синхронизаций
argocd app history go-simple-api

# Принудительный refresh (проверить Git прямо сейчас)
argocd app get go-simple-api --hard-refresh

# Синхронизировать вручную
argocd app sync go-simple-api

# Логи приложения (из ArgoCD)
argocd app logs go-simple-api

# Посмотреть ресурсы приложения
argocd app resources go-simple-api
```

---

## Итоговое задание: Полный GitOps цикл (15 минут)

### Цель
Выполнить полный цикл GitOps: изменение в Git -> автодеплой -> проверка -> откат.

### Шаги

```bash
# 1. Убедитесь что Application работает
argocd app get go-simple-api
kubectl get pods -n wit2-devops

# 2. Внесите изменение в Git
cd wit2-devops-project
# Измените что-то в values.yaml (например LOG_LEVEL: debug)
sed -i '' 's/LOG_LEVEL: "info"/LOG_LEVEL: "debug"/' k8s/helm/go-simple-api/values.yaml
git add . && git commit -m "config: enable debug logging" && git push

# 3. Дождитесь синхронизации (или argocd app sync go-simple-api)
argocd app get go-simple-api --hard-refresh

# 4. Проверьте что изменение применилось
kubectl get pods -n wit2-devops
# Поды перезапустились с новым env

# 5. Проверьте через API
kubectl port-forward svc/go-api-go-simple-api 8080:80 -n wit2-devops &
curl http://localhost:8080/health
kill %1

# 6. Откатите через git revert
git revert HEAD --no-edit && git push

# 7. Дождитесь и проверьте
argocd app get go-simple-api --hard-refresh
kubectl get pods -n wit2-devops
```

---

## Очистка (опционально)

```bash
# Удалить Application (и ресурсы в кластере)
argocd app delete go-simple-api --cascade -y

# Или через kubectl
kubectl delete -f k8s/argocd/application.yaml

# Удалить ArgoCD (если нужно)
helm uninstall argocd -n argocd
kubectl delete namespace argocd
```

---

## Troubleshooting

### ArgoCD не видит репозиторий

```bash
# Проверить подключение
argocd repo list

# Для приватного репо -- добавить credentials
argocd repo add https://github.com/user/repo \
  --username git --password <GITHUB_TOKEN>
```

### Application в статусе Unknown

```bash
# Проверить repo-server
kubectl logs -n argocd -l app.kubernetes.io/name=argocd-repo-server --tail=50

# Часто проблема: неверный path или repoURL
```

### Sync failed: "namespace not found"

Убедитесь что в Application есть:
```yaml
syncOptions:
  - CreateNamespace=true
```

### Self-heal не работает

Проверьте что в syncPolicy указано:
```yaml
syncPolicy:
  automated:
    selfHeal: true
```

И что Application не на паузе:
```bash
argocd app get go-simple-api | grep "Sync Policy"
```
