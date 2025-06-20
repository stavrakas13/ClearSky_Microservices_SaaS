# Google Auth Service

Provides Google OAuth integration: redirects users to Google, handles callback, issues JWT stored in a cookie, and publishes login events.

## Endpoints

- GET /auth/google/login  
- GET /auth/google/callback  
- GET /auth/logout  
- GET /protected  (requires valid JWT in cookie)

## RabbitMQ Integration

- **Consumer:** listens on queue `google_auth.request` bound to `orchestrator.commands` with routing key `auth.login.google`.  
- **Publisher:** emits `user_logged_in` events to exchange `clearsky.events` (fanout).