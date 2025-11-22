# Инструкция для ментора - Урок 2

Пошаговый план проведения второго урока по CI/CD с GitHub Actions.

---

## ПОДГОТОВКА К УРОКУ (за день до)

### 1. Проверка окружения менти

Убедитесь что у менти установлено:

```bash
# Docker
docker --version
# Должно быть: Docker version 20.10.0 или выше

# Go
go version
# Должно быть: go1.21 или выше

# Git
git --version

# GitHub CLI (опционально, но полезно)
gh --version
```

### 2. Форк репозитория

Попросите менти сделать fork или создать свой репозиторий из template:

```bash
# Клонирование
git clone https://github.com/YOUR_USERNAME/wit2-devops-project.git
cd wit2-devops-project

# Проверка что все работает
make test
```

### 3. GitHub Settings

Убедитесь что в репозитории включены:
- Settings → Actions → General → "Allow all actions"
- Settings → Actions → General → Workflow permissions → "Read and write permissions"

---

## СТРУКТУРА УРОКА (2 часа)

### Блок 1: Теория - Проектирование пайплайна (15 минут)

**Цель:** Объяснить общую структуру CI/CD пайплайна

**Что рассказать:**

1. **Идеализированный пайплайн:**
   ```
   CODE → BUILD → TEST → PACKAGE → DEPLOY
   ```

2. **Показать на доске/слайдах:**
   - Стадии пайплайна
   - Как они связаны
   - Почему это называется "пайплайн"

3. **Практический пример:**
   - Открыть `.github/workflows/ci-cd.yml`
   - Показать структуру: `on`, `jobs`, `steps`
   - Объяснить зависимости через `needs:`

**Задание (5 минут):**
- Попросите менти нарисовать схему пайплайна для их текущего проекта на работе
- Обсудите что можно автоматизировать

---

### Блок 2: Matrix Strategies (20 минут)

**Цель:** Научить параллельному запуску с разными параметрами

**Демонстрация (10 минут):**

1. Откройте `LESSON-02.md` → секция Matrix Strategies
2. Объясните синтаксис:
   ```yaml
   strategy:
     matrix:
       go-version: ['1.21', '1.22', '1.23']
   ```

3. **Live coding:**
   - Создайте простой workflow с matrix
   - Покажите как менять параметры
   - Объясните `fail-fast: false`

4. **Когда использовать:**
   - Cross-platform тестирование
   - Разные версии runtime
   - Параллельные независимые задачи

**Практика (10 минут):**

```bash
# Задание: добавить matrix в workflow
cd .github/workflows
# Менти должна добавить matrix для Go версий
```

**Проверка:**
```bash
git add .
git commit -m "feat: add matrix testing"
git push origin main
```

Зайти в GitHub → Actions и показать 3 параллельных job'а

---

### Блок 3: Caching (15 минут)

**Цель:** Ускорить builds через кэширование

**Теория (5 минут):**

1. **Объясните проблему:**
   - Без кэша: каждый раз скачиваем зависимости (2-3 минуты)
   - С кэшем: используем сохраненные (30 секунд)

2. **Как работает кэш:**
   - `key` - уникальный идентификатор
   - `hashFiles('**/go.sum')` - хэш файла зависимостей
   - При изменении зависимостей → новый кэш

3. **Покажите в коде:**
   ```yaml
   - uses: actions/cache@v4
     with:
       path: ~/go/pkg/mod
       key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
   ```

**Практика (10 минут):**

1. **Измерение baseline:**
   - Settings → Actions → Caches → Delete all
   - Запустить workflow, засечь время

2. **С кэшем:**
   - Сделать изменение в README
   - Снова запустить workflow
   - Сравнить время (должно быть в 3-4 раза быстрее)

**Ожидаемые результаты:**
- Без кэша: ~2-3 минуты
- С кэшем: ~30-60 секунд

---

### Блок 4: Docker Build (25 минут)

**Цель:** Создать оптимизированный Docker образ

**Теория (10 минут):**

1. **Multi-stage build:**
   - Показать `Dockerfile`
   - Объяснить 2 stage: builder → runtime
   - Почему это уменьшает размер (700MB → 15MB)

2. **Оптимизации:**
   ```dockerfile
   # ❌ ПЛОХО: всегда пересобирается
   COPY . .
   RUN go mod download

   # ✅ ХОРОШО: кэширует зависимости
   COPY go.mod go.sum ./
   RUN go mod download
   COPY . .
   ```

3. **Security:**
   - Non-root user
   - Minimal base image (alpine)
   - Health check

**Практика (15 минут):**

```bash
# 1. Build образа
docker build -t go-simple-api:1.0.0 \
  --build-arg VERSION=1.0.0 \
  .

# 2. Проверка размера
docker images | grep go-simple-api
# Должно быть ~15MB

# 3. Запуск
docker run -d -p 8080:8080 --name test-api go-simple-api:1.0.0

# 4. Проверка
curl http://localhost:8080/health

# 5. Cleanup
docker stop test-api && docker rm test-api
```

**Важные моменты:**
- Если образ > 20MB → что-то не так
- Объяснить `.dockerignore`
- Показать слои через `docker history`

---

### ПЕРЕРЫВ (10 минут) ☕

---

### Блок 5: Docker Push & Secrets (20 минут)

**Цель:** Автоматизировать push образов в registry

**Теория (5 минут):**

1. **Container Registry опции:**
   - Docker Hub (публичный)
   - GitHub Container Registry (ghcr.io)
   - AWS ECR, GCP GCR, Azure ACR

2. **Secrets:**
   - `GITHUB_TOKEN` - автоматический
   - Custom secrets - создаются вручную

**Практика (15 минут):**

1. **Проверка permissions:**
   ```yaml
   permissions:
     contents: read
     packages: write
   ```

2. **Добавление Docker job в workflow:**
   - Открыть `.github/workflows/ci-cd.yml`
   - Показать секцию `docker:`
   - Объяснить каждый шаг

3. **Запуск:**
   ```bash
   git add .
   git commit -m "feat: add Docker build and push"
   git push origin main
   ```

4. **Проверка:**
   - GitHub → Actions → Watch workflow run
   - После завершения → Packages
   - Должен появиться образ

**Troubleshooting:**
Если не работает:
- Проверить permissions
- Проверить GITHUB_TOKEN
- Посмотреть логи job'а

---

### Блок 6: Версионирование (15 минут)

**Цель:** Научить правильно тегировать образы

**Теория (5 минут):**

1. **SemVer:**
   ```
   MAJOR.MINOR.PATCH
   1.2.3
   ```

2. **Docker tags стратегии:**
   - SHA: `main-abc1234`
   - Semantic: `1.2.3`, `1.2`, `1`
   - Latest: `latest`

3. **Показать docker/metadata-action:**
   ```yaml
   tags: |
     type=sha,prefix={{branch}}-
     type=semver,pattern={{version}}
   ```

**Практика (10 минут):**

```bash
# 1. Создать tag
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 2. Проверить в Actions
# Должны создаться теги: 1.0.0, 1.0, 1, latest

# 3. Pull образ
docker pull ghcr.io/USERNAME/PROJECT:1.0.0
docker pull ghcr.io/USERNAME/PROJECT:latest
```

**Обсудить:**
- Когда использовать каждый тип тега
- Как это связано с deployment (dev → staging → prod)

---

### Блок 7: Переиспользование кода (15 минут)

**Цель:** DRY принцип в GitHub Actions

**Теория (5 минут):**

1. **Composite Actions:**
   - Для переиспользования steps
   - Локальные actions в `.github/actions/`

2. **Reusable Workflows:**
   - Для переиспользования целых jobs
   - Можно вызывать из других workflows

**Практика (10 минут):**

1. **Показать готовый composite action:**
   ```bash
   cat .github/actions/setup-go/action.yml
   ```

2. **Использование:**
   ```yaml
   - uses: ./.github/actions/setup-go
     with:
       go-version: '1.23'
   ```

3. **Задание:**
   - Попросите менти создать composite action для Docker setup
   - Обсудите где еще можно применить

**Когда использовать:**
- Повторяющийся код в нескольких workflows
- Стандартизация процессов в команде
- Sharing между проектами

---

### Блок 8: Security Scanning (10 минут)

**Цель:** Базовое понимание security в CI/CD

**Быстрая демонстрация:**

1. **Trivy в workflow:**
   ```yaml
   - uses: aquasecurity/trivy-action@master
     with:
       scan-type: 'fs'
   ```

2. **Локальный запуск:**
   ```bash
   brew install trivy
   trivy fs .
   trivy image go-simple-api:1.0.0
   ```

3. **GitHub Security:**
   - Security → Code scanning alerts
   - Показать как выглядят alerts

**Не углубляться сильно** - это отдельная большая тема.

**Главное донести:**
- Security scanning должен быть в каждом пайплайне
- Fail fast если найдены CRITICAL уязвимости
- Регулярно обновлять зависимости

---

### Блок 9: Итоговое задание (15 минут)

**Цель:** Собрать все вместе

**Задание:**

Создать PR с полным CI/CD пайплайном:

```bash
git checkout -b feature/complete-cicd

# Убедиться что все файлы на месте:
# - Dockerfile
# - .dockerignore
# - .github/workflows/ci-cd.yml
# - .github/actions/setup-go/

git add .
git commit -m "feat: complete CI/CD pipeline setup"
git push origin feature/complete-cicd
```

**Проверка:**

В GitHub создать PR и убедиться что:
- ✅ Все jobs запустились
- ✅ Matrix создал 3 параллельных теста
- ✅ Кэш работает (видно в логах)
- ✅ Docker образ собрался
- ✅ Security scan прошел
- ✅ Все зеленое

После merge:
- ✅ Образ появился в ghcr.io
- ✅ С правильными тегами

---

## ОТВЕТЫ НА ЧАСТЫЕ ВОПРОСЫ

### Q: Почему не Docker Hub, а ghcr.io?

**A:**
- GitHub Container Registry бесплатный для публичных репо
- Интеграция с GitHub (те же permissions)
- Не нужно создавать отдельные токены
- Лучше для learning purposes

### Q: Зачем 3 версии Go тестировать?

**A:**
- Проверка backward compatibility
- Практика matrix strategies
- В реальности часто нужно поддерживать несколько версий

### Q: Можно ли использовать другой CI/CD (GitLab, Jenkins)?

**A:**
- Концепции одинаковые
- GitHub Actions выбран для простоты
- После понимания легко переключиться на другие

### Q: Сколько стоит GitHub Actions?

**A:**
- Публичные репо: бесплатно (unlimited minutes)
- Приватные: 2000 минут/месяц бесплатно
- Этого хватает для learning

### Q: Нужен ли мне Kubernetes чтобы использовать Docker?

**A:**
- Нет! Docker самостоятельный инструмент
- K8s - это следующий уровень (урок 3)
- Можно использовать Docker без K8s

---

## ДОМАШНЕЕ ЗАДАНИЕ

### Обязательное:

1. **Завершить все задания из `LESSON-02-PRACTICE.md`**
   - Задание 1: Docker build ✅
   - Задание 2: Matrix testing ✅
   - Задание 3: Полный workflow ✅

2. **Эксперименты:**
   - Добавить линтинг (golangci-lint)
   - Добавить badge в README (Build Status)
   - Попробовать разные Docker base images

### Дополнительное (опционально):

1. **Reusable workflow:**
   - Создать reusable workflow для тестирования
   - Использовать в нескольких проектах

2. **Advanced matrix:**
   - Добавить OS в matrix (ubuntu, macos)
   - Использовать include/exclude

3. **Release automation:**
   - Настроить автосоздание GitHub Release
   - Генерировать CHANGELOG автоматически

---

## МАТЕРИАЛЫ ДЛЯ МЕНТИ

Отправьте менти ссылки на:

1. **📖 Основной материал:**
   - `LESSON-02.md` - теория
   - `LESSON-02-PRACTICE.md` - практика
   - `CHEATSHEET-LESSON-02.md` - шпаргалка

2. **📚 Дополнительное чтение:**
   - [GitHub Actions Documentation](https://docs.github.com/en/actions)
   - [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
   - [12 Factor App](https://12factor.net/)

3. **🎥 Видео (если есть время):**
   - GitHub Actions Tutorial (YouTube)
   - Docker Deep Dive

---

## ПОДГОТОВКА К СЛЕДУЮЩЕМУ УРОКУ

**Урок 3: Kubernetes & Kustomize**

До следующего урока менти должна:

1. **Установить локальный K8s:**
   ```bash
   # Minikube (рекомендуется для начинающих)
   brew install minikube
   minikube start

   # Или Kind
   brew install kind
   kind create cluster
   ```

2. **Установить kubectl:**
   ```bash
   brew install kubectl
   kubectl version
   ```

3. **Пройти K8s basics:**
   - [Kubernetes Basics Tutorial](https://kubernetes.io/docs/tutorials/kubernetes-basics/)
   - Понять что такое Pod, Deployment, Service

4. **Посмотреть структуру k8s/ директории:**
   ```bash
   tree k8s/
   ```

---

## CHECKLIST ДЛЯ МЕНТОРА

Перед уроком:
- [ ] Проверил что все файлы на месте
- [ ] Протестировал workflow локально (act)
- [ ] Подготовил примеры для демонстрации
- [ ] Создал test branch для демо

Во время урока:
- [ ] Проверил окружение менти
- [ ] Показал live coding примеры
- [ ] Ответил на вопросы
- [ ] Менти выполнила практические задания

После урока:
- [ ] Отправил материалы
- [ ] Дал домашнее задание
- [ ] Назначил дату следующего урока
- [ ] Попросил фидбек

---

**Удачи в проведении урока! 🚀**
