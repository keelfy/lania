const baseURL = process.env.NEXT_PUBLIC_BACKEND_URL

export type ApiFetcher = <T>(
  path: string,
  params?: URLSearchParams,
  options?: RequestInit,
) => Promise<T>

export async function apiFetcher<T>(
  path: string,
  params: URLSearchParams,
  options: RequestInit = {},
  cookies?: string,
  locale?: string,
): Promise<T> {
  try {
    const headers = new Headers(options.headers)
    if (cookies) {
      headers.set('Cookie', cookies)
    }
    // The API names cosmetics in this language. An explicit ?locale= still wins.
    if (locale && !headers.has('Accept-Language')) {
      headers.set('Accept-Language', locale)
    }

    const url = new URL(baseURL + path)
    url.search = params.toString()
    const response = await fetch(url, {
      ...options,
      headers,
      credentials: 'include',
    })
    const contentType = response.headers.get('Content-Type')

    let body: object | string | Blob | undefined = undefined

    if (contentType?.includes('application/json')) {
      body = await response.json()
    } else if (contentType?.includes('application/zip')) {
      body = await response.blob()
    } else {
      body = await response.text()
    }

    if (response.status >= 400) {
      if (body && typeof body === 'object' && 'error' in body) {
        throw new Error(JSON.stringify(body.error))
      } else if (body && typeof body === 'string' && body.length > 0) {
        throw new Error(body)
      }
      throw new Error(`Status code: ${response.status}`)
    }

    return body as T
  } catch (error) {
    console.error('Error fetching data: ', error)
    throw error
  }
}
