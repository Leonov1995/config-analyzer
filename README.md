# Config Analyzer

Утилита на Go для анализа конфигурационных файлов YAML/JSON и выявления потенциально опасных настроек безопасности.

## Возможности

- 🔍 **6 правил безопасности** — TLS, надёжность пароля, алгоритмы хеширования, привязка сервера, уровень логирования, права доступа к файлам
- 📁 **Рекурсивное сканирование директорий** — анализ всех конфигурационных файлов в директории параллельно
- 🔒 **Проверка прав доступа** — обнаружение файлов, доступных для записи всем пользователям
- 🖥️ **Три режима в одном бинарнике** — CLI, REST API, gRPC
- ⚡ **Ограниченная конкурентность** — пул воркеров с семафором, без утечек горутин
- 📊 **Вывод по уровням серьёзности** — HIGH / MEDIUM / LOW с рекомендациями по исправлению

## Быстрый старт

```bash
# Сборка
make build
```

## Использование CLI

### Тестовые данные

Для ручного тестирования используйте файлы из `configs/`:

### Анализ файлов

```bash
# Анализ одного файла
./bin/config-analyzer analyze configs/bad_config.json

# Анализ всей директории
./bin/config-analyzer analyze configs/

# Тихий режим — показывает проблемы, но код выхода 0
./bin/config-analyzer analyze --silent configs/bad_config.json
./bin/config-analyzer analyze -s configs/bad_config.yaml

# Stdin с явным указанием формата
printf '{"server":{"port":8080}}' | ./bin/config-analyzer analyze --stdin --format json

# Stdin с автоопределением формата
cat configs/good_config.yaml | ./bin/config-analyzer analyze --stdin
```

**Флаги:**

| Флаг       | Краткий | Описание                                                              |
| ---------- | ------- | --------------------------------------------------------------------- |
| `--silent` | `-s`    | Не завершать с ошибкой, если найдены проблемы                         |
| `--stdin`  |         | Читать конфигурацию из stdin вместо файла                             |
| `--format` | `-f`    | Формат входных данных: `json` или `yaml` (определяется автоматически) |

### REST-сервер

```bash
# Запуск сервера (один из вариантов)
./bin/config-analyzer serve-http
./bin/config-analyzer serve-http --addr :8080

# «Плохая» конфигурация — ожидайте HIGH-проблемы
curl -s -X POST http://localhost:8080/analyze \
  -H "Content-Type: application/json" \
  -d @configs/bad_config.json | jq -C

# «Хорошая» конфигурация — проблем нет
curl -s -X POST http://localhost:8080/analyze \
  -H "Content-Type: application/json" \
  -d @configs/good_config.json | jq -C

# YAML через Content-Type
curl -s -X POST http://localhost:8080/analyze \
  -H "Content-Type: application/yaml" \
  --data-binary @configs/bad_config.yaml | jq -C

# Частичная конфигурация
curl -s -X POST http://localhost:8080/analyze \
  -H "Content-Type: application/json" \
  -d @configs/partial_config.json | jq -C

# Инлайн JSON
curl -s -X POST http://localhost:8080/analyze \
  -H "Content-Type: application/json" \
  -d '{"server":{"host":"0.0.0.0","port":8080},"tls":{"enabled":false},"auth":{"password":"123"}}' | jq -C
```

Все примеры REST-запросов — в `configs/rest_reqs.txt`.

### gRPC-сервер

```bash
# Запуск сервера (один из вариантов)
./bin/config-analyzer serve-grpc
./bin/config-analyzer serve-grpc --addr :50051

# «Чистая» конфигурация
grpcurl -plaintext \
  -d '{
    "config": {
      "server": { "host": "1.0.0.0", "port": 80 },
      "auth":   { "password": "StrongPassword123!" },
      "log":    { "output": "stdout", "level": "info" },
      "tls":    { "enabled": true },
      "storage": { "digestAlgorithm": "sha256" }
    }
  }' \
  localhost:50051 analyzer.AnalyzerService/Analyze

# «Проблемная» конфигурация
grpcurl -plaintext \
  -d '{
    "config": {
      "server": { "host": "0.0.0.0", "port": 80 },
      "auth":   { "password": "123" },
      "log":    { "output": "stdout", "level": "debug" },
      "tls":    { "enabled": false },
      "storage": { "digestAlgorithm": "md5" }
    }
  }' \
  localhost:50051 analyzer.AnalyzerService/Analyze

# Минимальный запрос
grpcurl -plaintext \
  -d '{"config":{"server":{"host":"0.0.0.0","port":3000}}}' \
  localhost:50051 analyzer.AnalyzerService/Analyze
```

Все примеры gRPC-запросов — в `configs/grpc_reqs.txt`.

## Правила безопасности

| Правило               | Серьёзность | Срабатывает, когда…                                  |
| --------------------- | ----------- | ---------------------------------------------------- |
| Проверка TLS          | **HIGH**    | `tls.enabled` = `false`                              |
| Надёжность пароля     | **HIGH**    | Пароль задан, но < 8 символов или равен `"password"` |
| Алгоритм хеширования  | **HIGH**    | `storage.digest-algorithm` = `md5` или `sha1`        |
| Права доступа к файлу | **HIGH**    | Файл конфигурации доступен для записи всем (`o+w`)   |
| Привязка сервера      | MEDIUM      | `server.host` пустой или `0.0.0.0`                   |
| Уровень логирования   | LOW         | `log.level` = `debug`                                |

## Примеры конфигураций

Все готовые файлы находятся в директории `configs/`:

| Файл                  | Формат | Описание                               |
| --------------------- | ------ | -------------------------------------- |
| `bad_config.json`     | JSON   | Конфигурация со всеми типами проблем   |
| `bad_config.yaml`     | YAML   | То же, в YAML                          |
| `good_config.json`    | JSON   | Безопасная конфигурация без проблем    |
| `good_config.yaml`    | YAML   | То же, в YAML                          |
| `good_config.yml`     | YML    | То же, в .yml                          |
| `partial_config.json` | JSON   | Частичная конфигурация (только сервер) |
| `partial_config.yaml` | YAML   | То же, в YAML                          |
| `empty.json`          | JSON   | Пустой объект `{}`                     |
| `empty.yaml`          | YAML   | Пустой YAML-документ                   |
| `rest_reqs.txt`       | —      | Готовые curl-команды для REST          |
| `grpc_reqs.txt`       | —      | Готовые grpcurl-команды для gRPC       |

### Все тесты

```bash
# Через make
make test

# Напрямую
go test ./... -v -count=1
```

### Отдельные пакеты

```bash
# Сервисный слой (22 теста)
go test ./internal/service/ -v

# Правила безопасности
go test ./internal/rule/ -v

# Загрузчик конфигураций
go test ./internal/config/ -v

# HTTP-обработчик
go test ./internal/server/http/ -v

# gRPC-сервер
go test ./internal/server/grpc/ -v

# Анализатор
go test ./internal/analyzer/ -v
```

### Покрытие

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Архитектура

```
cmd/                              # Точка входа — один бинарник
├── internal/
│   ├── model/                    # Доменные модели (Config, Issue, Report, Summary)
│   ├── rule/                     # Интерфейс Rule + 6 правил безопасности
│   ├── analyzer/                 # Оркестратор правил
│   ├── config/                   # Загрузчик JSON/YAML (без Viper)
│   ├── service/                  # Бизнес-логика: параллельный анализ, сбор файлов
│   ├── server/                   # Общие утилиты: graceful shutdown, сигналы
│   │   ├── http/                 # REST-обработчик + сервер
│   │   └── grpc/                 # gRPC-сервер
│   │       └── proto/            # Protobuf-определения
│   └── cli/                      # Cobra CLI команды
├── configs/                      # Примеры конфигураций и запросов
├── Makefile
└── README.md
```

## Разработка

```bash
# Установка зависимостей
go mod tidy

# Тесты
make test

# Линтинг
make vet

# Перегенерация protobuf (требуется protoc)
make proto
```

## Добавление новых правил

1. Создайте файл `internal/rule/<имя>.go`:

```go
package rule

import "config-analyzer/internal/model"

type MyRule struct{}

func (r MyRule) Check(cfg *model.Config) *model.Issue {
    if /* условие */ {
        return &model.Issue{
            Severity:       model.HIGH,
            Message:        "описание проблемы",
            Recommendation: "рекомендация по исправлению",
            Path:           "config.path",
        }
    }
    return nil
}
```

2. Зарегистрируйте правило в `internal/rule/rule.go`:

```go
func NewDefaultRules() []Rule {
    return []Rule{
        // ... существующие правила
        MyRule{},
    }
}
```

Анализатор подхватит правило автоматически.
