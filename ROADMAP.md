# DevOps Learning Roadmap -- 5 занятий

Структурированный план обучения DevOps с фокусом на CI/CD и GitOps.

---

## Общая структура программы

```
Урок 1: Git & GitHub Basics             [done]
   |
Урок 2: CI/CD Pipeline (GitHub Actions) [done]
   |
Урок 3: Kubernetes (Manifests + Helm)   [done]
   |
Урок 4: ArgoCD & GitOps                 [current]
   |
Урок 5: Monitoring & Best Practices     [planned]
```

---

## Урок 1: Git & GitHub Basics

**Статус:** Завершен

**Что закрыли:**
- Git основы (clone, commit, push, pull)
- Branching strategies (main, feature branches)
- Pull Requests и Code Review
- Markdown для документации

**Pet-проект:** Создан Go Simple API, настроен GitHub-репозиторий, написаны unit-тесты

---

## Урок 2: CI/CD Pipeline -- GitHub Actions

**Статус:** Завершен

**Что закрыли:**
- Проектирование CI/CD пайплайна
- Matrix strategies (параллельные тесты на Go 1.25/1.26)
- Caching зависимостей
- Docker multi-stage build и push в ghcr.io
- Secrets, permissions, SBOM
- Composite actions и reusable workflows
- Версионирование (SemVer, Docker tags)

**Материалы:** [`lesson-02/`](./lesson-02/)

---

## Урок 3: Kubernetes -- развертывание приложения

**Статус:** Завершен

**Что закрыли:**
- Kubernetes абстракции (Pod, Deployment, Service, Ingress, Namespace)
- YAML-манифесты для приложения
- Helm charts (шаблонизация, install/upgrade/rollback)
- Ingress-nginx controller
- Диагностика (describe, logs, events)

**Материалы:** [`lesson-03/`](./lesson-03/)

---

## Урок 4: ArgoCD & GitOps

**Статус:** Текущий

**Что закроем:**
- GitOps концепция (Push vs Pull модель)
- ArgoCD установка и настройка
- Создание ArgoCD Application (UI, YAML, CLI)
- Sync strategies (Manual, Auto, Self-Heal, Prune)
- Обновление приложения через Git
- Rollback (через ArgoCD и через git revert)
- ArgoCD CLI
- Полная картина: CI (GitHub Actions) + CD (ArgoCD)

**Материалы:** [`lesson-04/`](./lesson-04/)

**Стенд:**
- Colima с K8s (k3s)
- ingress-nginx
- ArgoCD (Helm)

---

## Урок 5: Monitoring & Best Practices

**Статус:** Планируется

**Что закроем:**
- Prometheus метрики (уже есть в коде: `/metrics`)
- Grafana дашборды
- Structured logging
- Alerting
- Best practices: CI/CD, Docker, K8s, Security

---

## Итоговый pipeline (после всех уроков)

```
Developer --> Git Push --> GitHub Actions (CI)
                              |
                              +-- Test (matrix Go 1.25/1.26)
                              +-- Security scan (Trivy)
                              +-- Docker Build + Push (ghcr.io)
                              +-- Update image tag in Git
                              |
                              v
                         Git Repo (source of truth)
                              |
                              v
                    ArgoCD (auto sync)
                              |
                              v
                    Kubernetes
                              |
                              v
                    Prometheus --> Grafana --> Alerts
```

---

## Структура проекта

```
wit/
├── lessons/
│   ├── lesson-02/           # CI/CD Pipeline
│   │   ├── LESSON-02.md
│   │   ├── LESSON-02-INSTRUCTOR.md
│   │   ├── LESSON-02-PRACTICE.md
│   │   └── *.yml            # Примеры reusable workflows
│   ├── lesson-03/           # Kubernetes
│   │   ├── LESSON-03.md
│   │   ├── LESSON-03-INSTRUCTOR.md
│   │   └── LESSON-03-PRACTICE.md
│   └── lesson-04/           # ArgoCD & GitOps
│       ├── LESSON-04.md
│       ├── LESSON-04-INSTRUCTOR.md
│       └── LESSON-04-PRACTICE.md
└── wit2-devops-project/     # Pet-проект
    ├── main.go              # Go Simple API
    ├── Dockerfile           # Multi-stage build
    ├── Makefile
    ├── .github/workflows/   # CI/CD
    └── k8s/
        ├── manifests/       # Обычные K8s манифесты
        ├── helm/            # Helm chart
        └── argocd/          # ArgoCD Application манифесты
```

---

## Чеклист прогресса

### После Урока 1:
- [x] Git workflow (branches, PR, merge)
- [x] Базовые unit-тесты
- [x] Документация

### После Урока 2:
- [x] CI/CD pipeline (GitHub Actions)
- [x] Matrix testing
- [x] Docker build < 20MB
- [x] Автопуш в registry
- [x] Security scanning

### После Урока 3:
- [x] K8s манифесты (Deployment, Service, Ingress)
- [x] Helm chart (install, upgrade, rollback)
- [x] Локальный кластер (Colima)

### После Урока 4:
- [ ] ArgoCD установлен
- [ ] GitOps workflow настроен
- [ ] Auto sync + self-heal
- [ ] Rollback через git revert

### После Урока 5:
- [ ] Prometheus + Grafana
- [ ] Дашборды и алерты
- [ ] Production checklist
