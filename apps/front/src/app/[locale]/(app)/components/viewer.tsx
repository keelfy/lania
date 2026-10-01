import { getBasket } from '@/lib/api-endpoints'
import { getCurrentSession } from '@/lib/get-current-session'
import { serverApiFetcher } from '@/lib/server'
import ViewerSync from '@/providers/viewer-sync'

// The session and the basket belong to the request. They are read here, behind a Suspense boundary of their own,
// so the rest of the page does not wait for them and can be prerendered.
export default async function Viewer() {
  const session = await getCurrentSession()
  // Only a signed in user has a basket.
  const basket = session?.active
    ? await getBasket(serverApiFetcher).catch((err) => {
        console.error(err)
        return []
      })
    : []
  return <ViewerSync session={session} basket={basket} />
}
