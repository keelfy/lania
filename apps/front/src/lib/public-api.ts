import { apiFetcher, ApiFetcher } from './fetcher'

// Reads public data without the visitor's cookies, so a cached result is the same for every visitor.
// The locale names cosmetics in the language of the page.
export function publicApiFetcherFor(locale?: string): ApiFetcher {
  return async <T>(
    url: string,
    params: URLSearchParams = new URLSearchParams(),
    options: RequestInit = {},
  ) => apiFetcher<T>(url, params, options, undefined, locale)
}

export const publicApiFetcher = publicApiFetcherFor()
