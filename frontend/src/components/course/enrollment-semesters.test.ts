import { expect, it } from "vitest"
import { recentEnrollmentSemesters } from "./enrollment-semesters"

it("按五个学年选择，保留春夏学期且不受列表顺序影响", () => {
  expect(
    recentEnrollmentSemesters([
      "2015-2016-1",
      "2026-2027-1",
      "2021-2022-2",
      "2022-2023-1",
      "2022-2023-2",
      "2022-2023-3",
      "2025-2026-2",
    ])
  ).toEqual([
    "2026-2027-1",
    "2022-2023-1",
    "2022-2023-2",
    "2022-2023-3",
    "2025-2026-2",
  ])
})

it("历史账号以可用目录的最新学年计算，不强依赖当前日期", () => {
  expect(
    recentEnrollmentSemesters(["2009-2010-1", "2010-2011-1", "2014-2015-1"])
  ).toEqual(["2010-2011-1", "2014-2015-1"])
})

it("空目录和无法识别的学期不误删", () => {
  expect(recentEnrollmentSemesters([])).toEqual([])
  expect(recentEnrollmentSemesters(["特殊学期"])).toEqual(["特殊学期"])
})
