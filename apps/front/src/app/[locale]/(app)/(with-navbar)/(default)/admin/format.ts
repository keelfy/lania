export function formatDate(millis: number | undefined, locale: string) {
  if (millis === undefined) return '—'
  return new Date(millis).toLocaleDateString(locale, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

export function formatDateTime(millis: number | undefined, locale: string) {
  if (millis === undefined) return '—'
  return new Date(millis).toLocaleString(locale, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
