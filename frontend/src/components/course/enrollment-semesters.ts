export function recentEnrollmentSemesters(semesters: string[]): string[] {
  const years = semesters.map((semester) =>
    Number(semester.match(/^(\d{4})-\d{4}-\d+$/)?.[1])
  )
  const validYears = years.filter(Number.isFinite)
  if (!validYears.length) return semesters
  const earliest = Math.max(...validYears) - 4
  // Keep unrecognized entries selectable by default rather than silently losing them.
  return semesters.filter(
    (_, index) => !Number.isFinite(years[index]) || years[index] >= earliest
  )
}
