# Hello World Cloud-Native Application

## Overview

This project is a cloud-native "Hello World" application designed to demonstrate modern software engineering, backend development, caching, observability, containerization, and cloud deployment practices.

The application consists of three primary components:

1. Frontend (UI)
2. Backend API
3. Database Layer

The solution will be containerized using Docker and deployed to Google Cloud Platform (GCP).

---

# Objectives

The project aims to demonstrate:

- Frontend to Backend communication
- Database access patterns
- Cache integration
- REST API design
- Application monitoring
- Containerization
- Cloud deployment
- Performance optimization through caching
- SRE and observability practices

---

# High-Level Architecture

```text
+------------------+
|     Frontend     |
|      React       |
+---------+--------+
          |
          |
          v
+---------+--------+
|     Backend      |
|      Golang      |
+----+--------+----+
     |        |
     |        |
     v        v
+--------+  +--------+
| Redis  |  | MongoDB|
| Cache  |  | Storage|
+--------+  +--------+

          |
          |
          v
+------------------+
| Prometheus       |
+------------------+
          |
          v
+------------------+
| Grafana          |
+------------------+
```

---

# Functional Requirements

## Frontend

### FR-1

Frontend shall display a message received from the backend API.

Example:

```html
<h1>Hello World</h1>
```

### FR-2

Frontend shall invoke the backend endpoint:

```http
GET /message?fromdb=1&cache=1&sql=0
```

### FR-3

Frontend shall display the value returned by the backend.

Example response:

```json
{
  "message": "Hello World",
  "source": "mongodb"
}
```

### FR-4

Frontend shall handle API failures gracefully.

Example:

```text
Unable to load message.
```

---

## Backend

### FR-5

Backend shall expose REST APIs for retrieving application messages.

### FR-6

Backend shall support cache-enabled and cache-disabled execution paths.

### FR-7

Backend shall retrieve data from:

- Database
- Cache

depending on request parameters.

### FR-8

Backend shall expose health-check endpoints.

```http
GET /health
```

Response:

```json
{
  "status": "UP"
}
```

### FR-9

Backend shall expose Prometheus metrics endpoint.

```http
GET /metrics
```

---

## Database

### FR-10

Database shall store application messages.

Sample document:

```json
{
  "_id": "1",
  "message": "Hello World"
}
```

### FR-11

Database shall support either:

- MongoDB (Preferred)
- PostgreSQL (Future enhancement)

---

## Cache

### FR-12

Redis shall be used as the caching layer.

### FR-13

If caching is enabled:

```text
1. Check Redis
2. Return cached value if present
3. Read Database if cache miss
4. Populate Redis
5. Return response
```

### FR-14

Cache responses shall include source information.

Example:

```json
{
  "message": "Hello World",
  "source": "cache"
}
```

---

# API Specification

## GET /message

### Request

```http
GET /message?fromdb=1&cache=1&sql=0
```

### Query Parameters

| Parameter | Description |
|------------|-------------|
| fromdb | Read from database |
| cache | Enable cache lookup |
| sql | Select SQL or NoSQL database |

### Response

```json
{
  "message": "Hello World",
  "source": "mongodb"
}
```

### HTTP Status Codes

| Code | Description |
|--------|-------------|
| 200 | Success |
| 400 | Invalid Request |
| 500 | Internal Error |

---

# Monitoring Requirements

## Prometheus Metrics

### Request Metrics

```text
http_requests_total
```

### Response Latency

```text
http_request_duration_seconds
```

### Cache Metrics

```text
cache_hits_total
cache_misses_total
```

### Database Metrics

```text
db_reads_total
db_errors_total
```

---

# Grafana Dashboard Requirements

## Application Dashboard

Display:

- Request Count
- Error Count
- Success Rate
- Response Time

## Cache Dashboard

Display:

- Cache Hits
- Cache Misses
- Cache Hit Ratio

## Database Dashboard

Display:

- Query Count
- Query Latency
- Database Errors

## Infrastructure Dashboard

Display:

- CPU Usage
- Memory Usage
- Network Usage

---

# Non-Functional Requirements

## Performance

### NFR-1

Average API latency should be less than:

```text
100 ms (Cache Hit)
```

### NFR-2

Average API latency should be less than:

```text
500 ms (Database Read)
```

---

## Reliability

### NFR-3

Application shall support graceful shutdown.

### NFR-4

Application shall recover from transient Redis failures.

---

## Security

### NFR-5

Application configuration shall be supplied through environment variables.

### NFR-6

No secrets shall be hardcoded in source code.

---

## Observability

### NFR-7

Structured logging shall be implemented.

Example:

```json
{
  "request_id":"123",
  "endpoint":"/message",
  "duration_ms":10
}
```

### NFR-8

Application shall expose Prometheus-compatible metrics.

---

# Technology Stack

## Frontend

- React
- Axios
- Nginx

## Backend

- Golang
- Gin Framework

## Cache

- Redis

## Database

- MongoDB

## Monitoring

- Prometheus
- Grafana

## Containerization

- Docker
- Docker Compose

## Cloud

- Google Cloud Platform (GCP)

---

# Local Development Components

```text
frontend
backend
mongodb
redis
prometheus
grafana
```

---

# Deployment Architecture (GCP)

## Option A - Cloud Run

```text
Frontend   -> Cloud Run
Backend    -> Cloud Run
MongoDB    -> Atlas / VM
Redis      -> Memorystore
Monitoring -> Managed Prometheus
Grafana    -> Grafana Cloud
```

## Option B - Kubernetes (Preferred for SRE Learning)

```text
Frontend   -> GKE
Backend    -> GKE
MongoDB    -> StatefulSet
Redis      -> StatefulSet
Prometheus -> Helm Chart
Grafana    -> Helm Chart
```

---

# Future Enhancements

- PostgreSQL Support
- OpenTelemetry Tracing
- Distributed Tracing
- TLS Encryption
- Kubernetes Deployment
- CI/CD Pipeline
- GitHub Actions
- Load Testing using k6
- Horizontal Pod Autoscaling
- JWT Authentication
- API Gateway
- Rate Limiting

---

# Success Criteria

The project is considered complete when:

1. Frontend displays "Hello World" received from backend.
2. Backend reads message from database.
3. Redis caching is operational.
4. Prometheus metrics are collected.
5. Grafana dashboards display application metrics.
6. All services run successfully using Docker Compose.
7. Application is deployed successfully to GCP.
8. End-to-end validation is completed.