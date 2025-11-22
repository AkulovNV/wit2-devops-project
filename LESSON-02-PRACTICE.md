# ПРАКТИЧЕСКИЕ ЗАДАНИЯ - УРОК 2

Этот документ содержит пошаговые инструкции для выполнения практических заданий урока 2.

---

## ЗАДАНИЕ 1: Docker Build (20 минут)

### Цель
Создать оптимизированный Docker образ и убедиться что он работает.

### Шаги выполнения

#### 1. Проверяем что все файлы на месте

```bash
ls -la Dockerfile .dockerignore
```

Должны быть:
- `Dockerfile` - multi-stage build
- `.dockerignore` - исключения для Docker

#### 2. Собираем Docker образ

```bash
docker build \
  -t go-simple-api:1.0.0 \
  --build-arg VERSION=1.0.0 \
  --build-arg BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  --build-arg GIT_COMMIT=$(git rev-parse --short HEAD) \
  .
```

**Что происходит:**
- `-t go-simple-api:1.0.0` - тег образа
- `--build-arg VERSION=1.0.0` - передаем версию в build
- Сборка проходит в 2 stage (builder → runtime)

#### 3. Проверяем размер образа

```bash
docker images | grep go-simple-api
```

**Ожидаемый результат:**
```
go-simple-api   1.0.0   abc123def456   2 minutes ago   15MB
```

Размер должен быть **< 20MB** (если больше - что-то не так с multi-stage)

#### 4. Запускаем контейнер

```bash
docker run -d \
  --name go-api-test \
  -p 8080:8080 \
  -e LOG_LEVEL=debug \
  go-simple-api:1.0.0
```

#### 5. Проверяем что API работает

```bash
# Health check
curl http://localhost:8080/health

# Version endpoint
curl http://localhost:8080/api/version

# Metrics
curl http://localhost:8080/metrics
```

**Ожидаемый результат для /health:**
```json
{
  "status": "ok",
  "timestamp": "2024-11-16T10:30:00Z",
  "version": "1.0.0"
}
```

#### 6. Проверяем логи

```bash
docker logs go-api-test
```

Должны видеть JSON логи с level=info

#### 7. Останавливаем и удаляем контейнер

```bash
docker stop go-api-test
docker rm go-api-test
```

### Проверка знаний

**Вопросы:**
1. Почему образ получился таким маленьким (15MB)?
   <details>
   <summary>Ответ</summary>
   Используется multi-stage build: сборка в golang:1.23-alpine (700MB+), а в финальный образ копируется только статический бинарник в alpine:latest
   </details>

2. Что делает флаг `-ldflags="-w -s"`?
   <details>
   <summary>Ответ</summary>
   -w удаляет DWARF debug информацию, -s удаляет таблицу символов. Это уменьшает размер бинарника на 20-30%
   </details>

3. Зачем нужен `.dockerignore`?
   <details>
   <summary>Ответ</summary>
   Исключает ненужные файлы из Docker build context, ускоряет build и уменьшает размер контекста
   </details>

---

## ЗАДАНИЕ 2: Matrix тестирование (20 минут)

### Цель
Настроить параллельное тестирование на 3 версиях Go.

### Шаги выполнения

#### 1. Проверяем workflow файл

```bash
cat .github/workflows/ci-cd.yml | grep -A 20 "job: test"
```

Должна быть matrix strategy с версиями: 1.21, 1.22, 1.23

#### 2. Локально тестируем на разных версиях

Используя Docker:

```bash
# Тест на Go 1.21
docker run --rm -v $(pwd):/app -w /app golang:1.21 go test -v

# Тест на Go 1.22
docker run --rm -v $(pwd):/app -w /app golang:1.22 go test -v

# Тест на Go 1.23
docker run --rm -v $(pwd):/app -w /app golang:1.23 go test -v
```

Все тесты должны проходить на всех версиях.

#### 3. Коммитим и пушим изменения

```bash
git add .github/workflows/ci-cd.yml
git commit -m "feat: add matrix testing for Go 1.21, 1.22, 1.23"
git push origin main
```

#### 4. Проверяем GitHub Actions

Переходим в GitHub → Actions → выбираем запущенный workflow

**Что должны увидеть:**
- 3 параллельных job'а для test:
  - Test on Go 1.21
  - Test on Go 1.22
  - Test on Go 1.23

#### 5. Анализируем время выполнения

Смотрим сколько времени заняла каждая job:
- Первый запуск (без кэша): ~2-3 минуты
- Второй запуск (с кэшем): ~30-60 секунд

### Эксперимент: добавляем новую версию

Попробуем добавить Go 1.20:

```yaml
strategy:
  fail-fast: false
  matrix:
    go-version: ['1.20', '1.21', '1.22', '1.23']
```

Коммитим и смотрим что произойдет. Скорее всего тесты упадут на 1.20 из-за несовместимости зависимостей.

### Проверка знаний

**Вопросы:**
1. Зачем нужен `fail-fast: false`?
   <details>
   <summary>Ответ</summary>
   Чтобы продолжить выполнение остальных jobs даже если одна упала. По умолчанию fail-fast: true останавливает все при первой ошибке
   </details>

2. Сколько jobs запустится при matrix из 3 версий Go и 2 OS?
   <details>
   <summary>Ответ</summary>
   6 jobs (3 версии × 2 OS = 6 комбинаций)
   </details>

3. Как запустить job только для одной комбинации matrix?
   <details>
   <summary>Ответ</summary>
   Использовать условие: `if: matrix.go-version == '1.23' && matrix.os == 'ubuntu-latest'`
   </details>

---

## ЗАДАНИЕ 3: Caching оптимизация (15 минут)

### Цель
Настроить кэширование и измерить ускорение.

### Шаги выполнения

#### 1. Измеряем baseline (без кэша)

Очищаем кэш в GitHub:
- Settings → Actions → Caches → Delete all caches

Запускаем workflow и засекаем время.

#### 2. Проверяем настройки кэша в workflow

```yaml
- name: Cache Go modules
  uses: actions/cache@v4
  with:
    path: |
      ~/go/pkg/mod
      ~/.cache/go-build
    key: ${{ runner.os }}-go-${{ matrix.go-version }}-${{ hashFiles('**/go.sum') }}
```

#### 3. Запускаем второй раз (с кэшем)

Делаем любое изменение (например в README) и пушим:

```bash
echo "test" >> README.md
git add README.md
git commit -m "test: trigger cache test"
git push
```

Засекаем время выполнения. Должно быть **в 3-4 раза быстрее**.

#### 4. Проверяем что кэш работает

В логах job должна быть строка:
```
Cache restored successfully
Cache Key: Linux-go-1.23-abc123...
```

#### 5. Эксперимент: меняем зависимости

Добавляем новую зависимость:

```bash
go get github.com/gin-gonic/gin
```

Коммитим и пушим. Теперь go.sum изменился, значит кэш будет создан заново.

### Docker cache

Проверяем настройки Docker cache:

```yaml
- name: Build and push
  uses: docker/build-push-action@v5
  with:
    cache-from: type=gha
    cache-to: type=gha,mode=max
```

**type=gha** - GitHub Actions cache для Docker layers

### Проверка знаний

**Вопросы:**
1. Что делает `hashFiles('**/go.sum')`?
   <details>
   <summary>Ответ</summary>
   Создает хэш всех файлов go.sum в проекте. Если зависимости изменились - хэш изменится и будет создан новый кэш
   </details>

2. Зачем нужен `restore-keys`?
   <details>
   <summary>Ответ</summary>
   Fallback ключи: если точный кэш не найден, используется частичное совпадение. Например если go.sum изменился, но версия Go та же
   </details>

3. Сколько времени хранится кэш?
   <details>
   <summary>Ответ</summary>
   7 дней с момента последнего использования. После этого автоматически удаляется
   </details>

---

## ЗАДАНИЕ 4: Docker Push в Registry (25 минут)

### Цель
Настроить автоматический push образов в GitHub Container Registry (ghcr.io).

### Шаги выполнения

#### 1. Включаем GitHub Packages

- Settings → Actions → General → Workflow permissions
- Выбираем "Read and write permissions"
- Сохраняем

#### 2. Проверяем permissions в workflow

```yaml
permissions:
  contents: read
  packages: write  # Для push в ghcr.io
```

#### 3. Проверяем настройки Docker login

```yaml
- name: Log in to GitHub Container Registry
  uses: docker/login-action@v3
  with:
    registry: ghcr.io
    username: ${{ github.actor }}
    password: ${{ secrets.GITHUB_TOKEN }}
```

**GITHUB_TOKEN** - автоматически доступен, не нужно создавать вручную.

#### 4. Коммитим и пушим

```bash
git add .
git commit -m "feat: add Docker build and push to ghcr.io"
git push origin main
```

#### 5. Проверяем что образ появился

После завершения workflow:
- Переходим в репозиторий → Packages
- Должен появиться образ `go-simple-api` с тегами

#### 6. Локально скачиваем образ из registry

```bash
# Логинимся в ghcr.io (нужен GitHub Personal Access Token)
echo $GITHUB_TOKEN | docker login ghcr.io -u YOUR_USERNAME --password-stdin

# Скачиваем образ
docker pull ghcr.io/YOUR_USERNAME/wit2-devops-project:main-abc1234

# Запускаем
docker run -p 8080:8080 ghcr.io/YOUR_USERNAME/wit2-devops-project:main-abc1234
```

#### 7. Проверяем теги образа

В зависимости от того как запущен workflow, должны быть разные теги:

**При push в main:**
```
ghcr.io/user/repo:main-abc1234
ghcr.io/user/repo:latest
```

**При создании тега v1.2.3:**
```
ghcr.io/user/repo:1.2.3
ghcr.io/user/repo:1.2
ghcr.io/user/repo:1
ghcr.io/user/repo:latest
```

**При PR:**
```
ghcr.io/user/repo:pr-123
```

### Эксперимент: создаем release

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

Проверяем что создались теги:
- `1.0.0`
- `1.0`
- `1`
- `latest`

### Проверка знаний

**Вопросы:**
1. В чем разница между `GITHUB_TOKEN` и Personal Access Token?
   <details>
   <summary>Ответ</summary>
   GITHUB_TOKEN автоматически создается для каждого workflow run и имеет ограниченные permissions. PAT создается вручную и может иметь больше прав
   </details>

2. Зачем нужно несколько тегов для одного образа (1.2.3, 1.2, 1)?
   <details>
   <summary>Ответ</summary>
   Для гибкости: можно зафиксировать точную версию (1.2.3) или получать патчи автоматически (1.2), или мажорные обновления (1)
   </details>

3. Что делает `docker/metadata-action`?
   <details>
   <summary>Ответ</summary>
   Автоматически генерирует теги и labels для Docker образа на основе Git событий (push, tag, PR)
   </details>

---

## ЗАДАНИЕ 5: Composite Action (20 минут)

### Цель
Создать переиспользуемый action для setup Go с кэшем.

### Шаги выполнения

#### 1. Проверяем созданный composite action

```bash
cat .github/actions/setup-go/action.yml
```

#### 2. Используем его в workflow

Заменяем дублирующийся код:

**Было:**
```yaml
- name: Setup Go
  uses: actions/setup-go@v5
  with:
    go-version: ${{ matrix.go-version }}

- name: Cache Go modules
  uses: actions/cache@v4
  with:
    path: ~/go/pkg/mod
    key: ...
```

**Стало:**
```yaml
- name: Setup Go with cache
  uses: ./.github/actions/setup-go
  with:
    go-version: ${{ matrix.go-version }}
```

#### 3. Создаем еще один composite action для Docker setup

`.github/actions/setup-docker/action.yml`:

```yaml
name: 'Setup Docker Buildx'
description: 'Setup Docker Buildx with cache'

runs:
  using: 'composite'
  steps:
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v3

    - name: Login to GitHub Container Registry
      uses: docker/login-action@v3
      with:
        registry: ghcr.io
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}
```

Использование:

```yaml
- name: Setup Docker
  uses: ./.github/actions/setup-docker
```

#### 4. Коммитим и проверяем

```bash
git add .github/actions/
git commit -m "feat: add composite actions for reusability"
git push
```

### Когда использовать Composite Actions?

- Повторяющиеся последовательности шагов
- Нужно использовать в нескольких workflows
- Хотим стандартизировать процесс

### Проверка знаний

**Вопросы:**
1. В чем разница между Composite Action и Reusable Workflow?
   <details>
   <summary>Ответ</summary>
   Composite Action - набор шагов (steps), Reusable Workflow - целая job или несколько jobs
   </details>

2. Можно ли в Composite Action использовать другие actions?
   <details>
   <summary>Ответ</summary>
   Да, можно использовать любые публичные или локальные actions через uses:
   </details>

---

## ЗАДАНИЕ 6: Security сканирование (15 минут)

### Цель
Настроить автоматическое сканирование уязвимостей.

### Шаги выполнения

#### 1. Проверяем Trivy job в workflow

```yaml
security:
  name: Security Scan
  steps:
    - uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'  # filesystem scan
```

#### 2. Запускаем локально

```bash
# Установка Trivy (macOS)
brew install trivy

# Сканируем код
trivy fs .

# Сканируем Docker образ
trivy image go-simple-api:1.0.0
```

#### 3. Проверяем результаты в GitHub Security

- Security → Code scanning alerts

Должны видеть результаты Trivy сканирования.

#### 4. Эксперимент: добавляем уязвимую зависимость

```bash
# Добавляем старую версию с известной уязвимостью
go get github.com/gin-gonic/gin@v1.7.0
```

Коммитим и смотрим что Trivy найдет уязвимости.

### Проверка знаний

**Вопросы:**
1. В чем разница между `scan-type: fs` и `scan-type: image`?
   <details>
   <summary>Ответ</summary>
   fs - сканирует файловую систему (исходный код), image - сканирует Docker образ
   </details>

2. Что такое SARIF?
   <details>
   <summary>Ответ</summary>
   Static Analysis Results Interchange Format - стандартный формат для результатов security анализа
   </details>

---

## ИТОГОВОЕ ЗАДАНИЕ: Полный Pipeline (30 минут)

### Цель
Собрать все вместе и создать полноценный CI/CD pipeline.

### Требования

1. ✅ Валидация кода (gofmt, go vet)
2. ✅ Matrix тестирование (Go 1.21, 1.22, 1.23)
3. ✅ Кэширование (Go modules и Docker layers)
4. ✅ Security сканирование (Trivy)
5. ✅ Docker build & push (ghcr.io)
6. ✅ Правильные permissions
7. ✅ Версионирование (semantic tags)

### Проверка

Создайте PR и убедитесь что:

```bash
git checkout -b feature/complete-pipeline
# Делаем какое-то изменение
echo "// Update" >> main.go
git add .
git commit -m "feat: complete CI/CD pipeline setup"
git push origin feature/complete-pipeline
```

В GitHub:
- Создаем PR
- Все проверки проходят зеленым
- Появляется summary с результатами

После merge в main:
- Docker образ появляется в ghcr.io
- Security alerts пустые
- Coverage > 70%

---

## TROUBLESHOOTING

### Проблема: Docker build fails with "permission denied"

**Решение:**
```yaml
permissions:
  packages: write  # Добавить эту permission
```

### Проблема: Cache не работает

**Проверьте:**
1. `hashFiles('**/go.sum')` правильно указан
2. Путь к cache правильный: `~/go/pkg/mod`
3. Кэш не старше 7 дней

### Проблема: Matrix tests слишком долго выполняются

**Оптимизация:**
1. Включить cache для Go modules
2. Использовать `actions/setup-go` с `cache: true`
3. Убрать ненужные версии из matrix

### Проблема: Trivy находит много уязвимостей

**Решения:**
1. Обновить зависимости: `go get -u ./...`
2. Использовать `ignore-unfixed: true`
3. Настроить .trivyignore для false positives

---

## ДОПОЛНИТЕЛЬНЫЕ ЧЕЛЛЕНДЖИ

### Челлендж 1: Reusable Workflow

Создайте reusable workflow для тестирования и используйте его в нескольких проектах.

### Челлендж 2: Matrix с exclusions

Настройте matrix с 4 версиями Go и 3 OS, но исключите Windows + Go 1.21.

### Челлендж 3: Conditional deployment

Настройте deployment только для main ветки и только если все тесты прошли.

### Челлендж 4: Release automation

Настройте автоматическое создание GitHub Release при push тега.

---

## ЗАКЛЮЧЕНИЕ

После выполнения всех заданий вы научились:

- ✅ Проектировать эффективный CI/CD pipeline
- ✅ Использовать matrix для параллельных запусков
- ✅ Оптимизировать builds с помощью caching
- ✅ Создавать оптимизированные Docker образы
- ✅ Работать с secrets и permissions
- ✅ Переиспользовать код через composite actions
- ✅ Версионировать Docker образы правильно
- ✅ Сканировать код на уязвимости

**Следующий шаг:** Урок 3 - Kubernetes и ArgoCD
