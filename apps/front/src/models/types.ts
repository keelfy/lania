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

export type SearchHit<T> = {
  id: string
  source: T
  score: number
}
