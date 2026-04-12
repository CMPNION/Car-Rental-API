# 🚗 Car Rental API - SRE Observability Stack

## Quick Start

### Prerequisites
- Docker 20.10+
- Docker Compose 2.0+
- 4GB+ RAM
- 10GB+ disk space

### Deploy the Stack

```bash
# Build and start all services
docker-compose up -d --build

# This will start:
# - Backend API (Go) on port 4000
# - Frontend (Nuxt) on port 3000
# - Prometheus on port 9090
# - Grafana on port 3001
# - Node Exporter on port 9100
```

### Access Services

| Service | URL | Credentials |
|---------|-----|-------------|
| **Frontend** | http://localhost:3000 | N/A |
| **Backend API** | http://localhost:4000 | N/A |
| **Health Check** | http://localhost:4000/health | N/A |
| **Metrics** | http://localhost:4000/metrics | N/A |
| **Prometheus** | http://localhost:9090 | N/A |
| **Grafana** | http://localhost:3001 | admin / admin |
| **Node Exporter** | http://localhost:9100/metrics | N/A |

---

## Architecture

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

## File Structure

```
Car-Rental-API/
├── Dockerfile.backend                  # Multi-stage build for Go backend
├── Dockerfile.frontend                 # Multi-stage build for Nuxt frontend
├── docker-compose.yml                  # Docker Compose configuration
├── docker-swarm.yml                    # Docker Swarm configuration (BONUS +10)
├── .dockerignore                       # Docker build context exclusions
│
├── observability/
│   ├── prometheus/
│   │   ├── prometheus.yml              # Prometheus scrape config
│   │   └── alert_rules.yml             # Alert rules (6 alerts)
│   └── grafana/
│       ├── provisioning/
│       │   ├── datasources/
│       │   │   └── datasources.yml     # Auto-configure Prometheus datasource
│       │   └── dashboards/
│       │       └── dashboards.yml      # Auto-import dashboards
│       └── dashboards/
│           └── car-rental-sre-dashboard.json  # SRE Dashboard
│
├── SRE_Report.md                       # Full SRE documentation
├── Presentation_Slides.md              # Presentation slides
└── README_OBSERVABILITY.md             # This file
```

---

## SLOs & Error Budgets

### SLI #1: API Availability
- **SLO**: 99.9% success rate (non-5xx responses)
- **Monthly Error Budget**: 43.2 minutes of allowed downtime
- **Measurement**: `(1 - (5xx_requests / total_requests)) × 100%`

### SLI #2: API Latency
- **SLO**: 99% of requests served within 500ms
- **Monthly Error Budget**: 1,000 slow requests per 100k total
- **Measurement**: `requests_with_latency_<500ms / total_requests × 100%`

### Error Budget Tracking
View in Grafana dashboard:
- **SLO Compliance - Availability**: Shows current availability %
- **SLO Compliance - Latency**: Shows latency compliance %
- **Error Budget Remaining**: Shows remaining monthly budget %

---

## Alert Rules

### Critical Alerts
1. **HighErrorRate**: Error rate > 5% for 2 minutes
2. **BackendServiceDown**: Backend unreachable for 1 minute
3. **DatabaseConnectionsHigh**: DB connections > 50 for 2 minutes

### Warning Alerts
1. **HighResponseTime**: Avg response time > 1s for 5 minutes
2. **FrontendServiceDown**: Frontend unreachable for 1 minute
3. **ActiveRequestsSpike**: Active requests > 100 for 3 minutes

View alerts in Prometheus: http://localhost:9090/alerts

---

## Monitoring Guide

### Golden Signals Dashboard

The Grafana dashboard displays four key metrics:

1. **Traffic**: Request rate by endpoint (requests/second)
2. **Latency**: Average response time with color-coded thresholds
3. **Errors**: 5xx error rate as percentage of total traffic
4. **Saturation**: Active requests, DB connections, CPU, memory usage

### Useful Prometheus Queries

```promql
# Current request rate
sum(rate(http_requests_total[5m]))

# Average response time
rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m])

# Error rate (5xx)
sum(rate(http_requests_total{route=~".*5xx"}[5m])) / sum(rate(http_requests_total[5m]))

# Availability (30-day window)
(1 - (sum(rate(http_requests_total{route=~".*5xx"}[30d])) / sum(rate(http_requests_total[30d])))) * 100

# Error budget remaining
(0.001 - (sum(rate(http_requests_total{route=~".*5xx"}[30d])) / sum(rate(http_requests_total[30d])))) / 0.001 * 100
```

---

## Testing & Validation

### Verify Backend Health

```bash
curl http://localhost:4000/health
# Expected response:
# {"status":"healthy","timestamp":"...","version":"1.0.0","database":"connected"}
```

### Verify Metrics Endpoint

```bash
curl http://localhost:4000/metrics
# Expected: Prometheus-format metrics
```

### Verify Prometheus Scraping

1. Open http://localhost:9090/targets
2. All targets should show "UP" status (green)

### Verify Grafana Dashboard

1. Open http://localhost:3001
2. Login: admin / admin
3. Navigate to "Car Rental API - SRE Dashboard"
4. Verify all panels show data

### Test Alerting

```bash
# Stop backend to trigger BackendServiceDown alert
docker-compose stop backend

# Wait 1-2 minutes
# Check Prometheus alerts: http://localhost:9090/alerts
# Verify alert shows FIRING status

# Restart backend
docker-compose start backend
```

---

## Docker Swarm (BONUS +10 Points)

Deploy using Docker Swarm for high availability:

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

# Update with rolling deployment
docker service update --image car-rental-backend:v2 car-rental_backend

# Rollback if needed
docker service update --rollback car-rental_backend
```

### Swarm Features
- **Backend**: 2 replicas for load balancing
- **Node Exporter**: Global mode (runs on all nodes)
- **Rolling updates**: Zero-downtime deployments
- **Auto-rollback**: On deployment failure
- **Resource limits**: Per-container CPU/memory limits

---

## Troubleshooting

### Backend Won't Start

```bash
# Check logs
docker-compose logs backend

# Common issues:
# - Port 4000 already in use
# - SQLite permission issues
```

### Frontend Can't Connect to Backend

```bash
# Verify backend is running
docker-compose ps

# Check network connectivity
docker-compose exec frontend wget -qO- http://backend:4000/health
```

### Prometheus Not Scraping Metrics

```bash
# Check Prometheus config
docker-compose exec prometheus cat /etc/prometheus/prometheus.yml

# Verify targets
curl http://localhost:9090/api/v1/targets

# Reload config without restart
curl -X POST http://localhost:9090/-/reload
```

### Grafana Dashboard Empty

```bash
# Check datasource configuration
# Navigate to: Configuration > Data Sources > Prometheus
# Ensure URL is: http://prometheus:9090

# Verify data exists in Prometheus first
curl http://localhost:9090/api/v1/query?query=up
```

---

## Cleanup

```bash
# Stop all services
docker-compose down

# Remove volumes (deletes all data)
docker-compose down -v

# Remove images
docker-compose down --rmi all

# Remove everything
docker system prune -a --volumes
```

---

## Deliverables Checklist

- [x] Dockerfile for backend (multi-stage build)
- [x] Dockerfile for frontend (multi-stage build)
- [x] docker-compose.yml with all services
- [x] Health check endpoint (`/health`)
- [x] Prometheus metrics endpoint (`/metrics`)
- [x] Prometheus configuration with scrape targets
- [x] Alert rules (6 alerts: 3 critical, 3 warning)
- [x] Grafana dashboard with Golden Signals
- [x] SLO compliance panels
- [x] Error budget tracking
- [x] SRE Report (SRE_Report.md)
- [x] Presentation Slides (Presentation_Slides.md)
- [x] Docker Swarm configuration (docker-swarm.yml) - BONUS

---

## Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [SRE Book (Google)](https://sre.google/sre-book/table-of-contents/)
- [The Four Golden Signals](https://sre.google/sre-book/monitoring-distributed-systems/#xref_monitoring_golden-signals)

---

## License

This project is for educational purposes as part of the SRE Midterm Project.
