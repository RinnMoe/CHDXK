// Interpret the API's ISO week key, rather than the browser's local date.
export function formatHotPeriodDescription(
  period: "week" | "month",
  periodKey?: string
): string {
  const match = /^(\d{4})-(\d{2})$/.exec(periodKey ?? "")
  if (!match) return ""
  const year = Number(match[1])
  const value = Number(match[2])
  if (year < 1000 || value < 1) return ""
  if (period === "month") return value <= 12 ? `${year}年${value}月` : ""
  const monday = (isoYear: number) => {
    const day = new Date(Date.UTC(isoYear, 0, 4))
    day.setUTCDate(day.getUTCDate() - ((day.getUTCDay() + 6) % 7))
    return day
  }
  const start = monday(year)
  start.setUTCDate(start.getUTCDate() + (value - 1) * 7)
  if (start >= monday(year + 1)) return ""
  const end = new Date(start)
  end.setUTCDate(end.getUTCDate() + 6)
  const date = (day: Date, includeYear: boolean) =>
    `${includeYear ? `${day.getUTCFullYear()}年` : ""}${day.getUTCMonth() + 1}月${day.getUTCDate()}日`
  return `${date(start, true)}—${date(end, start.getUTCFullYear() !== end.getUTCFullYear())}`
}
