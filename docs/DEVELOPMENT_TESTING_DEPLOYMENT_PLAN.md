# Hello World Cloud-Native Application: Development, Testing, and Deployment Plan

Based on [DESIGN.md](./DESIGN.md) and [REQUIREMENT.md](./REQUIREMENT.md), this document outlines an end-to-end plan to take the project from zero code to a production-grade local Dockerized stack and GCP deployment.

---

## 1. Architecture & Technology Baseline

```mermaid
flowchart TD
    subgraph Client ["Client Layer"]
        Browser["User Browser"]
    end

    subgraph FrontendService ["Frontend Container (Nginx :80)"]
        ReactApp["React SPA"]
        NginxProxy["Nginx Reverse Proxy (/api -> backend:8080)"]
    end

    subgraph BackendService ["Backend Container (Gin :8080)"]
        GinRouter["Gin HTTP Router"]
        MW["Middleware (Logger, Request-ID, Prometheus)"]
        MsgService["Message Service (Cache-Aside Logic)"]
        Repo["Message Repository"]
        Cache["Redis Cache Client"]
    end

    subgraph DataStorage ["Data & Cache Layer"]
        Redis[(Redis :6379)]
        MongoDB[(MongoDB :27017)]
    end

    subgraph Observability ["Observability Stack"]
        Prometheus["Prometheus (:9090)"]
        Grafana["Grafana (:3000)"]
    end

    Browser -->|GET /| ReactApp
    Browser -->|GET /api/message...| NginxProxy
    NginxProxy -->|Proxy /message...| GinRouter
    GinRouter --> MW --> MsgService
    MsgService --> Cache
    MsgService --> Repo
    Cache -.->|Fallback on error| Repo
    Cache --> Redis
    Repo --> MongoDB
    Prometheus -->|Scrape /metrics| BackendService
    Grafana -->|Datasource| Prometheus
```

### Component Responsibility Matrix

| Component | Technology | Target Directory / Port | Key Responsibilities |
|---|---|---|---|
| **Frontend** | React, Axios, Nginx (Alpine) | `frontend/` (Port 3000 local / 80 container) | Call `/api/message?fromdb=1&cache=1&sql=0`, display payload/source, handle loading & errors. Reverse proxy `/api/*` to backend. |
| **Backend** | Go (1.23+), Gin, go-redis, mongo-go-driver | `backend/` (Port 8080) | Parameter validation, cache-aside logic, graceful degradation, JSON structured logging, Prometheus metrics, graceful shutdown. |
| **Cache** | Redis 7 (Alpine) | Port 6379 | In-memory key-value cache (`hello-world:message:1`), TTL = 300s. Best-effort cache writes. |
| **Database** | MongoDB 7 | Port 27017 | Persistent message store (`hello_world.messages`), seeded with `{ _id: "1", message: "Hello World" }`. |
| **Monitoring** | Prometheus + Grafana | Port 9090 (Prom) / 3000 (Grafana) | Scraping `/metrics` every 5s; Grafana dashboards for Application, Cache, Database, and Infra. |

---

## 2. Phase-by-Phase Development Plan

```mermaid
flowchart LR
    P1["Phase 1: Local Application Core<br/>(Go API, Mongo, Redis, React)"] --> P2["Phase 2: Containerization & Observability<br/>(Dockerfiles, Compose, Grafana)"]
    P2 --> P3["Phase 3: GCP Deployment<br/>(Cloud Run / GKE, VPC, Secrets)"]
    P3 --> P4["Phase 4: Hardening & Extensions<br/>(k6 load tests, CI/CD, Postgres)"]
```

### Phase 1: Local Application Core (Backend & Frontend)

#### Step 1.1: Backend Project Setup & Clean Architecture Skeleton
- **Path:** `backend/`
- **Actions:**
  - Initialize Go module: `go mod init github.com/personal/helloworld-backend`.
  - Establish folder layout:
    ```text
    backend/
      cmd/server/main.go
      internal/
        config/config.go
        handler/message_handler.go
        handler/health_handler.go
        service/message_service.go
        repository/message_repository.go
        repository/mongo_message_repository.go
        repository/postgres_message_repository.go (stub returning ErrNotImplemented)
        cache/message_cache.go
        cache/redis_message_cache.go
        metrics/prometheus.go
        middleware/request_id.go
        middleware/logger.go
        middleware/metrics.go
    ```
  - Define interfaces in `repository` and `cache` packages to enable easy mocking and future PostgreSQL extension.

#### Step 1.2: Configuration & Structured Logging
- Implement `internal/config`: Parse env variables (`HTTP_ADDR`, `MONGO_URI`, `MONGO_DB`, `REDIS_ADDR`, `REDIS_TTL_SECONDS`, `SHUTDOWN_TIMEOUT_SECONDS`, `LOG_LEVEL`) with defaults.
- Implement `middleware/logger.go`: Use `uber-go/zap` or standard library `log/slog` for zero-allocation JSON logs with fields: `timestamp`, `level`, `request_id`, `endpoint`, `duration_ms`, `status`, `source`.
- Implement `middleware/request_id.go`: Honor incoming `X-Request-ID` or generate `google/uuid`; set header in response.

#### Step 1.3: Data Layer & Cache Layer Implementation
- **MongoDB Repository:**
  - Connect with timeout context (e.g. 5s).
  - Implement `GetMessage(ctx)` querying `{ _id: "1" }` on collection `messages`.
  - Provide an idempotent initialization/seeding script (`init-mongo.js` or automatic startup seed) inserting `{ _id: "1", message: "Hello World" }`.
- **Redis Cache:**
  - Implement `Get` and `Set` with configured TTL (300s).
  - Key format: `hello-world:message:1`.
- **Service Layer (Cache-Aside Engine):**
  - Parameter validation:
    - `fromdb` must be `1` (return 400 if missing or != 1).
    - `sql` must be `0` (return 501 Not Implemented if `sql=1`).
    - `cache` accepts `0` or `1`.
  - If `cache=1`:
    - Attempt Redis `Get`. On hit: increment `cache_hits_total`, return `source: "cache"`.
    - On miss: increment `cache_misses_total`, query MongoDB.
    - On Redis error: increment `cache_errors_total`, log warning, fall back to MongoDB.
    - After DB read: best-effort Redis `Set` in background or inline.
  - Return `source: "mongodb"` on database read.

#### Step 1.4: Observability Middleware & Health Endpoints
- Implement `GET /health` returning `{"status": "UP"}`.
- Implement `GET /metrics` exporting Prometheus metrics:
  - `http_requests_total{method,status,endpoint}`
  - `http_request_duration_seconds{method,endpoint}` (histogram with buckets from 5ms to 1s)
  - `cache_hits_total`, `cache_misses_total`, `cache_errors_total`
  - `db_reads_total{database}`, `db_errors_total{database}`
- Implement Graceful Shutdown: Listen for `SIGINT`/`SIGTERM`, allow `SHUTDOWN_TIMEOUT_SECONDS` (10s) to drain inflight requests, close DB and Redis connections.

#### Step 1.5: Frontend Development (React SPA)
- **Path:** `frontend/`
- Initialize React project (Vite or Create React App) with Axios.
- UI Components:
  - Header & Status badge.
  - Card displaying `message` (e.g., `Hello World`) and execution `source` (`cache` or `mongodb`).
  - Loading spinner / indicator while the fetch is pending.
  - Graceful error banner: "Unable to load message." on 4xx/5xx or network disconnect.
  - Action buttons to test query params (`cache=1`, `cache=0`, refresh button).

---

### Phase 2: Containerization & Local Observability Stack

#### Step 2.1: Multi-stage Dockerfiles
- **`backend/Dockerfile`**:
  - Build stage: `golang:1.23-alpine` compiles statically linked binary (`CGO_ENABLED=0`).
  - Runtime stage: `distroless` or `alpine:latest` with non-root user (`USER nonroot:nonroot`), exposing port 8080.
- **`frontend/Dockerfile`**:
  - Build stage: `node:20-alpine` runs `npm run build`.
  - Runtime stage: `nginx:alpine` running as unprivileged user, serving React static bundle and hosting `nginx.conf`.
- **`nginx.conf`**:
  - Serves static assets at `/`.
  - Proxies `/api/(.*)` to `http://backend:8080/$1` (stripping `/api` prefix).
  - Injects security headers (`X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Content-Security-Policy`).

#### Step 2.2: Docker Compose Orchestration
- **File:** `docker-compose.yml`
- Services with health checks:
  1. `mongodb`: Mongo 7 image with health check `mongosh --eval "db.adminCommand('ping')"`.
  2. `redis`: Redis 7 image with health check `redis-cli ping`.
  3. `backend`: Depends on `mongodb` and `redis` with `condition: service_healthy`. Health check: `wget --spider http://localhost:8080/health`.
  4. `frontend`: Depends on `backend` with `condition: service_healthy`. Exposes port 3000 to host.
  5. `prometheus`: Scrapes `backend:8080/metrics` every 5 seconds.
  6. `grafana`: Provisions Prometheus datasource and pre-loads dashboards.

#### Step 2.3: Grafana Dashboards Provisioning
- Create automated YAML provisioning in `monitoring/grafana/provisioning/`:
  - `dashboards/application.json`: Request volume, error rate, P50/P95/P99 latency.
  - `dashboards/cache.json`: Cache hits, misses, hit ratio (`hits / (hits + misses)`), Redis error count.
  - `dashboards/database.json`: MongoDB reads, read duration histogram, connection errors.

---

### Phase 3: GCP Cloud Deployment

#### Option A: Serverless on Cloud Run (Recommended for MVP)
```mermaid
flowchart LR
    subgraph GCP ["Google Cloud Platform"]
        LB["Cloud Load Balancer / Custom Domain"]
        subgraph CloudRun ["Cloud Run Services"]
            CR_FE["frontend-service"]
            CR_BE["backend-service"]
        end
        VPC["Serverless VPC Access Connector"]
        MemRedis[("Cloud Memorystore (Redis)")]
        SecretMgr["Secret Manager"]
    end
    Atlas[("MongoDB Atlas / Compute Engine Mongo")]

    LB --> CR_FE
    LB --> CR_BE
    CR_FE -->|/api/*| CR_BE
    CR_BE --> VPC
    VPC --> MemRedis
    CR_BE -.-> SecretMgr
    CR_BE --> Atlas
```

1. **Artifact Registry:** Create repository `helloworld-repo` and push multi-arch Docker images for frontend and backend.
2. **Secret Manager:** Store `MONGO_URI` and `REDIS_AUTH` tokens securely.
3. **Networking & Cache:**
   - Create a Google Cloud Memorystore for Redis instance in the default VPC.
   - Configure a Serverless VPC Access connector to allow Cloud Run to access Memorystore on its private IP.
4. **Cloud Run Deployment:**
   - `backend`: Configure container port 8080, inject env vars and secrets, attach VPC connector, set min-instances=1 to eliminate cold starts, set concurrency to 80.
   - `frontend`: Configure container port 80, set `BACKEND_URL` or reverse proxy rule.
5. **Observability on GCP:**
   - Forward backend `/metrics` to Google Cloud Managed Service for Prometheus (GMP) or use Grafana Cloud agent.

#### Option B: Google Kubernetes Engine (GKE) for SRE / Lab Track
1. Setup GKE Standard or Autopilot cluster.
2. Deploy MongoDB and Redis using Kubernetes `StatefulSet` with `PersistentVolumeClaim` (or Bitnami Helm charts).
3. Deploy Backend and Frontend `Deployment` manifests with:
   - Resource requests and limits (e.g. backend: CPU 250m/500m, RAM 128Mi/256Mi).
   - Liveness Probe (`GET /health`), Readiness Probe (`GET /health`), Startup Probe.
   - `HorizontalPodAutoscaler` (HPA) targeting 70% CPU utilization.
   - `PodDisruptionBudget` (`minAvailable: 1`).
4. Install Prometheus and Grafana using `kube-prometheus-stack` Helm chart.

---

## 3. Comprehensive Testing Plan

```mermaid
flowchart TD
    UT["1. Unit Tests<br/>(Param validation, Mocks, Cache-Aside)"]
    IT["2. Integration Tests<br/>(Testcontainers / Real Mongo & Redis)"]
    E2E["3. End-to-End Test Scenarios<br/>(8 Scenarios in Docker Compose)"]
    PERF["4. Performance & SLA Testing<br/>(k6 Latency & Load Tests)"]
    CHAOS["5. Resilience & Chaos Tests<br/>(Redis down, Mongo down)"]

    UT --> IT --> E2E --> PERF --> CHAOS
```

### 3.1 Unit Testing Matrix (Go & React)

| Test Area | Target File | Scenario / Case | Expected Outcome |
|---|---|---|---|
| **Handler Validation** | `message_handler_test.go` | `fromdb=0` or missing | HTTP 400 Bad Request |
| **Handler Routing** | `message_handler_test.go` | `sql=1` | HTTP 501 Not Implemented |
| **Service Cache Hit** | `message_service_test.go` | Cache returns `"Hello World"` | Service returns `"cache"`, DB repo is NOT invoked |
| **Service Cache Miss** | `message_service_test.go` | Cache returns miss (`ErrCacheMiss`) | Mongo repo called, Redis `Set` invoked, returns `"mongodb"` |
| **Service Cache Fallback**| `message_service_test.go` | Redis client returns connection error | Warning logged, metric incremented, Mongo called, returns `"mongodb"` |
| **Graceful Shutdown** | `server_test.go` | Interrupt signal sent during sleep | Open requests complete before process exits |
| **Frontend Component** | `App.test.tsx` | Mock API success / failure | Displays message on 200, displays "Unable to load message." on 500/network error |

### 3.2 Integration Testing (Real Container Dependencies)
- Utilize `testcontainers-go` or local docker compose test fixture.
- Verify real MongoDB connection, document seeding, and query latency.
- Verify real Redis TTL expiration: ensure cached key expires after TTL seconds and subsequent read hits MongoDB.
- Verify `/metrics` endpoint increments Prometheus counters appropriately after HTTP calls.

### 3.3 End-to-End (E2E) Test Scenario Suite
Automate the 8 required E2E scenarios from Section 15 of DESIGN.md using a test runner (Bash / Python / Newman):

| # | Scenario Description | Execution Command / Action | Validation Assertion |
|---|---|---|---|
| **E2E-1** | Initial Cold Request | `curl -i "http://localhost:3000/api/message?fromdb=1&cache=1&sql=0"` | Status `200`, `source: "mongodb"`, `message: "Hello World"` |
| **E2E-2** | Cache Disabled Read | `curl -i "http://localhost:3000/api/message?fromdb=1&cache=0&sql=0"` | Status `200`, `source: "mongodb"` (Redis bypassed) |
| **E2E-3** | Cache Hit Validation | `curl -i "http://localhost:3000/api/message?fromdb=1&cache=1&sql=0"` (second call) | Status `200`, `source: "cache"` |
| **E2E-4** | Cache Fallback on Outage | `docker compose stop redis` then call endpoint with `cache=1` | Status `200`, `source: "mongodb"`, `cache_errors_total` counter incremented |
| **E2E-5** | DB Outage Failure | `docker compose stop mongodb` then call endpoint with `cache=0` | Status `500 Internal Error`, JSON error response |
| **E2E-6** | Invalid Parameters | `curl -i "http://localhost:3000/api/message?fromdb=abc"` | Status `400 Bad Request` |
| **E2E-7** | Unsupported SQL flag | `curl -i "http://localhost:3000/api/message?fromdb=1&cache=1&sql=1"` | Status `501 Not Implemented` |
| **E2E-8** | Observability Ingestion | `curl http://localhost:9090/api/v1/targets` | Target `backend:8080` state is `UP`, metric `http_requests_total` > 0 |

### 3.4 Performance & Latency SLA Testing (k6)
Define a k6 script (`tests/load/k6_test.js`) validating the non-functional requirements:
- **NFR-1 (Cache Hit SLA):** Average latency < 100 ms at 50 VU (virtual users).
- **NFR-2 (Database Read SLA):** Average latency < 500 ms at 20 VU (`cache=0`).
- **Assertion:** P95 < 150ms for cache hits, P95 < 600ms for database reads, 0% error rate under normal load.

---

## 4. Deployment & Operations Plan

### 4.1 Local Deployment Runbook

```bash
# 1. Clone repository & configure environment
cp .env.example .env

# 2. Build and launch all services with health checks
docker compose up --build -d

# 3. Verify health status of all containers
docker compose ps

# 4. Validate frontend in browser
open http://localhost:3000

# 5. Access Grafana dashboards
open http://localhost:3001 # (admin / admin)
```

### 4.2 GCP Cloud Run Deployment Workflow

```mermaid
sequenceDiagram
    autonumber
    actor Dev as Developer / CI
    participant AR as GCP Artifact Registry
    participant SM as GCP Secret Manager
    participant VPC as Serverless VPC Access
    participant CR as GCP Cloud Run
    participant Redis as Memorystore Redis
    participant Atlas as MongoDB Atlas

    Dev->>AR: Build & Push Docker images (backend & frontend)
    Dev->>SM: Store MONGO_URI and DB credentials
    Dev->>CR: Deploy Backend (link VPC, inject SM secrets, attach /health probe)
    CR->>VPC: Connect to Memorystore Redis (:6379)
    CR->>Atlas: Connect via secure Mongo connection string
    Dev->>CR: Deploy Frontend (configure reverse proxy to Backend URL)
    Dev->>CR: Run Smoke Tests on public Cloud Run URL
```

### 4.3 Environment Variable & Secret Configuration Matrix

| Variable | Environment | Source | Default / Example Value |
|---|---|---|---|
| `HTTP_ADDR` | Local / GCP | Plaintext Env | `:8080` |
| `MONGO_URI` | Local | Plaintext Env | `mongodb://mongodb:27017` |
| `MONGO_URI` | GCP Production | Secret Manager | `mongodb+srv://app_user:****@cluster.mongodb.net/hello_world` |
| `MONGO_DB` | All | Plaintext Env | `hello_world` |
| `REDIS_ADDR` | Local | Plaintext Env | `redis:6379` |
| `REDIS_ADDR` | GCP Production | Plaintext Env / Secret | `10.x.x.x:6379` (Memorystore Private IP) |
| `REDIS_TTL_SECONDS` | All | Plaintext Env | `300` |
| `SHUTDOWN_TIMEOUT_SECONDS` | All | Plaintext Env | `10` |
| `LOG_LEVEL` | All | Plaintext Env | `info` (or `debug` in dev) |

---

## 5. Traceability Matrix

| Requirement | Description | Component | Verification Stage |
|---|---|---|---|
| **FR-1 - FR-4** | React UI message display & error states | `frontend/` | Unit test & E2E-1 |
| **FR-5 - FR-7** | REST API & cache-enabled/disabled flow | `backend/` | Unit, Integration, E2E-2, E2E-3 |
| **FR-8** | Health check `GET /health` | `backend/` | Integration & Container health checks |
| **FR-9** | Prometheus metrics `GET /metrics` | `backend/` | Integration & E2E-8 |
| **FR-10 - FR-11**| MongoDB message storage & seed | `mongodb` | Integration test & seed script |
| **FR-12 - FR-14**| Redis cache-aside & `source` metadata | `backend/` | Unit test, E2E-3, E2E-4 |
| **NFR-1 - NFR-2**| Latency SLAs (<100ms cache, <500ms DB) | Backend + Infra | k6 performance test |
| **NFR-3 - NFR-4**| Graceful shutdown & Redis resilience | `backend/` | Unit test & E2E-4 (Chaos test) |
| **NFR-5 - NFR-6**| Env-based config & secret isolation | All | Config audit & Secret Manager |
| **NFR-7 - NFR-8**| JSON logging & Prometheus metrics | `backend/` | Log inspection & Grafana dashboard |

---

## 6. Execution Milestones & Delivery Schedule

```text
Week 1: Phase 1 - Local Core
  ├── Day 1: Backend skeleton, config, logging, and health endpoint
  ├── Day 2: MongoDB repository, seed script, and Redis cache client
  ├── Day 3: Cache-aside service logic, fallback handling, unit tests
  ├── Day 4: React UI development, loading & error states
  └── Day 5: Frontend-to-backend integration

Week 2: Phase 2 - Dockerization & Observability
  ├── Day 6: Multi-stage Dockerfiles (Backend & Frontend Nginx proxy)
  ├── Day 7: Docker Compose setup with health checks & dependencies
  ├── Day 8: Prometheus metrics instrumentation & scrape config
  └── Day 9-10: Grafana dashboards (Application, Cache, Database)

Week 3: Phase 3 - Testing & GCP Deployment
  ├── Day 11: End-to-end automated test suite (8 scenarios)
  ├── Day 12: k6 load testing & latency SLA validation
  ├── Day 13: GCP infrastructure provisioning (Cloud Run, Memorystore, Secrets)
  ├── Day 14: Deployment, smoke tests, and Managed Prometheus hookup
  └── Day 15: Documentation, runbooks, and final acceptance sign-off
```
