export type ImageSize = 'sm' | 'md' | 'lg'

export type Paginated<T> = {
  content: T[]
  page: number
  size: number
  totalPages: number
  totalElements: number
}

export type CursorPaginated<T> = {
  content: T[]
  totalElements: number
}

// A page of a source that only knows the token of the next page. The token is missing on the last page.
export type TokenPaginated<T> = {
  content: T[]
  nextPageToken?: string
}

export type SearchHit<T> = {
  id: string
  source: T
  score: number
}
