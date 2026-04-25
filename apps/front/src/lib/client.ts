import { apiFetcher } from './fetcher'

export async function clientApiFetcher<T>(
  url: string,
  params: URLSearchParams = new URLSearchParams(),
  options: RequestInit = {},
): Promise<T> {
  return apiFetcher<T>(url, params, options)
}
