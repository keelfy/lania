import { headers } from 'next/headers'
import { apiFetcher } from './fetcher'

export async function serverApiFetcher<T>(
  url: string,
  params: URLSearchParams = new URLSearchParams(),
  options: RequestInit = {},
): Promise<T> {
  const cookies = (await headers()).get('cookie') ?? ''
  return apiFetcher<T>(url, params, options, cookies)
}
