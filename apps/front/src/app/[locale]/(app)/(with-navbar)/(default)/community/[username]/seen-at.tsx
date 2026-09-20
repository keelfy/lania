'use client'

import { useTimeAgo } from 'next-timeago'

type Props = {
  // Unix epoch milliseconds. Missing means the player was never seen.
  date?: number
  locale: string
}

export default function SeenAt({ date, locale }: Props) {
  const { TimeAgo } = useTimeAgo()
  if (date === undefined) return <>&mdash;</>
  return (
    <span suppressHydrationWarning>
      <TimeAgo date={date} locale={locale} />
    </span>
  )
}
