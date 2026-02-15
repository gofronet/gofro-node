# gofro-node

Data-plane нода для open-source панели управления Xray. Управляет локальным процессом Xray Core, принимает команды по gRPC и синхронизирует конфигурацию.

```
Control Plane ── gRPC ──> gofro-node (data-plane) ──> Xray Core
         панели/ноды             управление ядром       проксирование трафика
```

## Возможности (MVP)

- запуск Xray Core при старте сервиса и управление жизненным циклом (start/stop/restart)
- прием конфигурации Xray по gRPC и запись в файл
- выдача текущей конфигурации и статуса ноды
- логирование gRPC запросов
- dev-режим с gRPC reflection

## Быстрый старт (локально)

```bash
export NODE_NAME="node-ru-1"
export XRAY_DEFAULT_CONFIG="xconf/config.json"
export XRAY_CORE_PATH="xray/xray"
export DEV_MODE="false"

go run ./cmd
```

gRPC слушает `:50051`.

## Установка на сервер (Linux + systemd)

Скрипт собирает бинарник, скачивает Xray и поднимает systemd-сервис.

```bash
sudo ./install.sh
```

Полезные переменные для `install.sh`:

- `APP_NAME` (по умолчанию `gofro-node`)
- `APP_USER` (по умолчанию `root`)
- `SERVICE_NAME` (по умолчанию `gofro-node`)
- `ENV_FILE` (по умолчанию `<repo>/.env`)
- `BIN_DIR` (по умолчанию `<repo>/bin`)
- `XRAY_DIR` (по умолчанию `<repo>/xray`)
- `XRAY_VERSION` (если нужен конкретный релиз Xray)

## Конфигурация окружения

- `NODE_NAME` (обязательный) — имя ноды в control plane
- `XRAY_DEFAULT_CONFIG` (по умолчанию `xconf/config.json`) — путь к базовому конфигу Xray
- `XRAY_CORE_PATH` (по умолчанию `xray/xray`) — путь к бинарнику Xray
- `DEV_MODE` (по умолчанию `false`) — включает gRPC reflection
- `XRAY_API_ADDRESS` (по умолчанию `127.0.0.1:8080`) — зарезервировано, сейчас не используется

## gRPC API (v1)

- `StartXray` — запуск Xray Core
- `StopXray` — остановка Xray Core
- `RestartXray` — перезапуск Xray Core
- `UpdateXrayConfig` — записывает новый конфиг в файл и обновляет его в менеджере (без рестарта)
- `GetNodeInfo` — статус Xray и имя ноды
- `GetCurrentConfig` — текущий конфиг в памяти

## Структура проекта

- `cmd/` — точка входа gRPC-сервера
- `internal/config/` — загрузка env и работа с конфигом Xray
- `internal/grpc_interceptors/` — gRPC interceptors
- `internal/xray_manager/` — управление процессом Xray и gRPC сервис
- `internal/xray_conn/` — gRPC клиент для Xray API (заготовка)
- `internal/gen/` — сгенерированные protobuf-структуры
- `xconf/` — базовый конфиг Xray
- `xray/` — бинарник Xray и базы geoip/geosite


## Roadmap

- регистрация ноды и аутентификация запросов control plane
- health/metrics endpoint (Prometheus)
- валидация и диффы конфигов перед применением
- удаленное обновление Xray Core и баз geoip/geosite
- безопасное хранение и ротация ключей
