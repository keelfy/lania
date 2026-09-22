import { parseMotd, type RawMotdDescription } from '@/lib/motd'
import { cn } from '@/lib/utils'
import React from 'react'

type Props = {
  description: RawMotdDescription
}

// Renders a Minecraft server MOTD: resolves both the modern chat-component
// format (nested `extra` arrays) and legacy `§`-coded plain strings.
export default function Motd({ description }: Props) {
  const segments = parseMotd(description)

  return (
    <>
      {segments.map((segment, segmentIndex) =>
        segment.text.split('\n').map((line, lineIndex) => (
          <React.Fragment key={`${segmentIndex}-${lineIndex}`}>
            {lineIndex > 0 && <br />}
            <span
              style={segment.color ? { color: segment.color } : undefined}
              className={cn(
                segment.bold && 'font-bold',
                segment.italic && 'italic',
                segment.underlined && 'underline',
                segment.strikethrough && 'line-through',
              )}
            >
              {line}
            </span>
          </React.Fragment>
        )),
      )}
    </>
  )
}
