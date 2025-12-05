export function setCurrentUser(user: any) {
  if (typeof window === 'undefined') return
  localStorage.setItem('currentUser', JSON.stringify(user))
  // also store user data under usuario{ID}_data
  const usuario = localStorage.getItem(`usuario${user.id}_data`)
  if (!usuario) {
    localStorage.setItem(`usuario${user.id}_data`, JSON.stringify(user))
  }
}

export function getCurrentUser() {
  if (typeof window === 'undefined') return null
  const raw = localStorage.getItem('currentUser')
  return raw ? JSON.parse(raw) : null
}

export function saveRecommendationMetrics(userId: number, metrics: any) {
  if (typeof window === 'undefined') return
  const key = `usuario${userId}`
  const raw = localStorage.getItem(key)
  const existing = raw ? JSON.parse(raw) : { count: 0, sums: {} }

  const sums = existing.sums || {}
  const keys = ['cpu_percent', 'cpu_percent_per_cpu', 'cpu_system_seconds', 'elapsed_ms', 'num_cpu', 'mem_sys']
  keys.forEach((k) => {
    sums[k] = (sums[k] || 0) + (metrics[k] || 0)
  })

  const updated = { count: (existing.count || 0) + 1, sums }
  localStorage.setItem(key, JSON.stringify(updated))
}

export function getUserMetricsAggregate(userId: number) {
  if (typeof window === 'undefined') return null
  const raw = localStorage.getItem(`usuario${userId}`)
  if (!raw) return null
  const parsed = JSON.parse(raw)
  const c = parsed.count || 1
  const s = parsed.sums || {}
  return {
    cpu_percent_avg: (s.cpu_percent || 0) / c,
    cpu_percent_per_cpu_avg: (s.cpu_percent_per_cpu || 0) / c,
    cpu_system_seconds_avg: (s.cpu_system_seconds || 0) / c,
    elapsed_ms_avg: (s.elapsed_ms || 0) / c,
    num_cpu_avg: (s.num_cpu || 0) / c,
    mem_sys_avg: (s.mem_sys || 0) / c
  }
}

export function getUserData(userId: number) {
  if (typeof window === 'undefined') return null
  const raw = localStorage.getItem(`usuario${userId}_data`)
  return raw ? JSON.parse(raw) : null
}

export function clearCurrentUser() {
  if (typeof window === 'undefined') return
  localStorage.removeItem('currentUser')
}
