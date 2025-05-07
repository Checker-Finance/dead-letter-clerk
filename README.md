[![License](https://img.shields.io/github/license/Checker-Finance/dead-letter-clerk)](LICENSE)
![Go Version](https://img.shields.io/badge/go-1.24-blue)
<p align="center">
  <img src="assets/logo.png" alt="Red Courier Logo" width="200" height="200">
</p>

# Dead Letter Clerk

**Dead Letter Clerk** is a scheduled data synchronizer that reads structured data from Redis (streams, lists, sorted sets) and persists it into PostgreSQL tables.

It is designed to be the reverse of [Red Courier](https://github.com/Checker-Finance/red-courier): instead of pushing from Postgres into Redis, Dead Letter Clerk listens for real-time or delayed messages in Redis and stores them durably in a relational database.

---

## Features

- Supports reading from Redis:
    - **Lists** (`LRANGE` + `LTRIM`)
    - **Sorted Sets** (`ZRANGEBYSCORE`) with checkpointing
    - **Streams** (`XREAD`) with last ID tracking
- Writes to Postgres using batch `INSERT INTO` statements
- Configurable per-task field mappings and schedules
- Cron-style or `@every` interval task scheduling
- YAML-based configuration with support for multiple tasks
- Simple, pluggable architecture with Redis + PG client wrappers

---

## Example Use Cases

- Persist ephemeral Redis queues (e.g. `new_orders`) to `orders` table
- Stream structured metrics from Redis sorted sets to time-series DB
- Store high-throughput events from Redis Streams to an audit log

---

## Configuration

Place your config in `config.yaml`:

```yaml
redis:
  addr: localhost:6379
  db: 0

postgres:
  host: localhost
  port: 5432
  user: postgres
  password: secret
  dbname: delivery
  sslmode: disable

tasks:
  - name: sync_orders_from_list
    redis_key: new_orders
    redis_type: list
    db_table: public.orders
    field_map:
      id: id
      amount: amount
    schedule: "@every 15s"

  - name: sync_events_from_stream
    redis_key: stream:events
    redis_type: stream
    db_table: public.event_log
    field_map:
      event_type: type
      created_at: created_at
    schedule: "@every 10s"
    checkpoint:
      enabled: true
      field: __stream_id
      last_value_key: checkpoint:stream:events