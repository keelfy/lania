import { Session } from '@ory/client-fetch'
import { notFound } from 'next/navigation'
import { getCurrentSession } from './get-current-session'

const ADMIN_ROLES = ['owner', 'admin']

// The site role is kept in metadata_public of the Ory identity. The API checks it again on every admin request.
export function isAdminSession(session: Session | undefined): boolean {
  if (session?.active !== true) return false
  const metadata = session.identity?.metadata_public as
    | { role?: unknown }
    | null
    | undefined
  return (
    typeof metadata?.role === 'string' && ADMIN_ROLES.includes(metadata.role)
  )
}

// Layouts do not render again on navigation, so every admin page has to call this itself.
// Anyone else gets a 404, so the section does not show that it exists.
export async function requireAdmin(): Promise<Session> {
  const session = await getCurrentSession()
  if (!session || !isAdminSession(session)) notFound()
  return session
}
