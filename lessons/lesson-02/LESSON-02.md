# УРОК 2: CI/CD Pipeline - От кода до Docker образа

**Цель урока:** Научиться проектировать и реализовывать эффективный CI/CD пайплайн с использованием GitHub Actions

**Длительность:** 2 часа

---

## 📋 СОДЕРЖАНИЕ

1. [Проектирование пайплайна](#1-проектирование-пайплайна)
2. [Matrix Strategies](#2-matrix-strategies)
3. [Caching зависимостей](#3-caching-зависимостей)
4. [Docker Build & Push](#4-docker-build--push)
5. [Secrets и Permissions](#5-secrets-и-permissions)
6. [Переиспользование кода](#6-переиспользование-кода)
7. [Стратегии версионирования](#7-стратегии-версионирования)
8. [SBOM (Software Bill of Materials)](#sbom-software-bill-of-materials)
9. [Практические задания](#8-практические-задания)

---

## 1. ПРОЕКТИРОВАНИЕ ПАЙПЛАЙНА

### Теория (5 минут)

**Идеализированный CI/CD pipeline** состоит из следующих стадий:

```
CODE → BUILD → TEST → PACKAGE → DEPLOY
  ↓       ↓      ↓       ↓         ↓
  │       │      │       │         │
  └─→ Lint & Format     │         │
          └─→ Unit Tests │         │
                 └─→ Docker Image  │
                        └─→ Kubernetes (ArgoCD)
```

### Стадии пайплайна:

#### 1. **Validation (Проверка кода)**
- Форматирование (`gofmt`, `prettier`)
- Линтинг (`golangci-lint`, `eslint`)
- Статический анализ (`go vet`)

#### 2. **Testing (Тестирование)**
- Unit тесты
- Integration тесты
- Coverage отчеты

#### 3. **Build (Сборка)**
- Компиляция бинарника
- Docker образ (multi-stage build)
- Оптимизация размера образа

#### 4. **Security (Безопасность)**
- Сканирование уязвимостей (Trivy, Snyk)
- SAST (Static Application Security Testing)
- Проверка зависимостей
- Генерация SBOM (Software Bill of Materials)

#### 5. **Package (Упаковка)**
- Тегирование образов
- Push в Container Registry
- Создание артефактов

#### 6. **Deploy (Развертывание)**
- ArgoCD синхронизация
- GitOps подход
- Rollback стратегия

### Практика: Базовая структура workflow

```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  # 1. Валидация кода
  validate:
    name: Code Validation
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Check formatting
        run: gofmt -l .

  # 2. Тестирование
  test:
    name: Run Tests
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run tests
        run: go test -v

  # 3. Сборка Docker образа
  build:
    name: Build & Push Docker
    needs: [validate, test]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build Docker image
        run: docker build -t myapp:latest .
```

**Ключевые концепции:**
- `on:` - триггеры (когда запускается пайплайн)
- `jobs:` - независимые задачи
- `needs:` - зависимости между jobs
- `steps:` - последовательные шаги внутри job

---

## 2. MATRIX STRATEGIES

### Теория (10 минут)

**Matrix Strategy** позволяет запускать одну и ту же job с разными параметрами параллельно.

### Зачем нужны Matrix?

1. **Тестирование на разных версиях**
   - Go: 1.21, 1.22, 1.23
   - Node.js: 18, 20, 22
   - Python: 3.10, 3.11, 3.12

2. **Тестирование на разных OS**
   - ubuntu-latest, windows-latest, macos-latest

3. **Параллельное выполнение задач**
   - Разные типы тестов
   - Разные окружения

### Базовый пример:

```yaml
jobs:
  test:
    strategy:
      matrix:
        go-version: [1.21, 1.22, 1.23]
        os: [ubuntu-latest, macos-latest]

    runs-on: ${{ matrix.os }}

    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}

      - name: Run tests
        run: go test -v
```

Это создаст **6 параллельных jobs** (3 версии Go × 2 OS)

### Продвинутые возможности:

#### 1. Include (добавить специфичные комбинации)

```yaml
strategy:
  matrix:
    go-version: [1.21, 1.22]
    os: [ubuntu-latest]
    include:
      # Добавляем специальную комбинацию
      - go-version: 1.23
        os: ubuntu-latest
        experimental: true
```

#### 2. Exclude (исключить комбинации)

```yaml
strategy:
  matrix:
    go-version: [1.21, 1.22, 1.23]
    os: [ubuntu-latest, macos-latest, windows-latest]
    exclude:
      # Не тестируем Go 1.21 на Windows
      - go-version: 1.21
        os: windows-latest
```

#### 3. Fail-fast (останавливать все при первой ошибке)

```yaml
strategy:
  fail-fast: false  # Продолжаем даже если одна комбинация упала
  matrix:
    go-version: [1.21, 1.22, 1.23]
```

### Практический пример для нашего проекта:

```yaml
jobs:
  test:
    name: Test on Go ${{ matrix.go-version }}

    strategy:
      fail-fast: false
      matrix:
        go-version: ['1.21', '1.22', '1.23']

    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go ${{ matrix.go-version }}
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}

      - name: Cache Go modules
        uses: actions/cache@v4
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ matrix.go-version }}-${{ hashFiles('**/go.sum') }}

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out

      - name: Check coverage
        run: go tool cover -func=coverage.out
```

---

## 3. CACHING ЗАВИСИМОСТЕЙ

### Теория (10 минут)

**Зачем нужен кэш?**

- Ускорение builds (вместо 2-3 минут → 30 секунд)
- Экономия трафика
- Стабильность (меньше сетевых запросов)

### Что кэшировать?

#### Go проекты:
```yaml
- name: Cache Go modules
  uses: actions/cache@v4
  with:
    path: ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    restore-keys: |
      ${{ runner.os }}-go-
```

#### Node.js проекты:
```yaml
- name: Cache npm dependencies
  uses: actions/cache@v4
  with:
    path: ~/.npm
    key: ${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}
    restore-keys: |
      ${{ runner.os }}-node-
```

#### Docker layers:
```yaml
- name: Set up Docker Buildx
  uses: docker/setup-buildx-action@v3

- name: Build and push
  uses: docker/build-push-action@v5
  with:
    context: .
    push: true
    tags: myapp:latest
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

### Структура кэша:

```yaml
with:
  path: ~/go/pkg/mod              # Что кэшировать
  key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}  # Уникальный ключ
  restore-keys: |                 # Fallback ключи
    ${{ runner.os }}-go-
```

**Как работает `key`:**
- `${{ runner.os }}` → `Linux`
- `${{ hashFiles('**/go.sum') }}` → хэш файла зависимостей
- Полный ключ: `Linux-go-a1b2c3d4...`

**Если go.sum изменился:**
- Старый кэш не подходит (ключ изменился)
- Создается новый кэш
- Старый кэш через 7 дней удаляется

### Сравнение производительности:

| Действие | Без кэша | С кэшем |
|----------|----------|---------|
| go mod download | 45s | 2s |
| npm install | 90s | 5s |
| Docker build | 180s | 30s |

---

## 4. DOCKER BUILD & PUSH

### Теория (10 минут)

### Multi-stage Build

**Проблема:** Go приложение размером 700MB+

**Решение:** Multi-stage build → 15MB

```dockerfile
# Stage 1: Build
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Копируем только файлы зависимостей (для кэширования)
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходники и собираем
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags="-w -s -X main.Version=${VERSION}" \
    -o app main.go

# Stage 2: Runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем только бинарник из builder stage
COPY --from=builder /build/app .

EXPOSE 8080

ENTRYPOINT ["/app/app"]
```

### Оптимизации:

1. **Порядок слоев** (от редко меняющихся к часто меняющимся):
   ```dockerfile
   COPY go.mod go.sum ./     # Меняются редко
   RUN go mod download       # Кэшируется
   COPY . .                  # Меняется часто
   ```

2. **Статическая компиляция:**
   ```bash
   CGO_ENABLED=0 GOOS=linux go build
   ```

3. **Удаление debug информации:**
   ```bash
   -ldflags="-w -s"  # -w: удаляет DWARF, -s: удаляет таблицу символов
   ```

### GitHub Actions: Build & Push

```yaml
jobs:
  docker:
    name: Build & Push Docker Image
    runs-on: ubuntu-latest

    permissions:
      contents: read
      packages: write

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to GitHub Container Registry
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ghcr.io/${{ github.repository }}
          tags: |
            type=sha,prefix={{branch}}-
            type=ref,event=branch
            type=semver,pattern={{version}}

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          build-args: |
            VERSION=${{ github.sha }}
```

---

## 5. SECRETS И PERMISSIONS

### Теория (10 минут)

### Типы секретов:

#### 1. GitHub Secrets (Settings → Secrets and variables → Actions)

```yaml
steps:
  - name: Login to Docker Hub
    uses: docker/login-action@v3
    with:
      username: ${{ secrets.DOCKERHUB_USERNAME }}
      password: ${{ secrets.DOCKERHUB_TOKEN }}
```

#### 2. GITHUB_TOKEN (автоматически доступен)

```yaml
- name: Create Release
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  run: gh release create v1.0.0
```

#### 3. Environment Secrets (для разных окружений)

```yaml
jobs:
  deploy:
    environment: production  # Секреты из окружения "production"
    steps:
      - name: Deploy
        env:
          API_KEY: ${{ secrets.PROD_API_KEY }}
```

### Permissions

**Минимальные права (Principle of Least Privilege):**

```yaml
jobs:
  build:
    permissions:
      contents: read        # Чтение кода
      packages: write       # Запись в GitHub Packages
      security-events: write  # Загрузка SARIF отчетов

    steps:
      - uses: actions/checkout@v4
      - name: Build Docker
        run: docker build .
```

**Типы permissions:**
- `actions: read|write` - доступ к Actions
- `contents: read|write` - доступ к репозиторию
- `packages: read|write` - GitHub Packages
- `pull-requests: read|write` - PR комментарии
- `security-events: write` - загрузка security отчетов

### Best Practices:

1. **НЕ логировать секреты**
   ```yaml
   # ❌ ПЛОХО
   - name: Debug
     run: echo "Token: ${{ secrets.API_TOKEN }}"

   # ✅ ХОРОШО
   - name: Debug
     run: echo "Token is set: ${{ secrets.API_TOKEN != '' }}"
   ```

2. **Использовать environment protection rules**
   - Требовать approval для production
   - Ограничивать deployment по веткам

3. **Ротация секретов**
   - Регулярно обновлять токены
   - Использовать short-lived токены

---

## 6. ПЕРЕИСПОЛЬЗОВАНИЕ КОДА

### Теория (10 минут)

### Способы переиспользования:

#### 1. Composite Actions

**Зачем нужна директория `.github/actions/`?**

Директория `.github/actions/` используется для хранения **переиспользуемых действий (actions)** внутри вашего репозитория. Это позволяет:

1. **Устранить дублирование кода** - вместо повторения одних и тех же шагов в разных workflows, вы создаете одно action
2. **Упростить поддержку** - изменения в одном месте автоматически применяются везде
3. **Улучшить читаемость** - workflows становятся компактнее и понятнее
4. **Версионирование** - можно создавать разные версии actions для разных проектов

**Структура:**
```
.github/
├── actions/              # Переиспользуемые actions
│   ├── setup-go/        # Action для настройки Go
│   │   └── action.yml   # Описание action
│   ├── docker-build/    # Action для сборки Docker
│   │   └── action.yml
│   └── run-tests/       # Action для запуска тестов
│       └── action.yml
└── workflows/           # CI/CD pipelines
    ├── ci-cd.yml
    └── release.yml
```

**Пример: часто используемая последовательность шагов**

Без composite action (повторяется в каждом workflow):
```yaml
# workflow-1.yml
- uses: actions/setup-go@v5
  with:
    go-version: '1.23'
- uses: actions/cache@v4
  with:
    path: ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}

# workflow-2.yml (то же самое повторяется)
- uses: actions/setup-go@v5
  with:
    go-version: '1.23'
- uses: actions/cache@v4
  with:
    path: ~/go/pkg/mod
    key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
```

С composite action (используется один раз):
```yaml
# workflow-1.yml
- uses: ./.github/actions/setup-go

# workflow-2.yml
- uses: ./.github/actions/setup-go
```

**Файл: `.github/actions/setup-go/action.yml`**

```yaml
name: 'Setup Go with cache'
description: 'Setup Go and cache modules'

inputs:
  go-version:
    description: 'Go version'
    required: true
    default: '1.23'

runs:
  using: 'composite'
  steps:
    - name: Setup Go
      uses: actions/setup-go@v5
      with:
        go-version: ${{ inputs.go-version }}

    - name: Cache Go modules
      uses: actions/cache@v4
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ inputs.go-version }}-${{ hashFiles('**/go.sum') }}
```

**Использование:**
```yaml
jobs:
  test:
    steps:
      - uses: actions/checkout@v4
      - uses: ./.github/actions/setup-go
        with:
          go-version: '1.23'
```

#### 2. Reusable Workflows

**Файл: `.github/workflows/test-template.yml`**

```yaml
name: Reusable Test Workflow

on:
  workflow_call:
    inputs:
      go-version:
        required: true
        type: string
    secrets:
      token:
        required: true

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ inputs.go-version }}
      - run: go test -v
```

**Использование:**
```yaml
jobs:
  call-test:
    uses: ./.github/workflows/test-template.yml
    with:
      go-version: '1.23'
    secrets:
      token: ${{ secrets.GITHUB_TOKEN }}
```

#### 3. Matrix + Reusable Workflows

```yaml
jobs:
  test:
    strategy:
      matrix:
        go-version: ['1.21', '1.22', '1.23']
    uses: ./.github/workflows/test-template.yml
    with:
      go-version: ${{ matrix.go-version }}
```

---

## 7. СТРАТЕГИИ ВЕРСИОНИРОВАНИЯ

### Теория (10 минут)

### SemVer (Semantic Versioning)

```
MAJOR.MINOR.PATCH

1.2.3
│ │ └─ Patch: исправления багов (обратно совместимы)
│ └─── Minor: новая функциональность (обратно совместима)
└───── Major: breaking changes (НЕ обратно совместимы)
```

### Стратегии тегирования Docker образов:

#### 1. **Git SHA (для каждого коммита)**
```yaml
tags: |
  ghcr.io/myorg/myapp:${{ github.sha }}
  ghcr.io/myorg/myapp:main-abc1234
```

#### 2. **Semantic Versioning (для релизов)**
```yaml
tags: |
  ghcr.io/myorg/myapp:1.2.3
  ghcr.io/myorg/myapp:1.2
  ghcr.io/myorg/myapp:1
  ghcr.io/myorg/myapp:latest
```

#### 3. **Branch-based (для feature веток)**
```yaml
tags: |
  ghcr.io/myorg/myapp:develop
  ghcr.io/myorg/myapp:feature-auth
```

### Автоматическое версионирование:

```yaml
- name: Generate version
  id: version
  run: |
    VERSION=$(git describe --tags --always --dirty)
    echo "version=${VERSION}" >> $GITHUB_OUTPUT
    echo "short_sha=${GITHUB_SHA::8}" >> $GITHUB_OUTPUT

- name: Build with version
  run: |
    docker build \
      --build-arg VERSION=${{ steps.version.outputs.version }} \
      --tag myapp:${{ steps.version.outputs.version }} \
      .
```

### docker/metadata-action (рекомендуется)

```yaml
- name: Docker metadata
  id: meta
  uses: docker/metadata-action@v5
  with:
    images: ghcr.io/${{ github.repository }}
    tags: |
      # Для веток: main -> main-abc1234
      type=sha,prefix={{branch}}-

      # Для тегов: v1.2.3 -> 1.2.3, 1.2, 1, latest
      type=semver,pattern={{version}}
      type=semver,pattern={{major}}.{{minor}}
      type=semver,pattern={{major}}

      # Для PR: pr-123
      type=ref,event=pr
```

### SBOM (Software Bill of Materials)

**SBOM** — это "ведомость материалов" для программного обеспечения, полный список всех компонентов, библиотек и зависимостей, которые используются в приложении.

#### Зачем нужен SBOM?

1. **Безопасность**
   - Быстрое выявление уязвимых зависимостей
   - Отслеживание использования библиотек с известными CVE
   - Пример: Log4Shell (CVE-2021-44228) — с SBOM можно за минуты найти все затронутые приложения

2. **Комплаенс и аудит**
   - Соответствие требованиям (SOC2, ISO 27001)
   - Проверка лицензий dependencies
   - Документирование для регуляторов

3. **Supply Chain Security**
   - Понимание цепочки поставок ПО
   - Обнаружение неожиданных зависимостей
   - Защита от атак на supply chain

4. **Управление зависимостями**
   - Инвентаризация всех компонентов
   - Планирование обновлений
   - Анализ устаревших библиотек

#### Форматы SBOM:

**1. SPDX (Software Package Data Exchange)**
```json
{
  "spdxVersion": "SPDX-2.3",
  "name": "go-simple-api",
  "packages": [
    {
      "name": "github.com/gorilla/mux",
      "versionInfo": "v1.8.0",
      "licenseConcluded": "BSD-3-Clause"
    }
  ]
}
```

**2. CycloneDX**
```xml
<bom>
  <components>
    <component type="library">
      <name>github.com/gorilla/mux</name>
      <version>1.8.0</version>
      <licenses>
        <license>
          <id>BSD-3-Clause</id>
        </license>
      </licenses>
    </component>
  </components>
</bom>
```

#### Генерация SBOM в CI/CD:

**Вариант 1: Syft (от Anchore)**
```yaml
- name: Generate SBOM
  uses: anchore/sbom-action@v0
  with:
    image: ghcr.io/${{ github.repository }}:latest
    format: spdx-json
    output-file: sbom.spdx.json
```

**Вариант 2: Trivy (универсальный)**
```yaml
- name: Generate SBOM with Trivy
  uses: aquasecurity/trivy-action@master
  with:
    scan-type: 'image'
    image-ref: ghcr.io/${{ github.repository }}:latest
    format: 'cyclonedx'
    output: 'sbom.cdx.json'
```

**Вариант 3: Docker Scout (встроенный в Docker)**
```bash
docker scout sbom myapp:latest --format spdx > sbom.spdx.json
```

#### Полный пример с SBOM в workflow:

```yaml
jobs:
  docker:
    name: Build & Push with SBOM
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to GHCR
        uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Build and push
        id: build
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ghcr.io/${{ github.repository }}:latest

      - name: Generate SBOM
        uses: anchore/sbom-action@v0
        with:
          image: ghcr.io/${{ github.repository }}@${{ steps.build.outputs.digest }}
          format: spdx-json
          output-file: sbom.spdx.json

      - name: Upload SBOM as artifact
        uses: actions/upload-artifact@v4
        with:
          name: sbom
          path: sbom.spdx.json
          retention-days: 90

      - name: Scan SBOM for vulnerabilities
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'sbom'
          scan-ref: 'sbom.spdx.json'
          format: 'table'
```

#### Best Practices:

1. **Генерируйте SBOM для каждого релиза**
   ```yaml
   on:
     release:
       types: [published]
   ```

2. **Храните SBOM вместе с образом**
   - Как артефакт в GitHub Actions
   - В Container Registry (attestation)
   - В S3/blob storage для долгосрочного хранения

3. **Автоматизируйте проверку SBOM**
   ```yaml
   - name: Check SBOM for critical vulnerabilities
     run: |
       trivy sbom sbom.spdx.json --severity CRITICAL --exit-code 1
   ```

4. **Подписывайте SBOM (Sigstore/Cosign)**
   ```yaml
   - name: Sign SBOM
     run: |
       cosign sign-blob sbom.spdx.json \
         --bundle sbom.bundle \
         --yes
   ```

#### Инструменты для работы с SBOM:

| Инструмент | Генерация | Анализ | Форматы |
|------------|-----------|---------|---------|
| **Syft** | ✅ | ❌ | SPDX, CycloneDX, JSON |
| **Trivy** | ✅ | ✅ | SPDX, CycloneDX |
| **Grype** | ❌ | ✅ | Читает SBOM и ищет CVE |
| **Docker Scout** | ✅ | ✅ | SPDX |
| **OSV-Scanner** | ❌ | ✅ | Google's OSV database |

#### Пример использования SBOM для поиска уязвимостей:

```bash
# 1. Генерация SBOM
syft ghcr.io/myorg/myapp:latest -o spdx-json > sbom.json

# 2. Анализ SBOM на уязвимости
grype sbom:sbom.json

# 3. Проверка лицензий
syft sbom.json -o table --select-catalogers all
```

#### Реальный сценарий использования:

**Проблема:** Обнаружена критическая уязвимость в библиотеке `golang.org/x/net`

**Решение с SBOM:**
```bash
# 1. Найти все образы, использующие уязвимую библиотеку
for sbom in artifacts/*.spdx.json; do
  if grep -q "golang.org/x/net" "$sbom"; then
    echo "Affected: $sbom"
  fi
done

# 2. Определить версии
jq '.packages[] | select(.name == "golang.org/x/net") | .versionInfo' sbom.json

# 3. Приоритизировать обновления
```

**Результат:** Вместо нескольких дней на поиск — несколько минут на идентификацию всех затронутых сервисов.

---

## 8. ПРАКТИЧЕСКИЕ ЗАДАНИЯ

### Задание 1: Создание Dockerfile (20 минут)

**Цель:** Создать оптимизированный Docker образ для Go приложения

**Требования:**
- Multi-stage build
- Размер итогового образа < 20MB
- Использовать Alpine Linux
- Внедрить VERSION через build-arg

**Проверка:**
```bash
docker build -t go-simple-api:1.0.0 --build-arg VERSION=1.0.0 .
docker images | grep go-simple-api
docker run -p 8080:8080 go-simple-api:1.0.0
curl http://localhost:8080/health
```

### Задание 2: Matrix тестирование (20 минут)

**Цель:** Настроить тестирование на 3 версиях Go параллельно

**Требования:**
- Go версии: 1.21, 1.22, 1.23
- Использовать кэш для go modules
- Проверять форматирование только на последней версии
- Генерировать coverage отчет

**Проверка:**
- В GitHub Actions должно запуститься 3 параллельных job
- Время выполнения < 2 минут (с кэшем)

### Задание 3: Полный CI/CD workflow (30 минут)

**Цель:** Создать полноценный пайплайн

**Структура workflow:**
```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  # 1. Валидация кода
  validate:
    # gofmt, go vet

  # 2. Matrix тестирование
  test:
    # 3 версии Go, кэш, coverage

  # 3. Security scan
  security:
    # Trivy scan для кода

  # 4. Build & Push Docker
  docker:
    needs: [validate, test, security]
    # Multi-stage build
    # Push в ghcr.io
    # Правильные теги
```

**Проверка:**
```bash
# Создайте PR
git checkout -b feature/ci-cd
git add .
git commit -m "feat: add CI/CD pipeline"
git push origin feature/ci-cd

# Проверьте в GitHub:
# - Все 4 jobs запустились
# - test job создал 3 параллельных задачи
# - Docker образ появился в ghcr.io
```

### Задание 4: Настройка secrets (10 минут)

**Цель:** Безопасно настроить секреты для Docker Hub

**Шаги:**
1. Создать Docker Hub Access Token
2. Добавить в GitHub Secrets:
   - `DOCKERHUB_USERNAME`
   - `DOCKERHUB_TOKEN`
3. Обновить workflow для использования Docker Hub вместо ghcr.io

**Проверка:**
```bash
# После push образ должен появиться в Docker Hub
docker pull yourusername/go-simple-api:latest
```

---

## ЧЕКЛИСТ ВЫПОЛНЕНИЯ УРОКА

- [ ] Понимаю структуру CI/CD пайплайна
- [ ] Могу использовать matrix strategies для параллельных запусков
- [ ] Настроил кэширование для ускорения builds
- [ ] Создал оптимизированный Dockerfile < 20MB
- [ ] Настроил автоматический build и push образов
- [ ] Понимаю работу с secrets и permissions
- [ ] Могу переиспользовать код через composite actions
- [ ] Настроил правильное версионирование образов
- [ ] Знаю что такое SBOM и зачем он нужен
- [ ] Настроил генерацию SBOM для Docker образов

---

## ДОПОЛНИТЕЛЬНЫЕ РЕСУРСЫ

### Документация:
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [SemVer Specification](https://semver.org/)

### Инструменты:
- [act](https://github.com/nektos/act) - локальный запуск GitHub Actions
- [Docker Slim](https://github.com/docker-slim/docker-slim) - оптимизация образов
- [Trivy](https://github.com/aquasecurity/trivy) - сканер уязвимостей

### Следующий урок:
**Урок 3: Kubernetes + ArgoCD - GitOps deployment**
- Kubernetes манифесты (Deployment, Service)
- Kustomize для управления конфигурациями
- ArgoCD настройка и синхронизация
- Rollback стратегии
