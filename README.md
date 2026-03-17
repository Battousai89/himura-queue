# Himura Queue

Высокопроизводительный сервис очередей на Go.

## Особенности

- **Бинарный TCP-протокол** — фреймы с заголовком (длина + тип команды)
- **Приоритетные очереди** — на основе `container/heap`
- **Отложенные сообщения** — с таймерами доставки
- **Exactly-once гарантии** — дедупликация с TTL
- **Шардирование** — consistent hashing по имени очереди
- **Динамический пул воркеров** — автомасштабирование под нагрузку
- **Периодические снапшоты** — восстановление состояния после перезагрузки
- **HTTP health-check** — мониторинг доступности

## Структура проекта

```
himura-queue/
├── cmd/
│   ├── server/         # TCP сервер
│   └── cli/            # CLI утилита
├── internal/
│   ├── config/         # TOML парсер конфигурации
│   ├── protocol/       # Бинарный протокол (кодирование/декодирование)
│   ├── queue/          # Ядро очереди (PriorityQueue, DelayedQueue, Shard, Manager)
│   ├── hashing/        # Consistent hashing
│   ├── deduplication/  # Exactly-once дедупликация
│   ├── worker/         # Динамический пул воркеров
│   ├── persistence/    # Снапшоты состояния
│   └── server/         # TCP сервер + HTTP health endpoint
├── pkg/models/         # Общие модели данных
├── config.toml         # Конфигурационный файл
└── go.mod
```

## Установка

```bash
git clone <repository>
cd himura-queue
go build -o himura-server ./cmd/server
go build -o himura-cli ./cmd/cli
```

## Конфигурация

Файл `config.toml`:

```toml
[server]
tcp_port = 9000      # Порт TCP сервера
http_port = 9001     # Порт HTTP health-check

[queue]
shard_count = 8      # Количество шардов

[worker]
min_workers = 4      # Минимальное число воркеров
max_workers = 100    # Максимальное число воркеров
idle_timeout_sec = 30 # Таймаут простоя воркера

[snapshot]
path = "data/snapshot.bin"  # Путь к файлу снапшота
interval_sec = 30           # Интервал создания снапшотов
```

Запуск с конфигом:
```bash
./himura-server -config config.toml
```

Запуск с флагами (переопределяют конфиг):
```bash
./himura-server -config config.toml -tcp-port 9000
```

## CLI утилита

### Push — отправка сообщения

```bash
# Базовое использование
./himura-cli push --queue myqueue --payload "hello world"

# С приоритетом (чем выше число, тем выше приоритет)
./himura-cli push --queue myqueue --payload "urgent" --priority 100

# С отложенной доставкой (5 секунд)
./himura-cli push --queue myqueue --payload "delayed" --delay 5s

# Полный набор опций
./himura-cli push \
    --host localhost \
    --port 9000 \
    --queue myqueue \
    --payload "message body" \
    --priority 50 \
    --delay 10s
```

### Pop — получение сообщения

```bash
# Базовое использование
./himura-cli pop --queue myqueue

# С указанием сервера
./himura-cli pop --host localhost --port 9000 --queue myqueue
```

Возвращает сообщение с наивысшим приоритетом. Если очередь пуста — выводит "No messages".

### Stats — статистика очереди

```bash
# Длина очереди
./himura-cli stats --queue myqueue

# С указанием сервера
./himura-cli stats --host localhost --port 9000 --queue myqueue
```

### Health — проверка доступности

```bash
# Проверка HTTP health endpoint
./himura-cli health

# С указанием хоста и порта
./himura-cli health --host localhost --http-port 9001
```

## Протокол

### Бинарный формат фрейма

```
+----------------+----------------+------------------+
|  Длина (4 байта) |  Тип (1 байт)  |  Данные (N байт)  |
+----------------+----------------+------------------+
```

Все числовые поля в big-endian.

### Типы команд

| Команда | Код | Описание |
|---------|-----|----------|
| PUSH | 1 | Добавить сообщение |
| POP | 2 | Получить сообщение |
| ACK | 3 | Подтвердить обработку |
| STATUS | 5 | Получить длину очереди |

### Формат запросов

**PUSH Request:**
```
+------------+--------------+------------+-------------+------------+
| Queue Len  | Queue        | Payload Len| Payload     | Priority   |
| (2 байта)  | (N байт)     | (4 байта)  | (M байт)    | (4 байта)  |
+------------+--------------+------------+-------------+------------+
| Delay (8 байт) |
+------------------+
```

**POP Request:**
```
+------------+--------------+
| Queue Len  | Queue        |
| (2 байта)  | (N байт)     |
+------------+--------------+
```

**ACK Request:**
```
+------------------+
| Message ID (8 байт) |
+------------------+
```

**STATUS Request:**
```
+------------+--------------+
| Queue Len  | Queue        |
| (2 байта)  | (N байт)     |
+------------+--------------+
```

## API

### TCP команды

```go
// Отправка сообщения
PUSH {queue: string, payload: []byte, priority: int, delay: duration}
→ {id: uint64}

// Получение сообщения
POP {queue: string}
→ {id: uint64, payload: []byte} | empty

// Подтверждение обработки
ACK {id: uint64}
→ {acked: bool}

// Статистика
STATUS {queue: string}
→ {length: uint64}
```

### HTTP эндпоинты

| Endpoint | Method | Описание |
|----------|--------|----------|
| `/health` | GET | Проверка доступности (200 OK) |

## Тестирование

```bash
# Unit тесты
go test ./... -short

# Бенчмарки
go test ./internal/queue -bench=. -benchtime=1s
go test ./internal/protocol -bench=. -benchtime=1s
go test ./internal/hashing -bench=. -benchtime=1s
```

### Результаты бенчмарков

```
BenchmarkPriorityQueuePush-12    1272819    84.70 ns/op
BenchmarkPriorityQueuePop-12     3823129    32.37 ns/op
BenchmarkManagerPush-12           727537   231.5 ns/op
BenchmarkManagerPop-12            749905   137.1 ns/op
BenchmarkDelayedQueue-12         1000000   102.3 ns/op
BenchmarkConsistentHash-12       3365708    33.30 ns/op
```

## Гарантии доставки

**Exactly-once** реализуется через:

1. Уникальный ID каждому сообщению (монотонный счётчик)
2. Дедупликация входящих сообщений (in-memory map с TTL)
3. ACK подтверждение обработки
4. Повторные сообщения с тем же ID игнорируются

## Персистентность

- Периодический снапшот всех очередей в бинарный файл
- Восстановление состояния при старте сервера
- Интервал снапшотов настраивается в конфиге

## Примеры использования

### 1. Запуск сервера

```bash
./himura-server -config config.toml
```

### 2. Отправка сообщений

```bash
# Обычное сообщение
./himura-cli push --queue orders --payload '{"order_id": 123}'

# Срочное сообщение (высокий приоритет)
./himura-cli push --queue orders --payload '{"order_id": 999, "vip": true}' --priority 100

# Отложенное уведомление
./himura-cli push --queue notifications --payload "reminder" --delay 1h
```

### 3. Обработка очереди

```bash
# Получение следующего сообщения
./himura-cli pop --queue orders

# Проверка длины очереди
./himura-cli stats --queue orders
```

### 4. Мониторинг

```bash
# Проверка здоровья
./himura-cli health
curl http://localhost:9001/health
```
