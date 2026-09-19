import { adminRequest, AdminError } from '@/lib/admin/server'

const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i
const actions = new Set(['owner', 'give', 'revoke', 'sync'])
const noStore = { 'Cache-Control': 'no-store' }

export async function POST(request: Request, { params }: { params: Promise<{ id: string; action: string }> }) {
  const origin = process.env.ADMIN_ORIGIN
  if (!origin || request.headers.get('origin') !== origin) {
    return Response.json({ error: 'Источник запроса не разрешён.' }, { status: 403, headers: noStore })
  }
  const { id, action } = await params
  if (!uuid.test(id) || !actions.has(action)) {
    return Response.json({ error: 'Неизвестное действие.' }, { status: 400, headers: noStore })
  }
  if (!request.headers.get('content-type')?.startsWith('application/x-www-form-urlencoded')) {
    return Response.json({ error: 'Некорректная форма.' }, { status: 415, headers: noStore })
  }
  // Bound the stream before decoding rather than trusting Content-Length.
  const reader = request.body?.getReader()
  const chunks: Uint8Array[] = []
  let size = 0
  if (reader) {
    while (true) {
      const { done, value } = await reader.read()
      if (done) break
      size += value.byteLength
      if (size > 8192) {
        await reader.cancel()
        return Response.json({ error: 'Форма слишком большая.' }, { status: 413, headers: noStore })
      }
      chunks.push(value)
    }
  }
  const bytes = new Uint8Array(size)
  let offset = 0
  for (const chunk of chunks) { bytes.set(chunk, offset); offset += chunk.byteLength }
  try {
    const response = await adminRequest(`/profiles/${id}/${action}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded', Origin: origin },
      body: new TextDecoder().decode(bytes),
    })
    return Response.json(await response.json(), { status: response.status, headers: noStore })
  } catch (error) {
    return Response.json({ error: error instanceof AdminError ? error.message : 'Не удалось выполнить действие.' }, { status: 503, headers: noStore })
  }
}
