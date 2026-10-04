import { copy } from "@/lib/copy"
import dayjs from "dayjs"

type DateInput = Date | string | number

const dateOnlyPattern = /^\d{4}-\d{2}-\d{2}$/

function toValidDay(value: DateInput): dayjs.Dayjs | null {
  const date = dayjs(value)
  return date.isValid() ? date : null
}

export function formatDateInputValue(value: DateInput): string {
  if (typeof value === "string" && dateOnlyPattern.test(value)) return value
  const date = toValidDay(value)
  if (!date) return String(value)
  return date.format("YYYY-MM-DD")
}

export function formatDateTime(value: DateInput): string {
  const date = toValidDay(value)
  if (!date) return String(value)
  return date.format("YYYY-MM-DD HH:mm")
}

export function formatReviewCardTime(value: DateInput): string {
  const date = toValidDay(value)
  if (!date) return String(value)

  const now = dayjs()
  const diffMinutes = now.diff(date, "minute")
  if (diffMinutes < 1) {
    return copy("copy.user.lib.date.de6785d99e02")
  }

  if (diffMinutes >= 1 && diffMinutes < 60) {
    return copy("copy.user.lib.date.dc0df49811cc", { v0: diffMinutes })
  }

  const diffHours = now.diff(date, "hour")
  if (diffHours >= 1 && diffHours < 24) {
    return copy("copy.user.lib.date.0696e14781c8", { v0: diffHours })
  }

  const diffDays = now.diff(date, "day")
  if (diffDays >= 1 && diffDays < 7) {
    return copy("copy.user.lib.date.9d6464d5cf6c", { v0: diffDays })
  }

  return formatDateTime(value)
}

export function isReviewCardTimeRelative(value: DateInput): boolean {
  const date = toValidDay(value)
  if (!date) return false

  const now = dayjs()
  const diffMinutes = now.diff(date, "minute")
  return diffMinutes >= 0 && diffMinutes < 7 * 24 * 60
}

export function formatNullableDateTime(
  value?: DateInput | null,
  fallback = ""
): string {
  return value == null || value === "" ? fallback : formatDateTime(value)
}

export function addDays(value: DateInput, days: number): Date {
  return (toValidDay(value) ?? dayjs()).add(days, "day").toDate()
}

export function formatRelativeDateInputValue(daysFromToday: number): string {
  return formatDateInputValue(addDays(new Date(), daysFromToday))
}
