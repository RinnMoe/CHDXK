import { describe, expect, it } from "vitest"
import { formatHotPeriodDescription } from "./hot-period"

describe("hot ranking period labels", () => {
  it("shows the actual Monday to Sunday for the reported week", () => {
    expect(formatHotPeriodDescription("week", "2026-41")).toBe(
      "2026年10月5日—10月11日"
    )
  })
  it("handles ISO weeks spanning calendar years", () => {
    expect(formatHotPeriodDescription("week", "2025-01")).toBe(
      "2024年12月30日—2025年1月5日"
    )
    expect(formatHotPeriodDescription("week", "2020-53")).toBe(
      "2020年12月28日—2021年1月3日"
    )
  })
  it("formats months and rejects invalid keys", () => {
    expect(formatHotPeriodDescription("month", "2026-10")).toBe("2026年10月")
    for (const key of [undefined, "2026-00", "2026-54", "2025-53", "bad"])
      expect(formatHotPeriodDescription("week", key)).toBe("")
    expect(formatHotPeriodDescription("month", "2026-13")).toBe("")
  })
})
