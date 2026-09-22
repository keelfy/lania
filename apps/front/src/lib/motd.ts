// `@types/minecraft-pinger` models a server's MOTD `description` as a
// single object with a *singular* nested `extra`. The real Minecraft
// protocol sends a full chat-component tree: `description` can itself be a
// legacy `§`-coded plain string, and `extra` is an array that can nest to
// any depth. This module works off that real shape instead of the
// inaccurate package types.

export type ChatComponent = {
  text?: string
  color?: string
  bold?: boolean
  italic?: boolean
  underlined?: boolean
  strikethrough?: boolean
  obfuscated?: boolean
  extra?: ChatComponent[]
}

export type RawMotdDescription = string | ChatComponent

export type MotdSegment = {
  text: string
  color?: string
  bold?: boolean
  italic?: boolean
  underlined?: boolean
  strikethrough?: boolean
}

const LEGACY_COLOR_CODES: Record<string, string> = {
  '0': '#000000',
  '1': '#0000AA',
  '2': '#00AA00',
  '3': '#00AAAA',
  '4': '#AA0000',
  '5': '#AA00AA',
  '6': '#FFAA00',
  '7': '#AAAAAA',
  '8': '#555555',
  '9': '#5555FF',
  a: '#55FF55',
  b: '#55FFFF',
  c: '#FF5555',
  d: '#FF55FF',
  e: '#FFFF55',
  f: '#FFFFFF',
}

// Splits a string on legacy `§` formatting codes into styled segments,
// starting from the formatting inherited from the parent chat component.
function parseLegacyText(raw: string, inherited: MotdSegment): MotdSegment[] {
  const segments: MotdSegment[] = []
  let current: MotdSegment = { ...inherited, text: '' }

  for (let i = 0; i < raw.length; i++) {
    const char = raw[i]

    if (char !== '§' || i + 1 >= raw.length) {
      current.text += char
      continue
    }

    const code = raw[i + 1].toLowerCase()
    i++

    if (current.text) segments.push(current)

    if (code in LEGACY_COLOR_CODES) {
      current = { ...inherited, text: '', color: LEGACY_COLOR_CODES[code] }
    } else if (code === 'l') {
      current = { ...current, text: '', bold: true }
    } else if (code === 'o') {
      current = { ...current, text: '', italic: true }
    } else if (code === 'n') {
      current = { ...current, text: '', underlined: true }
    } else if (code === 'm') {
      current = { ...current, text: '', strikethrough: true }
    } else if (code === 'r') {
      current = { ...inherited, text: '' }
    } else {
      // Unknown or unsupported code (e.g. obfuscated `k`): drop the code,
      // keep the text flowing with the current formatting.
      current = { ...current, text: '' }
    }
  }

  if (current.text) segments.push(current)

  return segments
}

// Flattens a chat-component tree (including nested `extra` arrays) into a
// flat, styled segment list. A child inherits formatting from its parent
// unless it overrides it.
function flattenComponent(
  component: ChatComponent,
  inherited: MotdSegment,
): MotdSegment[] {
  const own: MotdSegment = {
    text: '',
    color: component.color ?? inherited.color,
    bold: component.bold ?? inherited.bold,
    italic: component.italic ?? inherited.italic,
    underlined: component.underlined ?? inherited.underlined,
    strikethrough: component.strikethrough ?? inherited.strikethrough,
  }

  const segments = component.text ? parseLegacyText(component.text, own) : []

  for (const child of component.extra ?? []) {
    segments.push(...flattenComponent(child, own))
  }

  return segments
}

export function parseMotd(
  description: RawMotdDescription | undefined,
): MotdSegment[] {
  if (!description) return []

  if (typeof description === 'string') {
    return parseLegacyText(description, { text: '' })
  }

  return flattenComponent(description, { text: '' })
}
