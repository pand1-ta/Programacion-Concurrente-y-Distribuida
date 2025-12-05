import { NextResponse } from 'next/server'

const BACKEND = process.env.RECOMMENDER_BACKEND || 'http://localhost:8080'

export async function GET() {
  try {
    const res = await fetch(`${BACKEND}/users`, { cache: 'no-store' })
    const text = await res.text()
    // intentar parsear JSON, si no, devolver texto
    try {
      const json = JSON.parse(text)
      return NextResponse.json(json, { status: res.status })
    } catch {
      return new NextResponse(text, { status: res.status })
    }
  } catch (err: any) {
    return NextResponse.json({ error: 'Backend unreachable', detail: String(err) }, { status: 502 })
  }
}
