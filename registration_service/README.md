# MICROSERVICE

## (name)

# Registration Service

This repository contains a microservice for handling institution registrations via RabbitMQ and PostgreSQL.

## Prerequisites

* Docker & Docker Compose (v2)
* Go 1.24 (for local development)
* (Optional) `psql` CLI if you want manual DB dumps

## Project Structure

```text
registration_service/
├── Dockerfile            # Builds the Go service
├── docker-compose.yml    # Defines Postgres, RabbitMQ, and the service
├── db-init/              # SQL dump to initialize the database
│   └── init.sql          # Your schema + seed data
├── .env                  # Environment variables
├── go.mod, go.sum        # Go module files
├── main.go               # Entry point for the service
└── handlers/             # Message handler logic
    └── handler.go
└── dbService/            # Database initialization and queries
    └── db.go
```

## 1. Prepare the Database Initialization

1. Place your SQL dump in `db-init/init.sql`. This file should include all DDL and seed data.
2. Confirm your `docker-compose.yml` for the Postgres service:

   ```yaml
   services:
     postgres:
       image: postgres:14
       container_name: postgres
       restart: always
       environment:
         POSTGRES_USER: ${DB_USER}
         POSTGRES_PASSWORD: ${DB_PASS}
         POSTGRES_DB:   ${DB_NAME}
       volumes:
         - ./db-init:/docker-entrypoint-initdb.d:ro
         - pgdata:/var/lib/postgresql/data
       ports:
         - "5433:5432"  # host:container
   volumes:
     pgdata:
   ```
3. Ensure the dump covers the `institution` table and any sequences.

## 2. Configure Environment Variables

Copy or create a `.env` in the project root:

```ini
# .env
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASS=
DB_NAME=reg_service
DB_PASSWORD=

RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
RABBIT_USER=guest
RABBIT_PASS=guest

# (Optional) HTTP port your service listens on
PORT=8080
```

> **Note:** Inside Docker Compose, `DB_HOST` and RabbitMQ host names refer to service names (`postgres` and `rabbitmq`).

## 3. Build & Run with Docker Compose

Bring down any existing stack (removes containers & volumes):

```bash
docker compose down -v --remove-orphans
```

Build and start all services:

```bash
docker compose up -d --build
```

Verify all containers are running:

```bash
docker compose ps
```

You should see:

* `postgres` (Up)
* `rabbitmq` (Up)
* `service` (Up) — your Go application

## 4. Verify Database

Connect to Postgres and list tables:

```bash
docker compose exec postgres psql -U $DB_USER -d $DB_NAME -c "\dt"
```

You should see the `institution` table and any seed data rows.

## 5. Testing the Service

### 5.1 Create a Reply Queue

```bash
docker exec -it rabbitmq rabbitmqadmin declare queue name=registration.reply durable=true \
  --username $RABBIT_USER --password $RABBIT_PASS
```

### 5.2 Publish a Test Message

```bash
docker exec -it rabbitmq rabbitmqadmin publish \
  exchange=clearSky.events routing_key=institution.registered \
  payload='{"name":"Test Inst","email":"test@inst.com","director":"Alice"}' \
  properties='{"reply_to":"registration.reply","correlation_id":"reg-001"}' \
  --username $RABBIT_USER --password $RABBIT_PASS
```

### 5.3 Retrieve the Reply

```bash
docker exec -it rabbitmq rabbitmqadmin get \
  queue=registration.reply ackmode=ack_requeue_false count=1 \
  --username $RABBIT_USER --password $RABBIT_PASS
```

Expected response:

```json
{"status":"ok","message":"Institution registered successfully"}
```

## 6. Local Development (without Docker)

1. Install dependencies:

   ```bash
   ```

go mod tidy

````
2. Export env vars or source `.env`:
```bash
export $(grep -v '^#' .env | xargs)
````

3. Run the service:

   ```bash
   ```

go run main.go

```

Ensure Postgres and RabbitMQ are running locally and reachable via the URLs in your `.env`.

## Troubleshooting

- **`relation "institution" does not exist`**: The init script didn't run. Ensure you removed the volume with `docker compose down -v` before `up`.
- **Connection refused to Postgres**: Check that `DB_HOST` is set to `postgres` in the container context and ports are correctly mapped.
- **Service keeps restarting**: Run `docker compose logs -f service` to inspect errors, usually due to misconfigured env vars or DNS.

---

With these steps you can run, test, and develop the registration service end-to-end in Docker. Happy coding!

```
