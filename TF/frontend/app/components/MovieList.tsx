"use client"

import React, { useEffect, useState } from 'react'
import { saveRecommendationMetrics } from '../utils/storage'

function generos(genres: string) {
  const partes = genres.split('|');
  return "Géneros: " + partes.flat().join(', ');
}

export default function MovieList({ mode, filters, userId }: any) {
  const [movies, setMovies] = useState<any[]>([])
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    setLoading(true)
    const q = new URLSearchParams()
    if (filters.genre) q.set('genre', filters.genre)
    q.set('limit', String(filters.max || 10))

    if (mode === 'general') {
      // GET /api/movies?genre=..&page=1&limit=..
      fetch(`/api/movies?${q.toString()}`)
        .then((r) => r.json())
        .then((res) => {
          setMovies(res || [])
        })
        .catch(() => setMovies([]))
        .finally(() => setLoading(false))
    } else {
      // recommended: use WebSocket to backend /ws/recommend/{userId}
      // Build WS URL from NEXT_PUBLIC_RECOMMENDER_BACKEND or current origin
      const buildWsUrl = () => {
        // FORZAR siempre localhost para pruebas locales
        const forced = 'http://localhost:8080'
        const origin = forced
        const protocol = origin.startsWith('https') ? 'wss' : 'ws'
        // strip http(s):// if present
        const host = origin.replace(/^https?:\/\//, '')
        const url = `${protocol}://${host}/ws/recommend/${userId}?${q.toString()}`
        if (typeof window !== 'undefined') console.debug('[MovieList] WS URL (FORCED localhost) ->', url)
        return url
      }

      let ws: WebSocket | null = null
      let wsUrl = ''
      try {
        wsUrl = buildWsUrl()
        ws = new WebSocket(wsUrl)
      } catch (e) {
        console.error('[MovieList] WebSocket creation failed', e, wsUrl)
        setMovies([])
        setLoading(false)
      }

      if (ws) {
        ws.addEventListener('message', (ev) => {
          try {
            const data = JSON.parse(ev.data)
            if (Array.isArray(data)) {
              setMovies(data)
            } else if (data && data.movies) {
              setMovies(data.movies || [])
              if (data.metrics) saveRecommendationMetrics(userId, data.metrics)
            }
          } catch (err) {
            // If the backend sends plain text or unexpected payload, ignore
          } finally {
            setLoading(false)
          }
        })

        ws.addEventListener('error', () => {
          setMovies([])
          setLoading(false)
        })

        // Close socket when component unmounts or deps change
        const cleanup = () => {
          try {
            ws?.close()
          } catch (_) {}
        }

        // attach cleanup to effect
        // `useEffect` cleanup is returned below by returning the function
        return cleanup
      }
    }
  }, [mode, filters, userId])

  return (
    <div style={{ marginTop: 16 }}>
      {loading && <div>Cargando...</div>}
      {!loading && movies.length === 0 && <div>No hay películas</div>}
        <div className="container-fluid">
          <div className="row justify-content-center align-items-center">
            {movies.slice(0, filters.max).map((m: any) => (
              <div className="col-3" key={m.movieId}>
                <div className="container-fluid">
                  <div className="row justify-content-center align-items-center">
                    <div className="col-8 contenedor-pelicula">
                      <div className="container-fluid">
                        <div className="row justify-content-center align-items-center">
                          <div className="col-12 p-3">
                            <img src="./pelicula.png" alt="pelicula-placeholder" className='img-fluid img-pelicula'/>
                          </div>
                          <div className="col-12">
                            <p className='titulo-pelicula'>{m.title}</p>
                          </div>
                          <div className="col-12 mb-3">
                            <p className='generos'>{generos(m.genre)}</p>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
    </div>
  )
}
