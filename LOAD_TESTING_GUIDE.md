# 🔥 Load Testing Guide - SRE Demonstration

## Специальный endpoint для симуляции перегрузки

Добавлен endpoint: `GET /api/v1/loadtest`

### Параметры

| Параметр | Значение | Описание |
|----------|----------|----------|
| `type` | `cpu` | Нагрузка на CPU (вычисления) |
| `type` | `memory` | Нагрузка на память (аллокации) |
| `type` | `db` | Имитация нагрузки на БД |
| `type` | `goroutines` | Имитация утечки goroutine |
| `type` | `all` | Все типы нагрузки одновременно |
| `duration` | `10` (сек) | Длительность нагрузки (макс 60s) |
| `workers` | `4` | Количество воркеров (макс 20) |

---

## 🎯 Сценарии для Демо

### Сценарий 1: Trigger HighResponseTime Alert (WARNING)

**Цель**: Показать алерт "High Response Time"

**Шаг 1**: Запустить нагрузку
```bash
# CPU нагрузка на 5 минут (алерт срабатывает через 5 мин среднего >1s)
curl "http://localhost:4000/api/v1/loadtest?type=cpu&duration=60&workers=10"
```

**Шаг 2**: Параллельно генерировать обычные запросы
```bash
# В отдельном терминале - спам обычных запросов
for i in {1..1000}; do
  curl -s http://localhost:4000/api/v1/cars > /dev/null &
done
```

**Шаг 3**: Мониторинг
- Открой Grafana: http://localhost:3001
- Смотри панель "Avg Response Time" - должен вырасти
- Открой Prometheus: http://localhost:9090/alerts
- Через 5 минут алерт HighResponseTime станет FIRING

**Ожидание**: Average response time > 1.0s for 5 minutes → WARNING alert

---

### Сценарий 2: Показать Saturation (Перегрузка системы)

**Цель**: Показать рост активных запросов и CPU usage

```bash
# Максимальная нагрузка на 60 секунд
curl "http://localhost:4000/api/v1/loadtest?type=all&duration=60&workers=15"
```

**Мониторинг в Grafana**:
- "Active Requests" - вырастет
- "Database Connections" - может вырасти
- "Node Exporter - CPU Usage" - покажет рост CPU
- "Node Exporter - Memory Usage" - покажет рост RAM

---

### Сценарий 3: Memory Pressure

**Цель**: Показать влияние на память

```bash
# Нагрузка на память
curl "http://localhost:4000/api/v1/loadtest?type=memory&duration=30"
```

**Мониторинг**:
```bash
# В отдельном терминале - смотри память процесса
watch -n 1 'docker stats --no-stream car-rental-backend'
```

---

### Сценарий 4: Goroutine Leak Simulation

**Цель**: Показать утечку goroutine

```bash
# Утечка goroutine
curl "http://localhost:4000/api/v1/loadtest?type=goroutines&duration=30"
```

**Проверка**:
```bash
# Количество goroutine в процессе
curl http://localhost:4000/metrics | grep http_requests_active
```

---

## 🚀 Полная Демонстрация для Защиты

### Пошаговый Script

```bash
#!/bin/bash

echo "🎯 SRE Demo - Load Testing Sequence"
echo "===================================="

# 1. Show system is healthy
echo "1️⃣  Checking healthy system..."
curl -s http://localhost:4000/health | python3 -m json.tool
sleep 2

# 2. Show current metrics
echo "2️⃣  Current metrics..."
curl -s http://localhost:4000/metrics | head -n 20
sleep 2

# 3. Start CPU load
echo "3️⃣  Starting CPU load (60s, 10 workers)..."
curl "http://localhost:4000/api/v1/loadtest?type=cpu&duration=60&workers=10"
sleep 2

# 4. Generate normal traffic during load
echo "4️⃣  Generating normal traffic..."
for i in {1..50}; do
  curl -s http://localhost:4000/api/v1/cars > /dev/null &
done

# 5. Monitor Grafana dashboard
echo "5️⃣  Open Grafana: http://localhost:3001"
echo "    Watch 'Avg Response Time' panel"
echo "    Watch 'Active Requests' panel"

# 6. Check Prometheus alerts
echo "6️⃣  Check alerts: http://localhost:9090/alerts"
echo "    HighResponseTime should become FIRING after 5 min"

# 7. Show Docker stats
echo "7️⃣  Container resource usage:"
docker stats --no-stream car-rental-backend

echo "✅ Demo complete!"
```

---

## 🔥 Быстрые Команды для Презентации

### Минимальная (показать рост response time)
```bash
curl "http://localhost:4000/api/v1/loadtest?type=cpu&duration=30&workers=8"
```

### Средняя (показать все метрики)
```bash
curl "http://localhost:4000/api/v1/loadtest?type=all&duration=30&workers=10"
```

### Максимальная (полная перегрузка)
```bash
curl "http://localhost:4000/api/v1/loadtest?type=all&duration=60&workers=15"
```

---

## 📊 Ожидания vs Реальность

| Метрика | Норма | Под Нагрузкой | Алерт |
|---------|-------|---------------|-------|
| Response Time | < 100ms | > 1000ms | HighResponseTime (>1s for 5m) |
| Active Requests | 1-5 | 50-100+ | ActiveRequestsSpike (>100 for 3m) |
| CPU Usage | 5-15% | 70-100% | Node Exporter показывает пик |
| Memory | 50MB | 500MB+ | Node Exporter показывает рост |

---

## ⚠️ Важно

1. **duration макс = 60s** - защита от зависания
2. **workers макс = 20** - защита от OOM
3. **Endpoint только для демо** - в production нужно удалить или защитить
4. **Алерты требуют времени**:
   - HighResponseTime: 5 минут среднего > 1s
   - ActiveRequestsSpike: 3 минуты > 100 запросов

---

## 🎬 Live Demo Script для Защиты

```
1. Показать Grafana dashboard (нормальное состояние)
   → Все метрики зелёные

2. Запустить нагрузку:
   curl "http://localhost:4000/api/v1/loadtest?type=all&duration=60&workers=10"

3. Обновить Grafana
   → Response Time растёт (жёлтый → красный)
   → Active Requests растёт
   → CPU Usage растёт

4. Показать Prometheus alerts
   → HighResponseTime: PENDING → FIRING

5. Остановить нагрузку (просто ждать окончания 60s)

6. Показать восстановление
   → Метрики возвращаются к норме
   → Alert resolving
```

---

## 🛠️ Альтернативные Инструменты

### Apache Bench (ab)
```bash
# 1000 запросов, 50 параллельных
ab -n 1000 -c 50 http://localhost:4000/api/v1/cars
```

### Hey
```bash
# 1000 запросов за 30 секунд
hey -n 1000 -c 50 -z 30s http://localhost:4000/api/v1/cars
```

### Vegeta
```bash
# 100 req/s в течение 60 секунд
echo "GET http://localhost:4000/api/v1/cars" | vegeta attack -rate=100 -duration=60s | vegeta report
```

---

**💡 Совет**: Для демонстрации используй `type=all&duration=30&workers=10` - это даст видимый эффект за 30 секунд без долгого ожидания алертов.
