# 📊 Frontend Observability - Добавленные компоненты

## Что добавлено во фронтенд

### 1. Health Endpoint
**Файл**: `frontend/app/server/api/health.get.ts`

**URL**: `GET /api/health`

**Response**:
```json
{
  "status": "healthy",
  "timestamp": "2026-04-12T10:30:00.000Z",
  "version": "1.0.0",
  "service": "car-rental-frontend",
  "framework": "nuxt",
  "nodeVersion": "v22.x.x",
  "uptime": 3600,
  "memoryUsage": {
    "rss": 150000000,
    "heapTotal": 80000000,
    "heapUsed": 60000000,
    "external": 5000000,
    "arrayBuffers": 1000000
  }
}
```

**Используется**:
- Docker HEALTHCHECK instruction
- Kubernetes liveness/readiness probes
- Load balancer health checks

---

### 2. Metrics Endpoint (Prometheus Format)
**Файл**: `frontend/app/server/api/metrics.get.ts`

**URL**: `GET /api/metrics`

**Actions**:
- `?action=track&path=/cars` - Track page view
- `?action=track-api&endpoint=/api/v1/cars` - Track API call
- `?action=error` - Track error

**Prometheus Output**:
```prometheus
# HELP frontend_uptime_seconds Frontend uptime in seconds
# TYPE frontend_uptime_seconds gauge
frontend_uptime_seconds 3600

# HELP frontend_memory_rss_bytes Resident set size in bytes
# TYPE frontend_memory_rss_bytes gauge
frontend_memory_rss_bytes 150000000

# HELP frontend_memory_heap_bytes Heap size in bytes
# TYPE frontend_memory_heap_bytes gauge
frontend_memory_heap_bytes 60000000

# HELP frontend_page_views_total Total page views per path
# TYPE frontend_page_views_total counter
frontend_page_views_total{path="/"} 150
frontend_page_views_total{path="/cars"} 320
frontend_page_views_total{path="/profile"} 85

# HELP frontend_api_calls_total Total API calls per endpoint
# TYPE frontend_api_calls_total counter
frontend_api_calls_total{endpoint="/api/v1/cars"} 500
frontend_api_calls_total{endpoint="/api/v1/rentals"} 120

# HELP frontend_errors_total Total frontend errors
# TYPE frontend_errors_total counter
frontend_errors_total 3
```

---

### 3. Auto-Tracking Middleware
**Файл**: `frontend/app/middleware/analytics.global.ts`

**Что делает**:
- Автоматически трекает каждый переход на новую страницу
- Вызывается при каждом change route
- Отправляет метрику в `/api/metrics`

**Пример**:
```
User navigates to /cars → Middleware triggers → 
POST /api/metrics?action=track&path=/cars
```

---

### 4. Enhanced API Composable
**Файл**: `frontend/app/composables/useApi.ts`

**Что добавлено**:
- Автоматический трекинг всех API вызовов
- Автоматический трекинг ошибок
- Без влияния на основную логику (try/catch с ignore errors)

**Пример**:
```typescript
// Original call - now automatically tracked
const cars = await authFetch('/api/v1/cars')

// Automatically triggers:
// POST /api/metrics?action=track-api&endpoint=/api/v1/cars

// If error occurs:
// POST /api/metrics?action=error
```

---

## Grafana Dashboard - Новые Панели

### Panel 14: Frontend - Memory Usage (RSS)
- **Metric**: `frontend_memory_rss_bytes / 1024 / 1024`
- **Unit**: MB
- **Thresholds**: Green < 200MB, Yellow < 500MB, Red > 500MB
- **Purpose**: Monitor Node.js memory leaks

### Panel 15: Frontend - Page Views (by path)
- **Metric**: `rate(frontend_page_views_total[5m])`
- **Unit**: reqps
- **Labels**: path (/, /cars, /profile, etc.)
- **Purpose**: See which pages are most popular

### Panel 16: Frontend - API Calls (by endpoint)
- **Metric**: `rate(frontend_api_calls_total[5m])`
- **Unit**: reqps
- **Labels**: endpoint
- **Purpose**: See which backend endpoints are most used

### Panel 17: Frontend - Errors Total
- **Metric**: `frontend_errors_total`
- **Thresholds**: Green < 10, Yellow < 50, Red > 50
- **Purpose**: Track frontend-side errors

---

## Как это работает

### Page View Tracking Flow
```
User navigates to /cars
  ↓
analytics.global.ts middleware triggers
  ↓
$fetch('/api/metrics?action=track&path=/cars')
  ↓
Metrics counter increments
  ↓
Prometheus scrapes /api/metrics
  ↓
Grafana shows page view rate
```

### API Call Tracking Flow
```
Component calls authFetch('/api/v1/cars')
  ↓
useApi.ts tracks the call
  ↓
$fetch('/api/metrics?action=track-api&endpoint=/api/v1/cars')
  ↓
Metrics counter increments
  ↓
Prometheus scrapes /api/metrics
  ↓
Grafana shows API call rate by endpoint
```

### Error Tracking Flow
```
API call fails (network error, 5xx, etc.)
  ↓
useApi.ts catch block triggers
  ↓
$fetch('/api/metrics?action=error')
  ↓
Error counter increments
  ↓
Prometheus scrapes /api/metrics
  ↓
Grafana shows error count
```

---

## Тестирование

### 1. Проверить Health Endpoint
```bash
curl http://localhost:3000/api/health | python3 -m json.tool
```

**Expected**:
```json
{
  "status": "healthy",
  "service": "car-rental-frontend",
  "uptime": 123.456
}
```

### 2. Проверить Metrics Endpoint
```bash
curl http://localhost:3000/api/metrics
```

**Expected**: Prometheus-format output с метриками

### 3. Симулировать Page Views
```bash
# Открыть несколько страниц
curl http://localhost:3000/
curl http://localhost:3000/cars
curl http://localhost:3000/profile

# Проверить метрики
curl http://localhost:3000/api/metrics | grep page_views
```

### 4. Проверить в Grafana
1. Открой http://localhost:3001
2. Dashboard: "Car Rental API - SRE Dashboard"
3. Scroll down to "Frontend" panels
4. Should show:
   - Memory usage graph
   - Page views by path
   - API calls by endpoint
   - Error count

---

## Почему это важно для SRE

### 1. Full-Stack Observability
- **Backend**: API latency, error rate, DB connections
- **Frontend**: Memory leaks, page views, client errors
- **Infrastructure**: CPU, RAM, disk I/O

### 2. Early Detection
- Memory leaks in Node.js detected before OOM
- High error rate on specific pages
- Slow API endpoints affecting UX

### 3. Business Insights
- Most visited pages (traffic patterns)
- Most used API endpoints (usage patterns)
- Error correlation with deployments

---

## Files Changed/Created

| File | Type | Purpose |
|------|------|---------|
| `frontend/app/server/api/health.get.ts` | **NEW** | Health check endpoint |
| `frontend/app/server/api/metrics.get.ts` | **NEW** | Prometheus metrics endpoint |
| `frontend/app/middleware/analytics.global.ts` | **NEW** | Auto page view tracking |
| `frontend/app/composables/useApi.ts` | **MODIFIED** | Added API call tracking |
| `Dockerfile.frontend` | **MODIFIED** | Updated healthcheck URL |
| `observability/grafana/dashboards/car-rental-sre-dashboard.json` | **MODIFIED** | Added 4 frontend panels |

---

## Команды для Демо

```bash
# 1. Проверить health
curl http://localhost:3000/api/health

# 2. Проверить metrics
curl http://localhost:3000/api/metrics

# 3. Симулировать нагрузку на фронтенд
for i in {1..20}; do
  curl -s http://localhost:3000/cars > /dev/null &
  curl -s http://localhost:3000/profile > /dev/null &
  curl -s http://localhost:3000/rentals > /dev/null &
done

# 4. Показать рост memory
curl http://localhost:3000/api/metrics | grep frontend_memory

# 5. Показать page views
curl http://localhost:3000/api/metrics | grep page_views
```

---

## 💡 Benefits

✅ **No external dependencies** - Pure Nuxt server API routes  
✅ **Zero configuration** - Auto-provisioned with Grafana  
✅ **Automatic tracking** - No manual instrumentation needed  
✅ **Production-ready** - Non-blocking, error-tolerant  
✅ **Prometheus-compatible** - Standard text format  
✅ **Real-time insights** - See frontend health in Grafana  

---

**Теперь фронтенд полностью мониторингирован!** 🎯
