# User Management Microservice

This microservice handles user authentication and authorization for the clearSKY application, part of the SaaS Technologies course (NTUA, Spring 2024–2025).

## Technologies Used

- GoLang 1.22
- RabbitMQ for asynchronous messaging
- SQLite with GORM ORM
- Docker and Docker Compose
- JWT for stateless authentication

## Supported Features

### REST API Endpoints

| Endpoint         | Method | Description                  |
|------------------|--------|------------------------------|
| `/register`      | POST   | Register a new user          |
| `/login`         | POST   | Authenticate user and issue a JWT |
| `/auth/validate` | GET    | Validate a JWT via middleware |

### RabbitMQ Messaging

- **Queue:** `auth.request`
- **Exchange:** `orchestrator.commands` (topic)
- **Bindings:** `auth.register`, `auth.login`
- **Reply Queue:** Defined by the `reply_to` field in the request
- **Correlation ID:** Supported for message tracking

#### Sample Request (RabbitMQ)

```json
{
  "type": "login",
  "email": "student@example.com",
  "password": "mypassword123"
}
```

#### Sample Response

```json
{
  "status": "ok",
  "token": "<jwt_token_here>",
  "role": "student"
}
```

## Orchestrator Integration

- The service listens on the `auth.request` queue for messages with routing keys `auth.register` and `auth.login` from the `orchestrator.commands` exchange.
- It replies to the `reply_to` queue with the same `correlation_id` for RPC-style communication with the orchestrator.
- Make sure the orchestrator and this service use the same RabbitMQ instance and exchange/queue names.

## Execution Instructions (Dockerized)

1. Ensure `docker` and `docker-compose` are installed.
2. Build and start the services:

```bash
docker-compose up --build
```

3. Access Points:
   - API Service: http://localhost:8082
   - RabbitMQ Management UI: http://localhost:15672
     - Username: `guest`
     - Password: `guest`

## Project Structure

```
user_management_service/
├── cmd/                    # Application entry point
├── internal/
│   ├── config/             # Database configuration
│   ├── handler/            # HTTP handlers
│   ├── messaging/          # RabbitMQ consumer and producer logic
│   ├── middleware/         # JWT validation logic
│   └── model/              # GORM models
├── pkg/jwt/                # JWT utility functions
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## Environment

- The project uses a multi-stage Docker build for optimized container size.
- JWT tokens expire after 24 hours and are signed using HS256.
- SQLite database is stored locally in `auth_service.db`.

## Authors

- clearSKY Project Team [Group 12]
- National Technical University of Athens (NTUA)
- Course: Software as a Service Technologies (2024–2025)

## Implementation Status

| Feature                      | Status |
|-----------------------------|--------|
| User registration with role | Done   |
| JWT-based login             | Done   |
| Token validation            | Done   |
| RabbitMQ message handling   | Done   |
| Docker support              | Done   |
| Role included in JWT        | Done   |
