"use client"

import React, { useState } from 'react'
import AdminPanel from './AdminPanel'
import './styles.css'
import { useRouter } from 'next/navigation'
import { clearCurrentUser } from '../utils/storage'

export default function UserMenu() {
  const [open, setOpen] = useState(false)
  const router = useRouter()

  function volverAlMenu() {
    clearCurrentUser()
    router.push('/usuarios')
  }

  return (
    <div className="container-fluid contenedor-menu-metricas" style={{ position: 'relative' }}>
      <div className="row justify-content-end align-items-center">
        <div className="col-3">
          <img src="./metricas.png" alt="metricas-logo" className='img-fluid'/>
        </div>

        <div className="col-6 text-center">
          <button onClick={() => setOpen((s) => !s)} className='btn-menu'>Menu usuario</button>
        </div>

      <div className="col-3 text-start">
        <button onClick={() => volverAlMenu()} className='btn-menu'>Cerrar Sesión</button>
      </div>

      </div>
      {open && (
        <div style={{ position: 'absolute', right: 0, top: '100%', background: '#fff', padding: 12, border: '1px solid #ccc' }}>
          <AdminPanel />
        </div>
      )}
    </div>
  )
}
