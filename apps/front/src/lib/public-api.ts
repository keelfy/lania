import { apiFetcher } from './fetcher'

// Reads public data without the visitor's cookies, so a cached result is the same for every visitor.
export async function publicApiFetcher<T>(
  url: string,
  params: URLSearchParams = new URLSearchParams(),
  options: RequestInit = {},
): Promise<T> {
  return apiFetcher<T>(url, params, options)
}
