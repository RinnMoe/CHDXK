export type SearchRecord = Record<string, unknown>

export function stringParam(value: unknown): string | undefined {
  if (typeof value !== "string") return undefined
  const trimmed = value.trim()
  return trimmed || undefined
}

export function rawStringParam(value: unknown): string | undefined {
  return typeof value === "string" && value ? value : undefined
}

export function numberParam(value: unknown): number | undefined {
  if (typeof value !== "string" && typeof value !== "number") return undefined
  const number = Number(value)
  return Number.isFinite(number) ? number : undefined
}

export function positiveIntParam(value: unknown, fallback = 1): number {
  const number = numberParam(value)
  if (!number) return fallback
  return Math.max(1, Math.trunc(number))
}

export function optionalPositiveIntParam(value: unknown): number | undefined {
  const number = numberParam(value)
  if (!number) return undefined
  return Math.max(1, Math.trunc(number))
}

export function stringArrayParam(value: unknown): string[] | undefined {
  const values = Array.isArray(value)
    ? value
    : value === undefined
      ? []
      : [value]
  const strings = values
    .filter((item): item is string => typeof item === "string")
    .map((item) => item.trim())
    .filter(Boolean)
  return strings.length > 0 ? strings : undefined
}

export function enumParam<const T extends readonly string[]>(
  value: unknown,
  options: T
): T[number] | undefined {
  return typeof value === "string" && options.includes(value)
    ? (value as T[number])
    : undefined
}
