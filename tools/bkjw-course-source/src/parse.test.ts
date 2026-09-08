import { readFile } from "node:fs/promises"
import { join } from "node:path"

import { describe, expect, it } from "vitest"

import { detailWarnings, normalizeSemester, parseDetailPage, parseListingPage, parseSemesterOptions } from "./parse.js"
import { PageStructureError } from "./types.js"

const fixture = (name: string) => readFile(join(process.cwd(), "test", "fixtures", name), "utf8")

describe("bkjw parsers", () => {
  it("parses semester calendar values instead of visible labels", async () => {
    const options = parseSemesterOptions(await fixture("semester-calendar.html"))
    expect(options).toEqual([
      expect.objectContaining({ id: "222", normalized: "2025-2026-1" }),
      expect.objectContaining({ id: "242", normalized: "2025-2026-2" }),
    ])
    expect(normalizeSemester("2025-2026学年", "第一学期")).toBe("2025-2026-1")
  })

  it("preserves list fields and safe eams links", async () => {
    const page = parseListingPage(await fixture("listing.html"))
    expect(page.expectedCount).toBe(2)
    expect(page.availablePageSizes).toEqual([20, 1000])
    expect(page.lessons[0]).toMatchObject({
      id: "101",
      summary: {
        courseCode: "CS101",
        courseName: "程序设计",
        teachers: ["张三", "李四"],
        actualCount: 38,
        credit: 3,
      },
      links: {
        detail: "/eams/stdSyllabus!info.action?lesson.id=101",
        syllabus: "/eams/stdSyllabus!downloadSyllabus.action?lesson.id=101",
      },
    })
    expect(page.lessons[1]?.summary.actualCount).toBe(0)
    expect(page.lessons[1]?.summary.teachers).toEqual(["王五"])
  })

  it("extracts semantic headers from sortable title attributes", async () => {
    let html = await fixture("listing.html")
    for (const header of ["课程序号", "课程代码", "课程名称", "课程类别", "教学班", "实际", "上限", "学分", "学时/周", "起止周"]) {
      html = html.replace(`<th title="${header}">${header}</th>`, `<th title="点击按 [${header}] 排序">${header}</th>`)
    }
    const page = parseListingPage(html)
    expect(page.lessons).toHaveLength(2)
    expect(page.lessons[0]?.summary.courseName).toBe("程序设计")
  })

  it("resolves relative links inside the eams path", async () => {
    const html = (await fixture("listing.html"))
      .replace("/eams/stdSyllabus!info.action?lesson.id=101", "stdSyllabus!info.action?lesson.id=101")
      .replace("/eams/stdSyllabus!downloadSyllabus.action?lesson.id=101", "stdSyllabus!downloadSyllabus.action?lesson.id=101")
    const page = parseListingPage(html)
    expect(page.lessons[0]?.links.detail).toBe("/eams/stdSyllabus!info.action?lesson.id=101")
    expect(page.lessons[0]?.links.syllabus).toBe("/eams/stdSyllabus!downloadSyllabus.action?lesson.id=101")
  })

  it("uses the selected page size when the final page is shorter", async () => {
    const html = (await fixture("listing.html")).replace(
      "1</strong> - <strong>2</strong> of <strong>2",
      "1001</strong> - <strong>1002</strong> of <strong>1002",
    )
    const page = parseListingPage(html, 1000)
    expect(page.pageNo).toBe(2)
    expect(page.pageSize).toBe(1000)
    expect(page.maxPageNo).toBe(2)
  })

  it("records warnings for missing summary values", async () => {
    const html = (await fixture("listing.html")).replace("<td>通识课程</td>", "<td></td>")
    const page = parseListingPage(html)
    expect(page.lessons[0]?.warnings).toContain("列表页缺少课程类别")
  })

  it("parses detail basic, arrangement and restrictions", async () => {
    const detail = parseDetailPage(await fixture("detail.html"))
    expect(detail.basic).toMatchObject({
      semester: "2025-2026学年1学期",
      courseCode: "CS101",
      department: "信息学院",
      campus: "本部",
      assessment: "考试",
      teachers: ["张三", "李四"],
    })
    expect(detail.arrangement).toMatchObject({ totalHours: 48, weeklyHours: 4, startWeek: 1, endWeek: 12 })
    expect(detail.raw["上限"]).toBe("40")
    expect(detail.restrictions[0]?.items).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ name: "上限", value: "0" }),
        expect.objectContaining({ name: "年级", operator: "包含", value: "2025" }),
      ]),
    )
    expect(detail.remark).toBe("")
    expect(detailWarnings(detail)).toEqual([])
  })

  it("keeps an unsplittable restriction as raw text and reports missing fields", () => {
    const detail = parseDetailPage(`
      <table>
        <tr><td>课程序号:</td><td>CS101.01</td></tr>
        <tr><td>限制条件组:</td><td>仅限大二学生</td></tr>
      </table>
    `)
    expect(detail.restrictions).toEqual([{ title: "原始限制条件", items: [], raw: "仅限大二学生" }])
    expect(detailWarnings(detail)).toEqual([
      "详情页缺少字段：学期、课程代码、课程名称、课程类别、开课院系、校区、授课语言、考核方式、总课时、周课时、起止周、起始周、结束周、周数",
      "限制条件仅保留原始文本，未完全结构化",
    ])
  })

  it("fails closed when the listing table disappears", () => {
    expect(() => parseListingPage("<html><body>登录超时</body></html>")).toThrow(PageStructureError)
  })
})
