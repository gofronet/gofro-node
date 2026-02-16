
# gofro-node

`gofro-node` — data-plane агент платформы Gofro. Работает на сервере ноды, управляет локальным Xray Core и исполняет команды от `gofro-control`.

## Роль в системе

- исполняющий слой для control-plane;
- управление процессом Xray на локальной машине;
- хранение и применение конфигурации Xray.

## Взаимодействие с другими сервисами

- принимает gRPC-команды от `gofro-control`;
- не предоставляет публичный HTTP API для панели;
- напрямую взаимодействует с локальным бинарником Xray.

## Функциональность (MVP)

- gRPC сервер (по умолчанию `:50051`);
- команды `start/stop/restart` для Xray;
- обновление и чтение текущего конфига;
- выдача статуса ноды;
- gRPC reflection в `DEV_MODE=true`.

## Быстрый запуск

```bash
export NODE_NAME="node-ru-1"
export XRAY_DEFAULT_CONFIG="xconf/config.json"
export XRAY_CORE_PATH="xray/xray"
export DEV_MODE="false"

go run ./cmd
```

## Установка на сервер (Linux + systemd)

```bash
sudo ./install.sh
```

Переменные `install.sh`:

- `APP_NAME` (`gofro-node`)
- `APP_USER` (`root`)
- `SERVICE_NAME` (`gofro-node`)
- `ENV_FILE` (`<repo>/.env`)
- `BIN_DIR` (`<repo>/bin`)
- `XRAY_DIR` (`<repo>/xray`)
- `XRAY_VERSION` (опционально, конкретный релиз)

## Конфигурация окружения

- `NODE_NAME` (обязательная)
- `XRAY_DEFAULT_CONFIG` (по умолчанию `xconf/config.json`)
- `XRAY_CORE_PATH` (по умолчанию `xray/xray`)
- `DEV_MODE` (по умолчанию `false`)
- `XRAY_API_ADDRESS` (зарезервировано, сейчас не используется)

## gRPC API (v1)

- `StartXray`
- `StopXray`
- `RestartXray`
- `UpdateXrayConfig`
- `GetNodeInfo`
- `GetCurrentConfig`

## Интеграция с gofro-control

1. Поднимите `gofro-node` и откройте gRPC порт.
2. В `gofro-control` выполните `POST /v1/nodes/` с адресом ноды.
3. Все дальнейшие операции выполняйте через control-plane.

## Структура

- `cmd/` — запуск gRPC сервера
- `internal/config/` — env и работа с конфигом
- `internal/grpc_interceptors/` — interceptors
- `internal/xray_manager/` — менеджер Xray и gRPC сервис
- `internal/xray_conn/` — заготовка клиента к Xray API
- `internal/gen/` — protobuf-артефакты
- `xconf/` — дефолтный конфиг Xray
- `xray/` — бинарник Xray и базы geo

## Ограничения

- нет встроенной регистрации ноды и доверенной identity-модели;
- нет метрик/health endpoint по умолчанию;
- нет безопасного канала по умолчанию между компонентами.
