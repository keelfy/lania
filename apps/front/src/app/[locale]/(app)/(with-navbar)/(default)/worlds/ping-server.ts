import pinger from 'minecraft-pinger'
import { unstable_cache } from 'next/cache'

// Status of a season server for its card; undefined when it does not answer within a second.
export const pingServer = unstable_cache(
  async (publicAddress: string): Promise<pinger.Data | undefined> => {
    let timeoutId: ReturnType<typeof setTimeout> | undefined
    try {
      return await Promise.race([
        pinger.pingPromise(publicAddress, 25565),
        new Promise<undefined>((_, reject) => {
          timeoutId = setTimeout(reject, 1000)
        }),
      ])
    } catch (error) {
      console.error(error)
      return undefined
    } finally {
      clearTimeout(timeoutId)
    }
  },
  ['worlds-server-ping'],
  { revalidate: 15 },
)
