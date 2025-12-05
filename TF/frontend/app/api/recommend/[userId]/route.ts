import { NextResponse } from 'next/server'

// Este endpoint REST fue retirado: el frontend ahora usa WebSocket `/ws/recommend/{userId}`.
// Para evitar rupturas, respondemos con 410 Gone indicando que el proxy fue eliminado.

export async function GET() {
  return NextResponse.json({
    error: 'Gone',
    message: 'El proxy REST /api/recommend/{userId} fue eliminado. Use el WebSocket /ws/recommend/{userId}.'
  }, { status: 410 })
}
