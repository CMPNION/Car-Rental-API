import { defineEventHandler, getRouterParam } from '#imports'

// In-memory metrics store
const metrics = {
  pageViews: {} as Record<string, number>,
  apiCalls: {} as Record<string, number>,
  errors: 0,
  startTime: Date.now(),
}

// Track page views
const trackPageView = (path: string) => {
  metrics.pageViews[path] = (metrics.pageViews[path] || 0) + 1
}

// Track API calls
const trackApiCall = (endpoint: string) => {
  metrics.apiCalls[endpoint] = (metrics.apiCalls[endpoint] || 0) + 1
}

export default defineEventHandler((event) => {
  const action = getRouterParam(event, 'action')

  if (action === 'track') {
    const path = getQuery(event).path as string || '/'
    trackPageView(path)
    return { success: true, path, count: metrics.pageViews[path] }
  }

  if (action === 'track-api') {
    const endpoint = getQuery(event).endpoint as string || '/'
    trackApiCall(endpoint)
    return { success: true, endpoint, count: metrics.apiCalls[endpoint] }
  }

  if (action === 'error') {
    metrics.errors++
    return { success: true, totalErrors: metrics.errors }
  }

  // Return all metrics in Prometheus format
  const memUsage = process.memoryUsage()
  const uptime = process.uptime()
  const now = Date.now()

  let prometheusOutput = ''

  // Frontend uptime
  prometheusOutput += '# HELP frontend_uptime_seconds Frontend uptime in seconds\n'
  prometheusOutput += '# TYPE frontend_uptime_seconds gauge\n'
  prometheusOutput += `frontend_uptime_seconds ${Math.floor(uptime)}\n\n`

  // Memory usage
  prometheusOutput += '# HELP frontend_memory_rss_bytes Resident set size in bytes\n'
  prometheusOutput += '# TYPE frontend_memory_rss_bytes gauge\n'
  prometheusOutput += `frontend_memory_rss_bytes ${memUsage.rss}\n\n`

  prometheusOutput += '# HELP frontend_memory_heap_bytes Heap size in bytes\n'
  prometheusOutput += '# TYPE frontend_memory_heap_bytes gauge\n'
  prometheusOutput += `frontend_memory_heap_bytes ${memUsage.heapUsed}\n\n`

  // Page views
  prometheusOutput += '# HELP frontend_page_views_total Total page views per path\n'
  prometheusOutput += '# TYPE frontend_page_views_total counter\n'
  for (const [path, count] of Object.entries(metrics.pageViews)) {
    prometheusOutput += `frontend_page_views_total{path="${path}"} ${count}\n`
  }
  prometheusOutput += '\n'

  // API calls
  prometheusOutput += '# HELP frontend_api_calls_total Total API calls per endpoint\n'
  prometheusOutput += '# TYPE frontend_api_calls_total counter\n'
  for (const [endpoint, count] of Object.entries(metrics.apiCalls)) {
    prometheusOutput += `frontend_api_calls_total{endpoint="${endpoint}"} ${count}\n`
  }
  prometheusOutput += '\n'

  // Errors
  prometheusOutput += '# HELP frontend_errors_total Total frontend errors\n'
  prometheusOutput += '# TYPE frontend_errors_total counter\n'
  prometheusOutput += `frontend_errors_total ${metrics.errors}\n`

  return prometheusOutput
})
