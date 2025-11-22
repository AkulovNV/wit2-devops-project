# GO SIMPLE API - Pet Project для DevOps Program

> Практический проект для изучения DevOps: от кода до production в Kubernetes через GitOps

[![CI/CD Pipeline](https://github.com/devops-mentor/wit2-devops-project/workflows/CI/CD%20Pipeline/badge.svg)](https://github.com/devops-mentor/wit2-devops-project/actions)
[![Docker Image](https://img.shields.io/badge/docker-ghcr.io-blue)](https://github.com/devops-mentor/wit2-devops-project/pkgs/container/wit2-devops-project)
[![Go Version](https://img.shields.io/badge/go-1.23-00ADD8)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

---

**Общее:**
- 🗺️ [ROADMAP.md](ROADMAP.md) - План всех 5 уроков программы

---

## 📁 СТРУКТУРА ПРОЕКТА
```
wit2-devops-project/
├── .github/
│   ├── workflows/
│   │   └── ci-cd.yml                  # GitHub Actions CI/CD pipeline
│   └── dependabot.yml                 # Dependabot конфигурация
├── .gitignore                         # Git ignore
├── k8s/
│   ├── base/                          # Kustomize base (для всех окружений)
│   │   ├── deployment.yaml            # K8s Deployment
│   │   ├── service.yaml               # K8s Service
│   │   ├── configmap.yaml             # Configuration
│   │   └── kustomization.yaml         # Kustomize base config
│   └── overlays/
│       ├── dev/                       # Development overlay
│       │   ├── kustomization.yaml     # Dev config (1 replica)
│       │   └── deployment-patch.yaml  # Dev-specific patches
│       └── prod/                      # Production overlay
│           ├── kustomization.yaml     # Prod config (3 replicas)
│           └── deployment-patch.yaml  # Prod-specific patches
├── main.go                            # Основное приложение
├── main_test.go                       # Unit тесты
├── go.mod                             # Go модули
├── Dockerfile                         # Multi-stage build (700MB → 15MB)
├── Makefile                           # Make targets для разработки
└── README.md                          # Документация
```

---

## 🚀 QUICK START

### 1. Клонирование и установка

```bash
git clone https://github.com/devops-mentor/wit2-devops-project.git
cd wit2-devops-project

# Download dependencies
go mod download

# Run locally
make run
```

### 2. Локальное тестирование

```bash
# Run tests
make test

# Run with coverage
make test-cover

# Check formatting
make fmt-check
```

### 3. Docker build & run

```bash
# Build image
make docker-build

# Run container
make docker-run

# Push to registry
make docker-push
```

---

## 📊 API ENDPOINTS

### Health Check
```bash
curl http://localhost:8080/health

# Response:
{
  "status": "ok",
  "timestamp": "2024-11-02T10:30:45Z",
  "version": "1.0.0"
}
```

### Version Information
```bash
curl http://localhost:8080/api/version

# Response:
{
  "version": "1.0.0",
  "timestamp": "2024-11-02T10:30:45Z"
}
```

### Prometheus Metrics
```bash
curl http://localhost:8080/metrics

# Включает:
# - http_requests_total (counter)
# - http_request_duration_seconds (histogram)
# - app_info (gauge)
```

---

## 🧪 TESTING

### Unit Tests

```bash
# Run all tests
go test -v

# Run with coverage
go test -v -coverprofile=coverage.out
go tool cover -html=coverage.out

# Expected: 8+ unit tests covering:
# - Health endpoint
# - Version endpoint
# - 404 handling
# - Content-Type validation
# - JSON parsing
# - Middleware functionality
```

### Integration Testing (в pipeline)

```yaml
# GitHub Actions выполняет:
- go test -v -race
- gofmt check
- go vet
- docker build & push
- trivy security scan
```

---

## 🐳 DOCKER

### Build Image

```bash
# Multi-stage build
docker build -t go-simple-api:1.0.0 .

# Check size
docker images | grep go-simple-api
```

## ПОЛЕЗНЫЕ КОМАНДЫ

### GitHub CLI (gh)

```bash
# Создать PR
gh pr create --title "Title" --body "Description"

# Список PR
gh pr list

# Merge PR
gh pr merge 123 --merge

# Создать release
gh release create v1.0.0 --title "Release v1.0.0" --notes "Changes"

# Просмотр workflow runs
gh run list
gh run view 123456

# Логи workflow
gh run view 123456 --log
```

### Trivy (Security Scanner)

```bash
# Установка (macOS)
brew install trivy

# Сканировать код
trivy fs .

# Сканировать Docker образ
trivy image myapp:1.0.0

# Только CRITICAL и HIGH
trivy image --severity CRITICAL,HIGH myapp:1.0.0

# JSON output
trivy image -f json -o results.json myapp:1.0.0

# SARIF для GitHub
trivy image -f sarif -o trivy-results.sarif myapp:1.0.0
```

### Make (из Makefile)

```bash
# Помощь
make help

# Локальный запуск
make run

# Тесты
make test
make test-cover

# Линтинг
make lint
make fmt-check

# Docker
make docker-build
make docker-run
make docker-push

# Билд
make build
VERSION=2.0.0 make build  # С кастомной версией
```

---

## DEBUGGING GITHUB ACTIONS

### Локальный запуск (act)

```bash
# Установка
brew install act

# Запуск workflow
act push

# Конкретная job
act -j test

# С секретами
act -s GITHUB_TOKEN=xxx

# Dry run
act -n
```

### Enable Debug Logging

В GitHub Settings → Secrets добавить:
- `ACTIONS_RUNNER_DEBUG` = `true`
- `ACTIONS_STEP_DEBUG` = `true`

### SSH в runner (при ошибке)

```yaml
- name: Setup tmate session
  if: failure()
  uses: mxschmitt/action-tmate@v3
```

---

## SHORTCUTS

### Context переменные

```yaml
${{ github.sha }}           # Commit SHA
${{ github.ref }}           # refs/heads/main
${{ github.ref_name }}      # main
${{ github.actor }}         # username
${{ github.repository }}    # owner/repo
${{ runner.os }}            # Linux, macOS, Windows
${{ runner.temp }}          # /tmp/...
${{ secrets.GITHUB_TOKEN }} # Auto token
```

### Функции

```yaml
# Hash файлов
${{ hashFiles('**/go.sum') }}

# Форматирование строк
${{ format('Hello {0}', 'World') }}

# Условия
${{ github.ref == 'refs/heads/main' }}

# JSON
${{ toJSON(matrix) }}
${{ fromJSON('{"key": "value"}') }}

# Contains
${{ contains(github.ref, 'release') }}

# StartsWith / EndsWith
${{ startsWith(github.ref, 'refs/tags/v') }}
${{ endsWith(github.ref, '-stable') }}
```
