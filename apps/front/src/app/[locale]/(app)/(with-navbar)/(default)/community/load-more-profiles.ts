'use server'

import { getProfilesFeed } from '@/lib/api-endpoints'
import { serverApiFetcher } from '@/lib/server'
import { PublicProfile } from '@/models/profile'

export type CommunityFeedQuery = {
  // The column and the direction, like in the URL.
  sort: string
  search: string
  online: boolean
  staff: boolean
  seasonId?: string
}

export async function loadMoreProfiles(
  query: CommunityFeedQuery,
  cursor: string,
): Promise<{ profiles: PublicProfile[]; nextCursor?: string }> {
  const [col, dir] = query.sort.split('.')
  const page = await getProfilesFeed(serverApiFetcher, {
    col,
    dir,
    search: query.search,
    onlineOnly: query.online,
    staffOnly: query.staff,
    seasonId: query.seasonId,
    cursor,
  })
  return { profiles: page.content, nextCursor: page.nextCursor }
}
