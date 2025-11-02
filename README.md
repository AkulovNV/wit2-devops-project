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

## 🎯 ИСПОЛЬЗОВАНИЕ В ПРОГРАММЕ

### Lesson 1: GitHub Actions Advanced
- ✅ Matrix strategies (testing на 3 версиях Go)
- ✅ Caching dependencies
- ✅ Docker build & push
- ✅ **NEW**: Workflow templates для переиспользования

**Файл:** `.github/workflows/ci-cd.yml`

---

### Lesson 2: Docker Optimization
- ✅ Multi-stage Dockerfile (700MB → 15MB)
- ✅ Security scanning (Trivy)
- ✅ Semantic versioning
- ✅ **NEW**: Dependabot для отслеживания зависимостей

**Файлы:** `Dockerfile`, `.github/dependabot.yml`

---

### Lesson 3: Kubernetes Basics
- ✅ Deployment manifests
- ✅ Service configuration
- ✅ Health checks (readiness + liveness)
- ✅ Resource limits

**Файлы:** `k8s/base/deployment.yaml`, `k8s/base/service.yaml`

---

### Lesson 4: GitOps with ArgoCD
- ✅ Kustomize base + overlays
- ✅ Environment-specific configurations (dev vs prod)
- ✅ Git-based deployment

**Файлы:** `k8s/base/kustomization.yaml`, `k8s/overlays/*/kustomization.yaml`

---

### Lesson 5: Production Readiness & Monitoring
- ✅ Prometheus metrics built-in
- ✅ Graceful shutdown
- ✅ Structured JSON logging
- ✅ **NEW**: Prometheus + Grafana stack integration

**Файл:** `main.go` (metrics collection)

---

## 🚀 QUICK START

### 1. Клонирование и установка

```bash
git clone https://github.com/devops-mentor/go-simple-api.git
cd go-simple-api

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

### 4. Kubernetes deployment

```bash
# Deploy to dev
make k8s-dev

# Deploy to prod
make k8s-prod

# Check status
make k8s-status
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
# Builder stage: ~700MB
# Final image: ~15MB (distroless alpine)
```

### Push to Registry

```bash
# Login
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin

# Tag
docker tag go-simple-api:1.0.0 ghcr.io/USERNAME/go-simple-api:1.0.0

# Push
docker push ghcr.io/USERNAME/go-simple-api:1.0.0
```

---

## ☸️ KUBERNETES

### Deploy with Kustomize

```bash
# Deploy dev (1 replica, debug logs)
kubectl apply -k k8s/overlays/dev

# Deploy prod (3 replicas, warn logs, higher resources)
kubectl apply -k k8s/overlays/prod

# Check
kubectl get pods -n dev
kubectl get pods -n prod
```

### Health Checks

```yaml
readinessProbe:      # Ready to serve traffic?
  httpGet: /health
  
livenessProbe:       # Is pod alive?
  httpGet: /health
  
startupProbe:        # For slow-starting apps
  httpGet: /health
```

### Scaling

```bash
# Scale deployment
kubectl scale deployment/api -n dev --replicas=3

# Rolling update
kubectl rollout restart deployment/api -n dev
```

---

## 📝 GITHUB ACTIONS PIPELINE

### ci-cd.yml Stages

1. **Test** (Matrix: Go 1.21, 1.22, 1.23)
   - Run tests on multiple versions
   - Cache dependencies
   - Upload coverage

2. **Lint**
   - gofmt check
   - go vet

3. **Build Docker**
   - Multi-stage build
   - Push to ghcr.io
   - Tag with version or git-sha

4. **Security Scan**
   - Trivy vulnerability scanner
   - Upload to GitHub Security

5. **Update Manifests** (GitOps)
   - Update image tag in K8s manifests
   - Commit to repo
   - ArgoCD picks up changes

---

## 🔧 FEATURES

### Application

- ✅ HTTP API with 2 endpoints
- ✅ Prometheus metrics (built-in)
- ✅ JSON structured logging (logrus)
- ✅ Graceful shutdown (SIGTERM/SIGINT)
- ✅ Health checks for K8s
- ✅ Request middleware (metrics collection)
- ✅ Error handling

### DevOps

- ✅ GitHub Actions CI/CD
- ✅ Docker multi-stage build
- ✅ Kubernetes manifests
- ✅ Kustomize overlays (dev/prod)
- ✅ Dependabot automation
- ✅ Security scanning (Trivy)
- ✅ Prometheus metrics integration
- ✅ GitOps ready (ArgoCD)

### Code Quality

- ✅ Unit tests (8+ tests)
- ✅ >80% code coverage
- ✅ gofmt compliance
- ✅ go vet passing
- ✅ Clear error messages

---

## 📚 LEARNING OUTCOMES

После завершения программы с этим pet-проектом, вы сможете:

✅ **Lesson 1**: Писать GitHub Actions workflows с матрицей и кэшем  
✅ **Lesson 2**: Оптимизировать Docker images (700MB → 15MB)  
✅ **Lesson 3**: Писать K8s manifests (Deployment, Service, probes)  
✅ **Lesson 4**: Использовать Kustomize + ArgoCD для GitOps  
✅ **Lesson 5**: Интегрировать Prometheus мониторинг  

**Result:** Production-ready приложение с полным DevOps pipeline

---

## 🚀 DEPLOYMENT WORKFLOW

```
Developer pushes code
        ↓
GitHub Actions (CI):
  ├─ Run tests (matrix: 1.21, 1.22, 1.23)
  ├─ Lint & format
  ├─ Build Docker image
  ├─ Security scan (Trivy)
  ├─ Push to ghcr.io
  └─ Update K8s manifests
        ↓
Git repo (manifests updated)
        ↓
ArgoCD (CD):
  ├─ Detect manifest change
  ├─ Compare with K8s state
  └─ Apply via kubectl
        ↓
Kubernetes:
  ├─ Rolling update (zero-downtime)
  ├─ Health checks (readiness + liveness)
  └─ Prometheus metrics collection
        ↓
Monitoring:
  ├─ Prometheus scrapes metrics
  ├─ Grafana visualizes dashboards
  └─ Alerts on errors/slowness
```

---

## 📖 HELPFUL COMMANDS

```bash
# Development
make run              # Run locally
make test             # Run tests
make build            # Build binary
make clean            # Clean artifacts

# Docker
make docker-build     # Build image
make docker-run       # Run container
make docker-push      # Push to registry

# Kubernetes
make k8s-dev          # Deploy to dev
make k8s-prod         # Deploy to prod
make k8s-status       # Check status
make k8s-clean        # Delete deployments

# Help
make help             # Show all targets
```

---

## 🔗 USEFUL LINKS

- Go Docs: https://pkg.go.dev/
- Docker: https://docs.docker.com/
- Kubernetes: https://kubernetes.io/docs/
- GitHub Actions: https://docs.github.com/en/actions
- Prometheus: https://prometheus.io/docs/
- ArgoCD: https://argo-cd.readthedocs.io/

---

## ❓ TROUBLESHOOTING

### "Cannot connect to port 8080"
```bash
# Check if process is running
ps aux | grep app
lsof -i :8080

# Kill process
pkill -f "go run main.go"
```

### "Docker image too large"
```bash
# Should be ~15MB (multi-stage build)
docker images | grep go-simple-api

# If large, check Dockerfile is multi-stage
```

### "Kubernetes pods not starting"
```bash
# Check pod status
kubectl describe pod POD_NAME -n dev

# Check logs
kubectl logs POD_NAME -n dev

# Check readiness probe
kubectl get events -n dev
```

---

## 📄 NOTES

- Multi-stage Dockerfile обязателен (Lesson 2)
- GitHub Actions workflow должен быть в `.github/workflows/`
- Kustomize overlays НЕ должны копировать базовые файлы
- ArgoCD Application должна указывать на K8s манифесты
- Prometheus metrics собираются автоматически

---

## ✅ CHECKLIST ДЛЯ ИСПОЛЬЗОВАНИЯ

- [ ] Код скачан и работает локально
- [ ] Tests проходят
- [ ] Docker image собирается (15MB)
- [ ] Image может быть запущен локально
- [ ] Kubernetes manifests валидны
- [ ] Kustomize overlays работают
- [ ] GitHub Actions workflow запускается
- [ ] Dependabot создает PRs для updates
- [ ] ArgoCD может синхронизировать
- [ ] Prometheus metrics собираются

---

**Ready to start DevOps learning! 🚀**