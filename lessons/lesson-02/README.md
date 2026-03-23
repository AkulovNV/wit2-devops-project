# Reusable Workflows

Эта директория содержит примеры **переиспользуемых workflows** (reusable workflows) для демонстрации концепций CI/CD.

## 📚 Что такое Reusable Workflows?

Reusable workflows позволяют:
- **Избежать дублирования** кода между разными workflows
- **Стандартизировать** процессы CI/CD
- **Централизовать** обновления и улучшения
- **Упростить** сложные pipelines

## 📁 Доступные Reusable Workflows

### 1. `test-reusable.yml` - Тестирование Go приложений

**Описание:** Запускает тесты с указанной версией Go, проверяет покрытие кода.

**Входные параметры:**
- `go-version` (required) - версия Go для тестирования
- `coverage-threshold` (optional, default: 70) - минимальное покрытие кода
- `run-race-detector` (optional, default: true) - использовать race detector

**Выходные данные:**
- `coverage` - процент покрытия кода
- `test-result` - результат выполнения тестов

**Пример использования:**
```yaml
jobs:
  test:
    uses: ./.github/workflows/test-reusable.yml
    with:
      go-version: '1.23'
      coverage-threshold: 75
      run-race-detector: true
```

**Пример с Matrix:**
```yaml
jobs:
  test-matrix:
    strategy:
      matrix:
        go-version: ['1.21', '1.22', '1.23']
    uses: ./.github/workflows/test-reusable.yml
    with:
      go-version: ${{ matrix.go-version }}
      coverage-threshold: 70
```

---

### 2. `docker-build-reusable.yml` - Сборка Docker образов

**Описание:** Собирает и публикует Docker образы с поддержкой кэширования и метаданных.

**Входные параметры:**
- `image-name` (required) - имя образа (без registry)
- `dockerfile-path` (optional, default: './Dockerfile') - путь к Dockerfile
- `build-context` (optional, default: '.') - контекст сборки
- `push-image` (optional, default: true) - публиковать ли образ
- `registry` (optional, default: 'ghcr.io') - URL registry

**Секреты:**
- `registry-username` (optional) - имя пользователя registry
- `registry-password` (optional) - пароль или токен registry

**Выходные данные:**
- `image-tag` - полный тег образа с registry
- `digest` - digest образа

**Пример использования:**
```yaml
jobs:
  build:
    uses: ./.github/workflows/docker-build-reusable.yml
    with:
      image-name: ${{ github.repository }}
      push-image: true
      registry: 'ghcr.io'
    secrets:
      registry-username: ${{ github.actor }}
      registry-password: ${{ secrets.GITHUB_TOKEN }}
```

**Пример с использованием outputs:**
```yaml
jobs:
  build-docker:
    uses: ./.github/workflows/docker-build-reusable.yml
    with:
      image-name: 'myorg/myapp'

  scan-image:
    needs: build-docker
    runs-on: ubuntu-latest
    steps:
      - name: Scan Docker image
        run: |
          echo "Scanning: ${{ needs.build-docker.outputs.image-tag }}"
          trivy image ${{ needs.build-docker.outputs.image-tag }}
```

---

## 🎯 Примеры

### `example-reusable-usage.yml`

Демонстрационный workflow, показывающий различные способы использования reusable workflows:

1. **Одиночный вызов** с конкретными параметрами
2. **Matrix strategy** с различными версиями
3. **Использование outputs** для последующих jobs

Запуск:
```bash
# Через GitHub UI
Actions → Example - Using Reusable Workflow → Run workflow

# Или через GitHub CLI
gh workflow run example-reusable-usage.yml
```

---

## 🔄 Сравнение: Composite Actions vs Reusable Workflows

| Критерий | Composite Actions | Reusable Workflows |
|----------|-------------------|-------------------|
| **Расположение** | `.github/actions/` | `.github/workflows/` |
| **Использование** | Внутри steps | Внутри jobs |
| **Область** | Набор шагов | Полноценный job |
| **Runners** | Использует runner родителя | Свой runner |
| **Secrets** | Не поддерживает напрямую | Полная поддержка |
| **Outputs** | Поддерживает | Поддерживает |
| **Matrix** | Не поддерживает | Поддерживает |
| **Когда использовать** | Для повторяющихся шагов | Для повторяющихся jobs |

---

## 📖 Основные файлы в проекте

```
.github/
├── actions/                          # Composite Actions
│   └── setup-go/                     # Action для настройки Go
│       └── action.yml
├── workflows/                        # Workflows
│   ├── ci-cd.yml                     # Основной CI/CD pipeline
│   ├── test-reusable.yml             # ♻️ Reusable: Тестирование
│   ├── docker-build-reusable.yml     # ♻️ Reusable: Docker сборка
│   ├── example-reusable-usage.yml    # 📖 Пример использования
│   └── README.md                     # Эта документация
```

---

## 🎓 Дополнительные ресурсы

- [GitHub Docs: Reusing workflows](https://docs.github.com/en/actions/using-workflows/reusing-workflows)
- [GitHub Docs: Composite actions](https://docs.github.com/en/actions/creating-actions/creating-a-composite-action)
- [Best practices for reusable workflows](https://docs.github.com/en/actions/using-workflows/reusing-workflows#best-practices-for-reusable-workflows)

---

## 💡 Best Practices

1. **Именование:** Используйте суффикс `-reusable.yml` для reusable workflows
2. **Документация:** Подробно описывайте inputs, outputs и secrets
3. **Версионирование:** Используйте теги/бранчи для версионирования workflows
4. **Минимальные права:** Указывайте минимально необходимые permissions
5. **Validation:** Проверяйте входные параметры внутри workflow
6. **Outputs:** Предоставляйте полезные outputs для последующих jobs
7. **Defaults:** Устанавливайте разумные значения по умолчанию
