# gofro-node

Data-plane нода для сервиса GoFroNet VPN. Управляет ядром Xray на машине, принимает команды от control plane по gRPC и синхронизирует конфигурацию.

```
Control Plane ── gRPC ──> gofro-node (data-plane) ──> Xray Core
         подписки, ноды          управление ядром       проксирование трафика
```

## Что делает сейчас (MVP)

- поднимает Xray Core и управляет его жизненным циклом (start/stop/restart)
- принимает конфиг Xray по gRPC, сохраняет его в файл и применяет через рестарт
- отдает текущую конфигурацию и состояние ноды
- логирует gRPC запросы
- поддерживает dev-режим с reflection

## Быстрый старт

```bash
export NODE_NAME="node-ru-1"
export XRAY_DEFAULT_CONFIG="xconf/config.json"
export XRAY_CORE_PATH="xray/xray"
export DEV_MODE="false"

go run ./cmd
```

По умолчанию gRPC слушает `:50051`.

## Конфигурация окружения

- `NODE_NAME` (обязательный) — имя ноды в control plane
- `XRAY_DEFAULT_CONFIG` (по умолчанию `xconf/config.json`) — путь к базовому конфигу Xray
- `XRAY_CORE_PATH` (по умолчанию `xray/xray`) — путь к бинарнику Xray
- `DEV_MODE` (по умолчанию `false`) — включает gRPC reflection

## gRPC API (v1)

- `StartXray` — запуск Xray Core
- `StopXray` — остановка Xray Core
- `RestartXray` — перезапуск Xray Core
- `UpdateXrayConfig` — обновить конфиг, записать в файл и перезапустить
- `GetNodeInfo` — статус Xray и имя ноды
- `GetCurrentConfig` — текущий конфиг в памяти

## Структура проекта

- `cmd/` — точка входа gRPC-сервера
- `delivery/` — gRPC handlers и interceptors
- `xray_manager/` — управление процессом Xray
- `config/` — загрузка env и работа с конфигом Xray
- `xconf/` — базовый конфиг Xray
- `xray/` — бинарник Xray и базы geoip/geosite
- `gen/` — сгенерированные protobuf-структуры

## Разработка

```bash
go test ./...
```

Тесты используют локальный бинарник `xray/xray`.

## Roadmap

- регистрация ноды и аутентификация запросов control plane
- health/metrics endpoints (Prometheus)
- валидация и диффы конфигов перед применением
- удаленное обновление Xray Core и датабаз geoip/geosite
- безопасное хранение и ротация ключей
