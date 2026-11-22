# Hello World Cloud-Native Application - Design

## 1. Purpose

This document converts the requirements into a concrete system design for a Dockerized React and Go application. The initial implementation uses MongoDB and Redis, exposes Prometheus metrics, and can run locally with Docker Compose or on GCP.

## 2. Goals and Scope

### In scope

- React UI for loading and displaying a message.
- Gin REST API with health and metrics endpoints.
- MongoDB-backed message storage.
- Redis cache-aside behavior.
- Structured logging and Prometheus metrics.
- Docker Compose for all local dependencies.
- GCP deployment architecture.
- Unit, integration, load, and end-to-end validation.

### Out of scope for the initial release

- PostgreSQL implementation.
- Authentication and authorization.
- Public TLS certificate management.
- OpenTelemetry tracing.
- CI/CD and Kubernetes deployment.

PostgreSQL is included as a future repository interface implementation.

## 3. System Architecture

```text
Browser
  |
  | GET /api/message?... /api/health
  v
React UI (Nginx)
  |
  | reverse proxy /api -> /
  v
Go API (Gin)
  |
  +--> Redis (cache-aside)
  |
  +--> MongoDB (message storage)
  |
  +--> Prometheus /metrics

Prometheus --> Grafana
```

The frontend should call `/api/message`; Nginx removes the `/api` prefix before proxying to the backend. The backend itself exposes `/message`, `/health`, and `/metrics`.

## 4. Component Design

### 4.1 Frontend

- React application built and served by Nginx.
- Calls `GET /api/message?fromdb=1&cache=1&sql=0`.
- Displays `response.message`.
- Displays `Unable to load message.` for network or non-success responses.
- Shows a loading state while the request is pending.
- Uses environment configuration for the API base URL.
- Nginx serves the React bundle and proxies `/api/*` to the backend service.

### 4.2 Backend

Use a package layout that separates HTTP transport from application and infrastructure concerns:

```text
backend/
  cmd/server/main.go
  internal/
    config/
    handler/
    service/
    repository/
      message_repository.go
      mongo_message_repository.go
      postgres_message_repository.go
    cache/
      message_cache.go
      redis_message_cache.go
    metrics/
  Dockerfile
  go.mod
```

Recommended interfaces:

```go
type MessageRepository interface {
    GetMessage(ctx context.Context) (string, error)
}

type MessageCache interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value string) error
}

type MessageService interface {
    GetMessage(ctx context.Context, options MessageOptions) (MessageResult, error)
}
```

No cache key should include request booleans. Use a stable key such as `hello-world:message:1`.

### 4.3 MongoDB

Store one message document:

```json
{
  "_id": "1",
  "message": "Hello World"
}
```

Use a unique `_id` and a repository method that reads this document:

```text
db.messages.findOne({ _id: "1" })
```

Initialize the collection and seed document through a migration or idempotent startup seed script.

### 4.4 Redis

Implement cache-aside behavior:

1. When cache is enabled, read `hello-world:message:1`.
2. Return the cached message with `source: "cache"` on a hit.
3. On a miss, read MongoDB.
4. Cache the database value with a configurable TTL, for example 300 seconds.
5. Return the message with `source: "mongodb"` on a database hit.

Redis failures must not fail the API request. Log and count the Redis error, then continue to MongoDB. Cache writes are best-effort.

## 5. API Contract

### 5.1 GET /message

Request:

```http
GET /message?fromdb=1&cache=1&sql=0
```

Parameters:

| Parameter | Type | Default | Rules |
|-----------|------|---------|-------|
| `fromdb` | 0 or 1 | 1 | Must be `1` in the initial implementation |
| `cache` | 0 or 1 | 1 | `1` enables Redis cache-aside; `0` bypasses Redis |
| `sql` | 0 or 1 | 0 | `0` selects MongoDB; `1` is reserved and initially returns 501 |

Success:

```json
{
  "message": "Hello World",
  "source": "mongodb"
}
```

Other valid source values are `cache` and, after PostgreSQL implementation, `postgresql`.

The API should define the unsupported source and parameter behavior explicitly:

- Invalid parameter values or missing required values: `400 Invalid Request`.
- PostgreSQL request (`sql=1`): `501 Not Implemented`.
- Database failure or unexpected internal failure: `500 Internal Error`.
- If Redis fails but MongoDB succeeds, the request remains `200` and the source is `mongodb`.

## 6. Health Endpoints

### 6.1 GET /health

Return 200 only when the API process is healthy:

```json
{
  "status": "UP"
}
```

Use separate readiness behavior if Kubernetes is added later. In Docker Compose, health checks should ensure that a running API container also has reachable Redis and MongoDB dependencies.

Health and metrics endpoints should not require authentication.

## 7. Metrics

Expose a Prometheus-compatible `/metrics` endpoint.

Required metrics:

- `http_requests_total{method,status,endpoint}`
- `http_request_duration_seconds{method,endpoint}`
- `cache_hits_total`
- `cache_misses_total`
- `cache_errors_total`
- `db_reads_total{database}`
- `db_errors_total{database}`

Use a Gin middleware to instrument every HTTP request. Histograms should use latency buckets suitable for the 100 ms cache and 500 ms database targets.

## 8. Logging

Use JSON structured logs to a single stdout/stderr stream.

Example:

```json
{
  "timestamp": "2026-11-22T17:08:02Z",
  "level": "info",
  "request_id": "123",
  "endpoint": "/message",
  "duration_ms": 10,
  "status": 200,
  "source": "cache"
}
```

Generate a request ID for every request. Reuse `X-Request-ID` when supplied or generate a UUID. Include the request ID in the response headers and logs. Never log credentials or connection strings.

## 9. Configuration

All runtime settings must come from environment variables.

| Variable | Example | Purpose |
|----------|---------|---------|
| `HTTP_ADDR` | `:8080` | API listen address |
| `MONGO_URI` | `mongodb://mongodb:27017` | MongoDB connection |
| `MONGO_DB` | `hello_world` | MongoDB database |
| `REDIS_ADDR` | `redis:6379` | Redis connection |
| `REDIS_TTL_SECONDS` | `300` | Cache TTL |
| `SHUTDOWN_TIMEOUT_SECONDS` | `10` | Graceful shutdown deadline |
| `LOG_LEVEL` | `info` | Logging level |

No secrets should be committed. Docker Compose may use local `.env.example` values only; production secrets should use GCP Secret Manager or platform secret configuration.

## 10. Reliability and Resilience

- Configure MongoDB and Redis connection timeouts.
- Return quickly on cache failures and fall back to MongoDB.
- Use bounded MongoDB operations with context deadlines.
- Use a cache TTL and an explicit stable cache key to prevent unbounded stale data.
- Shut down the HTTP server gracefully, drain active requests, close clients, and honor the configured shutdown timeout.
- Redis and MongoDB failures must be represented separately in logs and metrics.

The initial design does not require retries or a circuit breaker, but service interfaces should allow them to be added later. If retries are added, use exponential backoff and a low retry count.

## 11. Security

- Run containers as non-root users.
- Use read-only container filesystems where practical.
- Restrict Docker Compose ports to the host only when needed.
- Do not expose MongoDB or Redis externally in the local deployment.
- Supply credentials only through environment variables or secret managers.
- Add TLS for GCP/external traffic after the initial deployment.
- Validate every query parameter before use.
- Set standard HTTP security headers on the frontend.

## 12. Docker Compose Design

Required services:

| Service | Purpose |
|---------|---------|
| `frontend` | React build served by Nginx |
| `backend` | Gin API |
| `mongodb` | Message storage |
| `redis` | Cache |
| `prometheus` | Metrics collection |
| `grafana` | Dashboards |

Compose dependencies should use service health checks rather than relying only on container start order.

Prometheus should scrape:

- `backend:8080/metrics`

Grafana should be provisioned with datasource and dashboard configuration.

## 13. GCP Deployment

### Initial preferred option: Cloud Run

- Deploy frontend and backend as separate Cloud Run services.
- Put the frontend behind an HTTPS load balancer or Cloud Run domain.
- Store MongoDB in MongoDB Atlas or a managed/self-hosted GCP VM.
- Use Memorystore for Redis.
- Export backend metrics to Managed Prometheus.
- Use Grafana Cloud or a deployed Grafana instance.
- Configure Secret Manager for database and Redis credentials.
- Set independent CPU, memory, concurrency, and minimum instance settings for frontend and backend.

### Kubernetes/SRE option

Use GKE for the preferred learning deployment:

- Deployments for frontend and backend.
- Statefulness and persistent volumes for MongoDB and Redis in the lab environment.
- Prometheus and Grafana installed with Helm.
- Kubernetes liveness, readiness, and startup probes.
- Horizontal Pod Autoscaler for the backend.
- Resource requests and limits.
- Pod disruption budgets for production-like services.

## 14. Grafana Dashboards

Create three dashboards:

1. **Application**
   - Request count.
   - Error count.
   - Success rate.
   - P50, P95, and P99 response time.

2. **Cache**
   - Cache hits.
   - Cache misses.
   - Cache hit ratio.
   - Cache errors.

3. **Database**
   - MongoDB query count.
   - Query latency.
   - Database errors.
   - Redis fallback behavior.

The infrastructure dashboard should show CPU, memory, and network usage when GKE or GCP managed infrastructure metrics are available.

## 15. Testing Strategy

### Unit tests

- Query parameter validation.
- Cache hit path.
- Cache miss plus database read.
- Redis failure fallback.
- MongoDB repository behavior.
- Metrics and request-ID middleware.
- Graceful shutdown.

### Integration tests

- Frontend-to-backend request.
- Backend-to-MongoDB request.
- Backend-to-Redis cache hit and miss.
- Health endpoint.
- Prometheus metrics endpoint.
- Container startup and dependency readiness.

### End-to-end validation

1. Frontend starts and loads `Hello World`.
2. Cache disabled forces a MongoDB read.
3. Cache enabled performs a cache miss and warms Redis.
4. A second cache-enabled request returns `source: "cache"`.
5. Redis is stopped; database-backed request still succeeds.
6. MongoDB is stopped; API returns 500 and metrics/logs record the error.
7. Invalid parameters return 400.
8. Docker Compose starts and Prometheus sees backend metrics.

### Performance validation

- Average cache hit latency below 100 ms.
- Average database read latency below 500 ms.
- Run k6 load tests against cache-hit and database-read paths.
- Verify P95 latency and error rate after the load test.

## 16. Delivery Phases

### Phase 1 - Local application

- Build the React frontend.
- Build the Gin API and configuration layer.
- Add MongoDB repository and seed data.
- Add Redis cache-aside behavior.
- Implement health, metrics, structured logging, and request IDs.

### Phase 2 - Containers

- Add backend, frontend, MongoDB, Redis Dockerfiles.
- Add Docker Compose and health checks.
- Configure Prometheus and Grafana dashboards.

### Phase 3 - GCP

- Deploy frontend and backend.
- Configure Managed Prometheus or Grafana.
- Use Memorystore or a documented Redis alternative.
- Configure secrets and external MongoDB access.
- Complete end-to-end validation.

### Phase 4 - Optional enhancements

- PostgreSQL repository.
- OpenTelemetry tracing.
- TLS and authentication.
- CI/CD and Kubernetes.
- Rate limiting, autoscaling, and k6 load testing.

## 17. Acceptance Criteria

The implementation is complete when:

- The frontend displays the message returned by the backend.
- MongoDB provides the message on a cache miss or cache-disabled request.
- Redis cache hits return `source: "cache"`.
- Redis failure does not break database-backed requests.
- `/health` returns status `UP`.
- `/metrics` exposes all required Prometheus metrics.
- Grafana displays application, cache, database, and infrastructure metrics.
- Docker Compose starts all services successfully.
- The application is deployed to GCP and passes end-to-end validation.
- Cache and database average latency meet their respective targets.
