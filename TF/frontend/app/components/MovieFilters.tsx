"use client"

import React from 'react'

export default function MovieFilters({ value, onChange }: any) {
  return (
    <div>
      <label style={{color: '#ffffff'}}>
        Género:
        <input
          value={value.genre}
          onChange={(e) => onChange({ ...value, genre: e.target.value })}
          placeholder="ej: action"
          style={{ marginLeft: 8, color: '#000000' }}
          className='input'
        />
      </label>
      <label style={{ marginLeft: 12, color: '#ffffff' }}>
        Máx. películas:
        <input
          type="number"
          value={value.max}
          onChange={(e) => onChange({ ...value, max: Number(e.target.value) })}
          min={1}
          style={{ width: 80, marginLeft: 8, color: '#000000' }}
          className='input'
        />
      </label>
    </div>
  )
}
