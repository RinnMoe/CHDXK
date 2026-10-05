import directory from "@/config/copy-catalog.json"

export interface CopyEntry {
  text: string
  audience: string
  location: string
  editable?: boolean
  label?: string
}
export const copyDirectory: Record<string, CopyEntry> = directory
let overrides: Record<string, string> = {}

export function setCopyOverrides(values: Record<string, string>) {
  overrides = values
}

export function copy(key: string, values: Record<string, unknown> = {}) {
  const entry = copyDirectory[key]
  const template =
    (entry?.editable ? overrides[key] : undefined) ?? entry?.text ?? key
  return template.replace(/\{(v\d+)\}/g, (_, name: string) => {
    const value = values[name]
    return typeof value === "string" ||
      typeof value === "number" ||
      typeof value === "boolean"
      ? String(value)
      : ""
  })
}

export function translateMessage(message: string) {
  const key = Object.keys(copyDirectory).find(
    (key) =>
      key.startsWith("copy.server.") && copyDirectory[key].text === message
  )
  return key ? copy(key) : message
}

export async function loadSiteCopy() {
  try {
    const response = await fetch("/api/site-copy", {
      credentials: "include",
      cache: "no-store",
      signal: AbortSignal.timeout(5000),
    })
    if (response.ok) setCopyOverrides(await response.json())
  } catch {
    // Built-in defaults keep login and navigation usable while the API is unavailable.
  }
}
