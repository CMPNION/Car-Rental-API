# SRE Observability Report: Car Rental API

## Executive Summary

This document outlines the Service Level Indicators (SLIs), Service Level Objectives (SLOs), and Error Budget calculations for the Car Rental API application. The observability stack has been implemented using Docker Compose with Prometheus for metrics collection and Grafana for visualization.

---

## Step 1: Containerized Architecture

### Services Deployed

| Service | Technology | Port | Purpose |
|---------|-----------|------|---------|
| Backend | Go (GORM + SQLite) | 4000 | REST API for car rental operations |
| Frontend | Nuxt 4 (Vue 3) | 3000 | Web UI for customers and admins |
| Prometheus | Prometheus | 9090 | Metrics collection and alerting |
| Grafana | Grafana | 3001 | Dashboard visualization |
| Node Exporter | Node Exporter | 9100 | Host-level metrics (CPU, memory, disk) |

### Dockerfiles

- **Backend**: Multi-stage build (golang:1.24.3-alpine → alpine:3.21) with CGO enabled for SQLite support
- **Frontend**: Multi-stage build (node:22-alpine → node:22-alpine) with production-optimized Nuxt output

### Architecture Diagram

```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│   Browser   │────>│   Frontend   │────>│   Backend    │
│             │     │  (Nuxt:3000) │     │   (Go:4000)  │
└─────────────┘     └──────────────┘     └──────┬───────┘
                                                       │
                                                       v
                                              ┌─────────────────┐
                                              │   SQLite (file) │
                                              └─────────────────┘

┌──────────────────────────────────────────────────────────────┐
│                    Observability Stack                        │
├────────────────┬──────────────────┬──────────────────────────┤
│   Prometheus   │     Grafana      │     Node Exporter        │
│    (:9090)     │     (:3001)      │        (:9100)           │
│                │                  │                          │
│ Scrapes:       │ Dashboards:      │ Exposes:                 │
│ - Backend      │ - Golden Signals │ - CPU Usage              │
│ - Frontend     │ - SLO Compliance │ - Memory Usage           │
│ - Node Exporter│ - Error Budget   │ - Disk I/O               │
└────────────────┴──────────────────┴──────────────────────────┘
```

---

## Step 2: Custom Reliability Targets (SLI, SLO, Error Budget)

### SLI #1: API Availability (Success Rate)

**Definition**: The percentage of successful HTTP requests (non-5xx responses) out of total requests.

**Formula**:
```
Availability SLI = (Total Requests - 5xx Requests) / Total Requests × 100%
```

**Why this matters for Car Rental API**: 
- Customers need to browse cars, make rentals, and process payments
- 5xx errors directly impact user experience and revenue
- Payment failures (5xx on `/api/v1/rentals/{id}/pay`) cause immediate customer dissatisfaction

**SLO Target**: **99.9% availability per 30-day window**

This means:
- Out of 100,000 requests, only 100 can fail with 5xx errors
- Allows for 43 minutes and 49 seconds of downtime per month

**Measurement**:
```promql
# Prometheus query for availability
(1 - (sum(rate(http_requests_total{route=~".*5xx"}[30d])) / sum(rate(http_requests_total[30d])))) * 100
```

---

### SLI #2: API Latency (Response Time)

**Definition**: The percentage of requests served within 500ms (p90 latency target).

**Formula**:
```
Latency SLI = (Requests served in < 500ms) / Total Requests × 100%
```

**Why this matters for Car Rental API**:
- Slow responses frustrate users browsing cars or completing bookings
- Payment processing should be fast to avoid duplicate submissions
- Admin dashboard queries (analytics) can be slightly slower but still need bounds

**SLO Target**: **99% of requests served within 500ms**

This means:
- Out of 100 requests, 99 must complete in under 500ms
- 1 request can be slower (allows for complex queries like admin metrics)

**Measurement**:
```promql
# Prometheus query for latency compliance
sum(rate(http_request_duration_seconds_bucket{le="0.5"}[30d])) / sum(rate(http_request_duration_seconds_count[30d])) * 100
```

---

### Error Budget Calculation

**Definition**: The maximum amount of time your service can be non-compliant with its SLO in a given month.

#### For Availability SLO (99.9%)

**Monthly Error Budget**:
```
Total minutes in a month = 30 days × 24 hours × 60 minutes = 43,200 minutes

Allowed downtime = 43,200 × (1 - 0.999)
                 = 43,200 × 0.001
                 = 43.2 minutes per month
                 = 43 minutes and 12 seconds
```

**In terms of requests**:
```
If service receives 100,000 requests/month:
  Allowed 5xx errors = 100,000 × (1 - 0.999)
                     = 100 requests
```

#### For Latency SLO (99% < 500ms)

**Monthly Error Budget**:
```
If service receives 100,000 requests/month:
  Allowed slow requests (>500ms) = 100,000 × (1 - 0.99)
                                 = 1,000 requests
```

#### Combined Error Budget Dashboard

The Grafana dashboard displays:
```
Error Budget Remaining = (SLO Target - Actual Error Rate) / SLO Target × 100%

For availability:
  = (0.001 - (5xx_rate / total_rate)) / 0.001 × 100%

Example:
  If current 5xx rate is 0.05% (0.0005):
  Budget remaining = (0.001 - 0.0005) / 0.001 × 100%
                   = 50%
```

**Interpretation**:
- **100% budget**: Perfect month, no errors
- **50% budget**: Half the allowed errors used
- **0% budget**: SLO violated, no room for more errors
- **Negative budget**: SLO already violated

---

## Step 3: Custom Monitoring & Dashboards (Grafana)

### Dashboard Overview: "Car Rental API - SRE Dashboard"

The Grafana dashboard is auto-provisioned and contains the following panels organized by the **Four Golden Signals**:

#### 1. Traffic (Request Rate)
- **Panel**: "Total Requests (5m)" - Current request rate in req/s
- **Panel**: "Request Rate (by endpoint)" - Time series breakdown per route
- **Purpose**: Understand load patterns and detect anomalies

#### 2. Latency (Response Time)
- **Panel**: "Avg Response Time" - Current average latency with thresholds (green < 0.5s, yellow < 1.0s, red > 1.0s)
- **Panel**: "Response Time Distribution" - Time series of average latency
- **Purpose**: Ensure users have a fast experience

#### 3. Errors (Error Rate)
- **Panel**: "Error Rate (5xx)" - Percentage of requests failing with 5xx, with thresholds (green < 1%, yellow < 5%, red > 5%)
- **Purpose**: Detect bugs, infrastructure issues, and cascading failures

#### 4. Saturation (Resource Utilization)
- **Panel**: "Active Requests" - Current concurrent requests (gauge)
- **Panel**: "Database Connections" - Open DB connections (gauge with thresholds)
- **Panel**: "Node Exporter - CPU Usage" - Host CPU utilization
- **Panel**: "Node Exporter - Memory Usage" - Host memory utilization
- **Purpose**: Know when systems are approaching capacity limits

#### SLO Compliance Panels
- **Panel**: "SLO Compliance - Availability" - Shows current availability % (target: 99.9%)
- **Panel**: "SLO Compliance - Latency (p90 < 500ms)" - Shows latency compliance %
- **Panel**: "Error Budget Remaining (Monthly)" - Remaining error budget as percentage

### Provisioning

The dashboard is automatically loaded via Grafana provisioning:
```
observability/grafana/
├── provisioning/
│   ├── datasources/
│   │   └── datasources.yml    # Auto-configures Prometheus
│   └── dashboards/
│       └── dashboards.yml      # Auto-imports dashboards
└── dashboards/
    └── car-rental-sre-dashboard.json
```

---

## Step 4: Alerting Validation (Prometheus)

### Alert Rules Configuration

Located in: `observability/prometheus/alert_rules.yml`

#### Alert #1: HighErrorRate (CRITICAL)

**Purpose**: Detect when the API is returning too many 5xx errors.

**PromQL Expression**:
```yaml
rate(http_requests_total{route=~".*5xx"}[5m]) / rate(http_requests_total[5m]) > 0.05
```

**Conditions**:
- Error rate exceeds 5% of total requests
- Sustained for 2 minutes
- Severity: **critical**

**Annotations**:
- Summary: "High HTTP error rate detected"
- Description: "More than 5% of requests are failing with 5xx errors. Current value: {{ $value | humanizePercentage }}"

**Response Runbook**:
1. Check Grafana dashboard for error spike
2. Review recent deployments or configuration changes
3. Check database connectivity
4. Review backend logs for stack traces
5. Consider rolling back if deployment-related

---

#### Alert #2: HighResponseTime (WARNING)

**Purpose**: Detect when the API is responding slowly.

**PromQL Expression**:
```yaml
rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m]) > 1.0
```

**Conditions**:
- Average response time exceeds 1.0 second
- Sustained for 5 minutes
- Severity: **warning**

**Annotations**:
- Summary: "High HTTP response time detected"
- Description: "Average response time is above 1 second. Current value: {{ $value }}s"

**Response Runbook**:
1. Check database query performance
2. Review admin metrics endpoint load
3. Check for slow external dependencies
4. Consider adding caching or query optimization

---

#### Alert #3: BackendServiceDown (CRITICAL)

**Purpose**: Detect when the backend is completely unreachable.

**PromQL Expression**:
```yaml
up{job="backend"} == 0
```

**Conditions**:
- Backend scrape target fails
- Sustained for 1 minute
- Severity: **critical**

---

#### Alert #4: FrontendServiceDown (WARNING)

**Purpose**: Detect when the frontend is unreachable.

**PromQL Expression**:
```yaml
up{job="frontend"} == 0
```

**Conditions**:
- Frontend scrape target fails
- Sustained for 1 minute
- Severity: **warning**

---

#### Alert #5: DatabaseConnectionsHigh (CRITICAL)

**Purpose**: Prevent database connection exhaustion.

**PromQL Expression**:
```yaml
db_connections_open > 50
```

**Conditions**:
- Open DB connections exceed 50
- Sustained for 2 minutes
- Severity: **critical**

---

#### Alert #6: ActiveRequestsSpike (WARNING)

**Purpose**: Detect unusual traffic spikes.

**PromQL Expression**:
```yaml
http_requests_active > 100
```

**Conditions**:
- Active requests exceed 100 concurrent requests
- Sustained for 3 minutes
- Severity: **warning**

---

### Alert Testing Procedure

To manually trigger an alert for validation:

#### Test 1: Trigger BackendServiceDown Alert

```bash
# Stop the backend service
docker-compose stop backend

# Wait 1-2 minutes
# Check Prometheus alerts: http://localhost:9090/alerts
# Verify alert shows FIRING status

# Restart backend
docker-compose start backend
```

#### Test 2: Trigger HighResponseTime Alert

```bash
# Generate slow requests using a load testing tool
# Or temporarily add a sleep to an endpoint for testing

# Example: Add temporary sleep to car handlers for testing
time.Sleep(2 * time.Second)

# Send requests for 5+ minutes
# Verify alert fires in Prometheus
```

---

## Deployment Instructions

### Prerequisites

- Docker 20.10+
- Docker Compose 2.0+
- 4GB+ RAM
- 10GB+ disk space

### Quick Start

```bash
# Build and start all services
docker-compose up -d --build

# Check service health
docker-compose ps

# View logs
docker-compose logs -f backend
docker-compose logs -f frontend

# Access services
echo "Frontend: http://localhost:3000"
echo "Backend Health: http://localhost:4000/health"
echo "Backend Metrics: http://localhost:4000/metrics"
echo "Prometheus: http://localhost:9090"
echo "Grafana: http://localhost:3001 (admin/admin)"
```

### Verification Steps

1. **Check Backend Health**:
   ```bash
   curl http://localhost:4000/health
   # Expected: {"status":"healthy","database":"connected",...}
   ```

2. **Check Metrics Endpoint**:
   ```bash
   curl http://localhost:4000/metrics
   # Expected: Prometheus-format metrics
   ```

3. **Verify Prometheus is Scraping**:
   - Open http://localhost:9090/targets
   - All targets should show "UP" status

4. **Verify Grafana Dashboard**:
   - Open http://localhost:3001
   - Login: admin / admin
   - Navigate to "Car Rental API - SRE Dashboard"

5. **Test Alerts**:
   - Open http://localhost:9090/alerts
   - Stop backend: `docker-compose stop backend`
   - Wait 1-2 minutes
   - Verify alert shows "FIRING" status
   - Restart backend: `docker-compose start backend`

---

## Bonus: Docker Swarm Deployment (Optional +10 Points)

To deploy using Docker Swarm instead of docker-compose:

```bash
# Initialize Docker Swarm
docker swarm init

# Deploy the stack
docker stack deploy -c docker-swarm.yml car-rental

# Check service status
docker service ls

# Scale backend to 3 replicas
docker service scale car-rental_backend=3

# View logs
docker service logs car-rental_backend

# Rollback if needed
docker service update --rollback car-rental_backend
```

### docker-swarm.yml

```yaml
version: '3.8'

services:
  backend:
    image: car-rental-backend:latest
    ports:
      - "4000:4000"
    environment:
      - JWT_SECRET=car-rental-secret-key
    deploy:
      replicas: 2
      restart_policy:
        condition: on-failure
      resources:
        limits:
          cpus: "1"
          memory: 512M
    networks:
      - car-rental-net

  frontend:
    image: car-rental-frontend:latest
    ports:
      - "3000:3000"
    environment:
      - NUXT_PUBLIC_API_BASE=http://backend:4000
    deploy:
      replicas: 1
      restart_policy:
        condition: on-failure
    networks:
      - car-rental-net

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./observability/prometheus:/etc/prometheus:ro
    deploy:
      replicas: 1
    networks:
      - car-rental-net

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_USER=admin
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - ./observability/grafana:/var/lib/grafana:ro
    deploy:
      replicas: 1
    networks:
      - car-rental-net

  node-exporter:
    image: prom/node-exporter:latest
    ports:
      - "9100:9100"
    deploy:
      mode: global
    networks:
      - car-rental-net

networks:
  car-rental-net:
    driver: overlay
```

---

## Conclusion

This observability stack provides:

✅ **Complete Containerization**: Multi-stage Dockerfiles for backend and frontend with health checks  
✅ **Mathematically Defined SLOs**: 99.9% availability and 99% < 500ms latency with calculated error budgets  
✅ **Real-Time Monitoring**: Grafana dashboard with Golden Signals and SLO compliance tracking  
✅ **Proactive Alerting**: 6 alert rules (3 critical, 3 warning) with runbooks  
✅ **Production-Ready**: Non-root users, health checks, restart policies, and resource limits  

The system is designed to detect, alert, and visualize reliability issues before they impact customers.
