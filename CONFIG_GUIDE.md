
# CONFIG_GUIDE.md

This document describes how to configure **Dead Letter Clerk**, a service that reads structured data from Redis and persists it to PostgreSQL. It is optimized for both human use and LLM prompting.

---

## Top-Level Structure

```yaml
redis:
  addr: ...
  password: ...
  db: ...

postgres:
  host: ...
  port: ...
  user: ...
  password: ...
  dbname: ...
  sslmode: ...

tasks:
  - name: ...
    ...
```

---

## Redis Configuration

| Key       | Type   | Required | Example            |
|-----------|--------|----------|--------------------|
| `addr`    | string | ✅        | `localhost:6379`    |
| `password`| string | ❌        | `""`                |
| `db`      | int    | ✅        | `0`                 |

---

## Postgres Configuration

| Key       | Type   | Required | Example       |
|-----------|--------|----------|---------------|
| `host`    | string | ✅        | `localhost`    |
| `port`    | int    | ✅        | `5432`         |
| `user`    | string | ✅        | `postgres`     |
| `password`| string | ✅        | `secret`       |
| `dbname`  | string | ✅        | `delivery`     |
| `sslmode` | string | ❌        | `disable`      |

---

## Task Configuration

Each task defines how to read from Redis and write into a Postgres table.

### Common Fields

| Key           | Type              | Required | Description                                    |
|----------------|-------------------|----------|------------------------------------------------|
| `name`         | string            | ✅        | Logical task name                              |
| `redis_key`    | string            | ✅        | Redis key to read from                         |
| `redis_type`   | string            | ✅        | One of: `list`, `stream`, `sorted_set`         |
| `db_table`     | string            | ✅        | Target Postgres table                          |
| `field_map`    | map[string]string | ✅        | Map from Redis field → Postgres column         |
| `schedule`     | string            | ✅        | Cron string or `@every` notation               |
| `checkpoint`   | object            | ❌        | Tracks which entries have already been written |

---

### Checkpoint

Optional tracking of Redis data to avoid duplication.

| Key             | Type   | Required | Description                                      |
|------------------|--------|----------|--------------------------------------------------|
| `enabled`        | bool   | ✅        | Whether to enable checkpointing                  |
| `field`          | string | ✅        | Field to track (e.g. stream ID or timestamp)     |
| `last_value_key` | string | ✅        | Redis key to store the last-seen value           |

---

## Examples

### Example 1: Read from a Redis list and insert into orders

```yaml
tasks:
  - name: sync_orders_from_list
    redis_key: new_orders
    redis_type: list
    db_table: public.orders
    field_map:
      id: id
      amount: amount
      created_at: created_at
    schedule: "@every 15s"
```

### Example 2: Read from a Redis stream and checkpoint by ID

```yaml
tasks:
  - name: sync_events_from_stream
    redis_key: stream:events
    redis_type: stream
    db_table: event_log
    field_map:
      event_type: type
      created_at: created_at
    schedule: "@every 10s"
    checkpoint:
      enabled: true
      field: __stream_id
      last_value_key: checkpoint:stream:events
```

### Example 3: Read from a sorted set and track by timestamp

```yaml
tasks:
  - name: sync_metrics
    redis_key: metrics:recent
    redis_type: sorted_set
    db_table: analytics.metrics
    field_map:
      timestamp: ts
      value: value
    schedule: "@every 30s"
    checkpoint:
      enabled: true
      field: timestamp
      last_value_key: checkpoint:metrics
```

---

## Prompt Template for LLMs

> You are configuring a Dead Letter Clerk task.  
> Output valid YAML where:  
> - Redis source is a stream called `audit:events`  
> - Destination table is `public.audit_log`  
> - You want to track by stream ID  
> - Schedule runs every 20 seconds

---

## Validation Rules

- `redis_type` must be one of `list`, `stream`, `sorted_set`
- `field_map` must map all required DB columns
- If `checkpoint.enabled` is true, `field` and `last_value_key` are required
- `schedule` must be a valid cron or `@every` value

---

## Testing

You can test the config by running:

```bash
go run ./cmd/clerk
```

and verifying that Redis records are inserted into Postgres tables on schedule.

---

## 🔗 Related

- [Main README](./README.md)
- [Dead Letter Clerk GitHub](https://github.com/Checker-Finance/dead-letter-clerk)
