"use client"

import React, { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import MovieFilters from '../components/MovieFilters'
import MovieList from '../components/MovieList'
import UserMenu from '../components/UserMenu'
import { getCurrentUser } from '../utils/storage'
import './styles.css'

export default function MenuPage() {
  const router = useRouter()
  const [mode, setMode] = useState<'general' | 'recommended'>('general')
  const [filters, setFilters] = useState({ genre: '', max: 10 })
  const [user, setUser] = useState<any>(null)

  useEffect(() => {
    const u = getCurrentUser()
    if (!u) router.replace('/usuarios')
    setUser(u)
  }, [router])

  if (!user) return null

  return (
    <main>
      <header className='header-menu'>
        <div className="container-fluid contenedor-header">
          <div className="row align-items-center">
            <div className="col-9">
              <div className="container-fluid">
                <div className="row">
                  <div className="col-12">
                    <h2 className='titulo-bienvenida'>Bienvenido, {user.name}</h2>
                  </div>
                  <div className="col-12">
                    <h1 className='titulo-menu'>Menú principal</h1>
                  </div>
                </div>
              </div>
            </div>

            <div className="col-3 text-end" style={{ paddingRight: 50 }}>
              <UserMenu />
            </div>
          </div>
        </div>
      </header>

      <section style={{ backgroundColor: '#ffe9faff', minHeight: '100vh' }}>

        <div className="container-fluid contenedor-seccion-contenido">
          <div className="row">
            <div className="col-12">

              <h2 className='subtitulo-seccion'>Selecciona el tipo de contenido que deseas ver</h2>

            </div>
            <div className="col-12 mt-2">

              <div className="container-fluid">
                <div className="row justify-content-start align-items-center">
                  <div className="col-3">
                    <label className={mode === 'general' ? 'activo' : 'no-activo'}>
                      <input type="radio" checked={mode === 'general'} onChange={() => setMode('general')} 
                      /> Películas generales
                    </label>
                  </div>
                  <div className="col-3">
                    <label className={mode === 'recommended' ? 'activo' : 'no-activo'}>
                      <input type="radio" checked={mode === 'recommended'} onChange={() => setMode('recommended')} 
                      /> Recomendadas
                    </label>
                  </div>
                </div>
              </div>

            </div>
          </div>
        </div>

        <div className="container-fluid">

          <div className="row align-items-center justify-content-start">

            <div className="col-10">

              <div className="container-fluid contenedor-seccion-filtros">
                <div className="row align-items-center">

                  <div className="col-4" style={{ paddingLeft: 40 }}>
                    <h3 className='subtitulo-filtros m-0'>Filtros:</h3>
                  </div>
                  <div className="col-8">
                    <MovieFilters value={filters} onChange={setFilters} />
                  </div>

                </div>
              </div>

            </div>

            <div className="col-11">
              <MovieList mode={mode} filters={filters} userId={user.id} />
            </div>

          </div>

        </div>
      </section>
    </main>
  )
}
