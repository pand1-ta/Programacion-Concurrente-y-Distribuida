"use client"

import React, { useEffect, useState } from 'react'
import { getCurrentUser, getUserMetricsAggregate } from '../utils/storage'

export default function AdminPanel() {
  const [agg, setAgg] = useState<any>(null)
  const [user, setUser] = useState<any>(null)

  useEffect(() => {
    const u = getCurrentUser()
    setUser(u)
    if (u) {
      setAgg(getUserMetricsAggregate(u.id))
    }
  }, [])

  if (!user) return <div>No hay usuario</div>

  if (!agg) return <div>Sin métricas. Haz recomendaciones para ver tus métricas.</div>

  return (
    <div style={{ minWidth: 260 }}>
      <h3>Panel admin</h3>
      <div>Usuario: {user.name}</div>
      <div>Email: {user.email}</div>
      <hr />
      <div>Métricas promedio:</div>
      <ul>
        <li>CPU % promedio: {agg.cpu_percent_avg?.toFixed(2)}</li>
        <li>CPU % por núcleo promedio: {agg.cpu_percent_per_cpu_avg?.toFixed(2)}</li>
        <li>Tiempo CPU promedio (s): {agg.cpu_system_seconds_avg?.toFixed(3)}</li>
        <li>Tiempo consulta (ms) promedio: {agg.elapsed_ms_avg?.toFixed(1)}</li>
        <li>Nodos promedio: {agg.num_cpu_avg?.toFixed(1)}</li>
        <li>Memoria promedio (mem_sys): {agg.mem_sys_avg?.toFixed(0)}</li>
      </ul>
    </div>
  )
}
