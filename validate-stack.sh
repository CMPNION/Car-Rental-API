#!/bin/bash

# SRE Observability Stack - Deployment Validation Script
# This script validates all components are working correctly

set -e

echo "=========================================="
echo "🚗 Car Rental API - SRE Stack Validation"
echo "=========================================="
echo ""

# Check if Docker is running
echo "✓ Checking Docker..."
if ! docker info > /dev/null 2>&1; then
    echo "✗ Docker is not running. Please start Docker first."
    exit 1
fi
echo "  Docker is running"
echo ""

# Check if services are running
echo "✓ Checking Services Status..."
docker-compose ps
echo ""

# Test Backend Health
echo "✓ Testing Backend Health..."
if curl -f -s http://localhost:4000/health > /dev/null; then
    echo "  ✅ Backend is healthy"
    curl -s http://localhost:4000/health | python3 -m json.tool 2>/dev/null || curl -s http://localhost:4000/health
else
    echo "  ❌ Backend health check failed"
fi
echo ""

# Test Metrics Endpoint
echo "✓ Testing Metrics Endpoint..."
if curl -f -s http://localhost:4000/metrics > /dev/null; then
    echo "  ✅ Metrics endpoint is responding"
    echo "  Sample metrics:"
    curl -s http://localhost:4000/metrics | head -n 10
else
    echo "  ❌ Metrics endpoint failed"
fi
echo ""

# Test Frontend
echo "✓ Testing Frontend..."
if curl -f -s http://localhost:3000/ > /dev/null; then
    echo "  ✅ Frontend is responding"
else
    echo "  ⚠️  Frontend may still be starting up"
fi
echo ""

# Test Prometheus
echo "✓ Testing Prometheus..."
if curl -f -s http://localhost:9090/-/healthy > /dev/null; then
    echo "  ✅ Prometheus is healthy"
    echo "  Checking scrape targets..."
    curl -s http://localhost:9090/api/v1/targets | python3 -c "
import sys, json
data = json.load(sys.stdin)
for target in data['data']['activeTargets']:
    status = target['health']
    label = target['labels']['job']
    emoji = '✅' if status == 'up' else '❌'
    print(f'    {emoji} {label}: {status}')
" 2>/dev/null || echo "    (Install python3 for detailed target info)"
else
    echo "  ❌ Prometheus health check failed"
fi
echo ""

# Test Grafana
echo "✓ Testing Grafana..."
if curl -f -s http://localhost:3001/api/health -u admin:admin > /dev/null; then
    echo "  ✅ Grafana is healthy"
else
    echo "  ⚠️  Grafana may still be starting up"
fi
echo ""

# Summary
echo "=========================================="
echo "📊 Service URLs:"
echo "=========================================="
echo "  Frontend:        http://localhost:3000"
echo "  Backend API:     http://localhost:4000"
echo "  Health Check:    http://localhost:4000/health"
echo "  Metrics:         http://localhost:4000/metrics"
echo "  Prometheus:      http://localhost:9090"
echo "  Grafana:         http://localhost:3001 (admin/admin)"
echo "  Node Exporter:   http://localhost:9100"
echo ""
echo "=========================================="
echo "🎯 Next Steps:"
echo "=========================================="
echo "  1. Open Grafana: http://localhost:3001"
echo "  2. Login: admin / admin"
echo "  3. Navigate to 'Car Rental API - SRE Dashboard'"
echo "  4. Check Prometheus alerts: http://localhost:9090/alerts"
echo ""
echo "✅ Validation Complete!"
echo "=========================================="
