type DateInput = Date | string | number

const dateOnlyPattern = /^\d{4}-\d{2}-\d{2}$/

function toValidDate(value: DateInput): Date | null {
  const date = value instanceof Date ? value : new Date(value)
  return Number.isNaN(date.getTime()) ? null : date
}

function pad2(value: number): string {
  return String(value).padStart(2, "0")
}

export function formatDateInputValue(value: DateInput): string {
  if (typeof value === "string" && dateOnlyPattern.test(value)) return value
  const date = toValidDate(value)
  if (!date) return String(value)
  return `${date.getFullYear()}-${pad2(date.getMonth() + 1)}-${pad2(
    date.getDate()
  )}`
}

export function formatDateTime(value: DateInput): string {
  const date = toValidDate(value)
  if (!date) return String(value)
  return `${formatDateInputValue(date)} ${pad2(date.getHours())}:${pad2(
    date.getMinutes()
  )}`
}

export function formatNullableDateTime(
  value?: DateInput | null,
  fallback = ""
): string {
  return value == null || value === "" ? fallback : formatDateTime(value)
}

export function addDays(value: DateInput, days: number): Date {
  const date = toValidDate(value)
  const next = date ? new Date(date) : new Date()
  next.setDate(next.getDate() + days)
  return next
}

export function formatRelativeDateInputValue(daysFromToday: number): string {
  return formatDateInputValue(addDays(new Date(), daysFromToday))
}
