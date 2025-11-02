# GO SIMPLE API - Pet Project для DevOps Program

---

## 📁 СТРУКТУРА ПРОЕКТА

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
