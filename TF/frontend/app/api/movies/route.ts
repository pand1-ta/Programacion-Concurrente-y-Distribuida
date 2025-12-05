import { NextResponse } from 'next/server'

const BACKEND = process.env.RECOMMENDER_BACKEND || 'http://localhost:8080'

export async function GET(req: Request) {
  try {
    const url = new URL(req.url)
    const target = new URL(`${BACKEND}/movies`)
    target.search = url.search

    const res = await fetch(target.toString(), { cache: 'no-store' })
    const text = await res.text()
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
