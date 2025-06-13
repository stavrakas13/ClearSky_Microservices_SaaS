# Stats Service

Stores exam and grade data and calculates grade distributions.
It consumes messages from RabbitMQ and replies with requested statistics.

## Environment variables
- `DB_HOST`, `DB_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` – PostgreSQL settings
- `DB_DSN` – optional full connection string used by docker-compose

- `RABBITMQ_URL` – AMQP URI for RabbitMQ

## Running with Docker Compose

```bash
# inside stats_service/
docker compose up --build
```

This starts PostgreSQL, RabbitMQ and the Go service.

## Message API

The service listens on the `clearSky.events` exchange using queue `stats_queue`.
Routing keys:
- `stats.persist_and_calculate`
- `stats.get_distributions`

Replies (for `get_distributions`) are sent to the sender provided `reply_to` queue
with the same `correlation_id`.

Example payloads:

```json
// stats.persist_and_calculate
{
  "exam": {"class_id":"X123","exam_date":"2025-06-01"},
  "grades": [{"student_id":"42","question_scores":[10,9,8]}]
}

// stats.get_distributions
{
  "class_id": "X123",
  "exam_date": "2025-06-01"
}
```

