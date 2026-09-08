import { parseHTML } from "linkedom"

import {
  type LessonDetail,
  type LessonRecord,
  type ListingPage,
  PageStructureError,
  type RestrictionGroup,
  type SemesterOption,
} from "./types.js"

const YEAR_PATTERN = /(\d{4}-\d{4})/
const TERM_PATTERN = /(?:学期|第)?\s*([1-9]\d*)/

function cleanText(value: string | null | undefined): string {
  return (value ?? "").replace(/\s+/g, " ").trim()
}

function normalizeLabel(value: string): string {
  return cleanText(value).replace(/[：:]+$/, "")
}

function normalizeHeader(cell: Element): string {
  const candidates = [cell.getAttribute("title") ?? "", cell.textContent ?? ""]
  for (const candidate of candidates) {
    const text = cleanText(candidate)
    if (!text) continue
    const sorted = text.match(/\[\s*([^\]]+?)\s*\]/)?.[1]
    const label = normalizeLabel(sorted ?? text)
    if (label && !label.startsWith("点击按")) return label
  }
  return ""
}

function normalizeAcademicYear(value: string): string {
  const match = YEAR_PATTERN.exec(value)
  return match?.[1] ?? cleanText(value)
}

export function normalizeTerm(value: string): string {
  const text = cleanText(value)
  if (/^\d{4}-\d{4}(?:学年)?$/.test(text)) return ""
  const labeled = text.match(/(?:学期|第)\s*([1-9]\d*|[一二三四五六七八九十]+)/)
  if (labeled?.[1]) return chineseOrArabicTerm(labeled[1])
  const afterAcademicYear = text.match(/\d{4}-\d{4}.*?([1-9]\d*|[一二三四五六七八九十]+)/)
  if (afterAcademicYear?.[1]) return chineseOrArabicTerm(afterAcademicYear[1])
  const match = TERM_PATTERN.exec(text)
  return match?.[1] ?? text
}

function chineseOrArabicTerm(value: string): string {
  if (/^\d+$/.test(value)) return value
  const digits: Record<string, number> = { 一: 1, 二: 2, 三: 3, 四: 4, 五: 5, 六: 6, 七: 7, 八: 8, 九: 9 }
  if (value === "十") return "10"
  const first = digits[value[0] ?? ""]
  const second = digits[value[1] ?? ""]
  if (value.length === 2 && value[0] === "十" && second !== undefined) return String(10 + second)
  if (value.length === 2 && value[1] === "十" && first !== undefined) return String(first * 10)
  return String(digits[value] ?? value)
}

export function normalizeSemester(academicYear: string, term: string): string {
  return `${normalizeAcademicYear(academicYear)}-${normalizeTerm(term)}`
}

export function parseSemesterOptions(html: string): SemesterOption[] {
  const { document } = parseHTML(html)
  const explicit = Array.from(document.querySelectorAll("[data-semester-id]"))
    .map((element) => {
      const id = cleanText(element.getAttribute("data-semester-id"))
      const academicYear = normalizeAcademicYear(
        element.getAttribute("data-academic-year") ?? element.textContent ?? "",
      )
      const term = normalizeTerm(element.getAttribute("data-term") ?? element.textContent ?? "")
      if (!id || !YEAR_PATTERN.test(academicYear) || !term) return null
      return {
        id,
        academicYear,
        term,
        label: cleanText(element.textContent),
        normalized: normalizeSemester(academicYear, term),
      } satisfies SemesterOption
    })
    .filter((option): option is SemesterOption => option !== null)

  if (explicit.length > 0) return dedupeSemesters(explicit)

  const selectedYear =
    cleanText(document.querySelector("#semesterCalendar_year")?.getAttribute("value")) ||
    cleanText(document.querySelector("#semesterCalendar_year")?.textContent) ||
    cleanText(document.querySelector("#semesterCalendar_yearTb td.ui-state-active")?.textContent) ||
    cleanText(document.querySelector("#semesterCalendar_yearTb td")?.textContent)
  const academicYear = normalizeAcademicYear(selectedYear)
  const terms = Array.from(document.querySelectorAll("#semesterCalendar_termTb td"))
  return dedupeSemesters(
    terms
      .map((element) => {
        const id = cleanText(element.getAttribute("val") ?? element.getAttribute("data-value"))
        const term = normalizeTerm(element.textContent ?? "")
        if (!id || !YEAR_PATTERN.test(academicYear) || !term) return null
        return {
          id,
          academicYear,
          term,
          label: `${academicYear}学年${term}学期`,
          normalized: normalizeSemester(academicYear, term),
        } satisfies SemesterOption
      })
      .filter((option): option is SemesterOption => option !== null),
  )
}

function dedupeSemesters(options: SemesterOption[]): SemesterOption[] {
  const result: SemesterOption[] = []
  const seen = new Set<string>()
  for (const option of options) {
    const key = `${option.id}|${option.normalized}`
    if (seen.has(key)) continue
    seen.add(key)
    result.push(option)
  }
  return result
}

function parseInteger(value: string): number | null {
  const match = cleanText(value).replace(/,/g, "").match(/-?\d+/)
  return match ? Number.parseInt(match[0], 10) : null
}

function parseNumber(value: string): number | null {
  const match = cleanText(value).replace(/,/g, "").match(/-?(?:\d+\.?\d*|\.\d+)/)
  return match ? Number.parseFloat(match[0]) : null
}

function splitPeople(value: string): string[] {
  const text = cleanText(value)
  if (!text || text === "0") return []
  return text
    .split(/[,，;；、/\s]+/)
    .map((name) => name.trim())
    .filter(Boolean)
}

function safeEamsPath(value: string): string {
  const raw = cleanText(value)
  if (!raw) return ""
  try {
    const url = new URL(raw, "http://bkjw.chd.edu.cn/eams/")
    if ((url.protocol !== "http:" && url.protocol !== "https:") || url.hostname !== "bkjw.chd.edu.cn" || !url.pathname.startsWith("/eams/")) return ""
    return `${url.pathname}${url.search}${url.hash}`
  } catch {
    return ""
  }
}

function parsePageRange(
  text: string,
  pageSizeHint?: number,
): { start: number; end: number; pageNo: number; pageSize: number; expectedCount: number } {
  const match = cleanText(text).match(/(\d+)\s*-\s*(\d+)\s+of\s+(\d+)/)
  if (!match) return { start: 1, end: 0, pageNo: 1, pageSize: pageSizeHint ?? 0, expectedCount: 0 }
  const start = Number.parseInt(match[1] ?? "1", 10)
  const end = Number.parseInt(match[2] ?? "0", 10)
  const expectedCount = Number.parseInt(match[3] ?? "0", 10)
  const observedPageSize = end >= start ? end - start + 1 : 0
  const pageSize = pageSizeHint && pageSizeHint > 0 ? pageSizeHint : observedPageSize
  const pageNo = pageSize > 0 ? Math.floor((start - 1) / pageSize) + 1 : 1
  return { start, end, pageNo, pageSize, expectedCount }
}

export function parseListingPage(html: string, pageSizeHint?: number): ListingPage {
  const { document } = parseHTML(html)
  const grid = document.querySelector("table.gridtable")
  if (!grid) throw new PageStructureError("列表页缺少课程网格表格（table.gridtable）")

  const headerRow = Array.from(grid.querySelectorAll("thead tr"))
    .reverse()
    .find((row) => row.querySelectorAll("th").length >= 5)
    ?? Array.from(grid.querySelectorAll("tr"))
      .reverse()
      .find((row) => row.querySelectorAll("th").length >= 5)
  if (!headerRow) throw new PageStructureError("列表页缺少课程表头")

  const headers = Array.from(headerRow.querySelectorAll("th")).map((cell) => normalizeHeader(cell))
  const requiredHeaders = ["课程序号", "课程代码", "课程名称", "课程类别", "教学班", "教师"]
  const missingHeaders = requiredHeaders.filter((header) => !headers.includes(header))
  if (missingHeaders.length > 0) {
    throw new PageStructureError(`列表页缺少必要表头：${missingHeaders.join("、")}`)
  }

  const indexOf = (header: string): number => headers.indexOf(header)
  const bodyRows = Array.from(grid.querySelectorAll("tbody tr"))
  if (bodyRows.length === 0) {
    bodyRows.push(
      ...Array.from(grid.querySelectorAll("tr")).filter(
        (row) => row !== headerRow && row.querySelector('input.box[name="lesson.id"]') !== null,
      ),
    )
  }
  const lessons: LessonRecord[] = []
  for (const row of bodyRows) {
    const id = cleanText(row.querySelector('input.box[name="lesson.id"]')?.getAttribute("value"))
    if (!id) continue
    const cells = Array.from(row.children).filter((element) => element.tagName.toLowerCase() === "td")
    const values = cells.map((cell) => cleanText(cell.textContent))
    const valueAt = (header: string): string => values[indexOf(header)] ?? ""
    const detailLink = safeEamsPath(row.querySelector('a[href*="stdSyllabus!info.action"]')?.getAttribute("href") ?? "")
    const syllabusLink = safeEamsPath(
      row.querySelector('a[href*="stdSyllabus!downloadSyllabus.action"]')?.getAttribute("href") ?? "",
    )
    if (!detailLink) throw new PageStructureError(`课程 ${id} 缺少详情页链接`)
    const warnings: string[] = []
    if (!syllabusLink) warnings.push("列表页缺少课程大纲链接")
    for (const header of [
      "课程序号",
      "课程代码",
      "课程名称",
      "课程类别",
      "教学班",
      "教师",
      "实际",
      "上限",
      "学分",
      "学时/周",
      "起止周",
    ]) {
      if (!valueAt(header)) warnings.push(`列表页缺少${header}`)
    }

    const raw: Record<string, string> = {}
    for (const [index, header] of headers.entries()) {
      if (header) raw[header] = values[index] ?? ""
    }
    lessons.push({
      id,
      summary: {
        lessonNo: valueAt("课程序号"),
        courseCode: valueAt("课程代码"),
        courseName: valueAt("课程名称"),
        courseType: valueAt("课程类别"),
        teachClass: valueAt("教学班"),
        teachers: splitPeople(valueAt("教师")),
        actualCount: parseInteger(valueAt("实际")),
        limitCount: parseInteger(valueAt("上限")),
        credit: parseNumber(valueAt("学分")),
        coursePeriod: valueAt("学时/周"),
        weekRange: valueAt("起止周"),
        raw,
      },
      detail: null,
      links: { detail: detailLink, syllabus: syllabusLink },
      warnings,
    })
  }

  const rangeText = Array.from(grid.parentElement?.querySelectorAll(".girdbar-pgbar, .gridbar-pgbar") ?? [])
    .map((element) => element.textContent ?? "")
    .find((text) => /of\s+\d+/.test(text)) ?? ""
  if (!rangeText && lessons.length > 0) throw new PageStructureError("列表页缺少分页总数")
  const range = parsePageRange(rangeText, pageSizeHint)
  const pageSize = range.pageSize || lessons.length
  const expectedCount = rangeText ? range.expectedCount : lessons.length
  const maxPageNo = pageSize > 0 ? Math.max(1, Math.ceil(expectedCount / pageSize)) : 1
  const availablePageSizes = Array.from(document.querySelectorAll('select[id$="_page_select"] option'))
    .map((option) => Number.parseInt(option.getAttribute("value") ?? "", 10))
    .filter((value) => Number.isFinite(value) && value > 0)
    .filter((value, index, values) => values.indexOf(value) === index)

  return {
    pageNo: range.pageNo,
    pageSize,
    maxPageNo,
    expectedCount,
    gridSignature: headers.join("|") || grid.className,
    lessons,
    availablePageSizes,
  }
}

function directCells(row: Element): Element[] {
  return Array.from(row.children).filter((element) => ["td", "th"].includes(element.tagName.toLowerCase()))
}

function collectDetailFields(document: Document): { fields: Record<string, string>; restrictionRow: Element | undefined } {
  const fields: Record<string, string> = {}
  let restrictionRow: Element | undefined
  for (const row of Array.from(document.querySelectorAll("tr"))) {
    if (row.parentElement?.closest("td")) continue
    const cells = directCells(row)
    if (cells.length === 0) continue
    const firstLabel = normalizeLabel(cells[0]?.textContent ?? "")
    if (firstLabel === "限制条件组") restrictionRow = row
    for (let index = 0; index < cells.length; index += 2) {
      const label = normalizeLabel(cells[index]?.textContent ?? "")
      if (!label) continue
      const value = cleanText(cells[index + 1]?.textContent)
      if (!(label in fields) || !fields[label]) fields[label] = value
    }
  }
  return { fields, restrictionRow }
}

function parseRestrictions(row: Element | undefined): RestrictionGroup[] {
  if (!row) return []
  const groups = new Map<string, RestrictionGroup>()
  const nestedRows = Array.from(row.querySelectorAll("table tr"))
  const direct = directCells(row)
  const rawValue = cleanText(direct[1]?.textContent)
  if (nestedRows.length === 0) {
    if (!rawValue || /^(无|无要求|不限|[-—])$/.test(rawValue)) return []
    return [{ title: "原始限制条件", items: [], raw: rawValue }]
  }
  let currentTitle = "默认限制组"
  for (const nestedRow of nestedRows) {
    const cells = directCells(nestedRow).map((cell) => cleanText(cell.textContent))
    if (cells.length === 0) continue
    if (cells.length === 1) {
      currentTitle = cells[0] || currentTitle
      continue
    }
    const rowRaw = cells.join(" ")
    if (cells.length >= 4) {
      const title = cells[0] || currentTitle
      currentTitle = title
      addRestrictionItem(groups, title, cells[1] ?? "", cells[2] ?? "", cells[3] ?? "", rowRaw)
      continue
    }
    if (cells.length === 3) {
      if (isRestrictionOperator(cells[1] ?? "")) {
        addRestrictionItem(groups, currentTitle, cells[0] ?? "", cells[1] ?? "", cells[2] ?? "", rowRaw)
      } else {
        currentTitle = cells[0] || currentTitle
        addRestrictionItem(groups, currentTitle, cells[1] ?? "", "", cells[2] ?? "", rowRaw)
      }
      continue
    }
    const left = cells[0] ?? ""
    const value = cells[1] ?? ""
    const match = left.match(/^(.+?)\s+(包含|不包含|等于|不等于|属于|不属于)$/)
    addRestrictionItem(groups, currentTitle, match?.[1]?.trim() ?? left, match?.[2] ?? "", value, rowRaw)
  }
  const raw = cleanText(row.textContent)
  if (groups.size === 0 && rawValue) return [{ title: "原始限制条件", items: [], raw: rawValue }]
  return Array.from(groups.values()).map((group) => ({ ...group, raw }))
}

function addRestrictionItem(
  groups: Map<string, RestrictionGroup>,
  title: string,
  name: string,
  operator: string,
  value: string,
  raw: string,
): void {
  const group = groups.get(title) ?? { title, items: [], raw: "" }
  group.items.push({ name, operator, value, raw })
  groups.set(title, group)
}

function isRestrictionOperator(value: string): boolean {
  return /^(包含|不包含|等于|不等于|属于|不属于)$/.test(value)
}

export function detailWarnings(detail: LessonDetail): string[] {
  const requiredLabels = [
    "学期",
    "课程代码",
    "课程名称",
    "课程类别",
    "开课院系",
    "校区",
    "授课语言",
    "考核方式",
    "总课时",
    "周课时",
    "起止周",
    "起始周",
    "结束周",
    "周数",
  ]
  const missing = requiredLabels.filter((label) => !(label in detail.raw))
  const warnings = missing.length > 0 ? [`详情页缺少字段：${missing.join("、")}`] : []
  if (detail.restrictions.some((group) => group.title === "原始限制条件" && group.items.length === 0)) {
    warnings.push("限制条件仅保留原始文本，未完全结构化")
  }
  return warnings
}

export function parseDetailPage(html: string): LessonDetail {
  const { document } = parseHTML(html)
  const { fields, restrictionRow } = collectDetailFields(document)
  if (Object.keys(fields).length === 0 || (!fields["课程代码"] && !fields["课程序号"])) {
    throw new PageStructureError("详情页没有识别到课程基本信息")
  }

  const raw = { ...fields }
  const detail: LessonDetail = {
    basic: {
      semester: fields["学期"] ?? "",
      courseCode: fields["课程代码"] ?? "",
      courseName: fields["课程名称"] ?? "",
      credit: parseNumber(fields["学分"] ?? ""),
      courseType: fields["课程类别"] ?? "",
      department: fields["开课院系"] ?? "",
      campus: fields["校区"] ?? "",
      language: fields["授课语言"] ?? "",
      assessment: fields["考核方式"] ?? "",
      teachers: splitPeople(fields["教师"] ?? ""),
    },
    arrangement: {
      totalHours: parseNumber(fields["总课时"] ?? ""),
      weeklyHours: parseNumber(fields["周课时"] ?? ""),
      weekRange: fields["起止周"] ?? "",
      startWeek: parseInteger(fields["起始周"] ?? ""),
      endWeek: parseInteger(fields["结束周"] ?? ""),
      weekCount: parseInteger(fields["周数"] ?? ""),
    },
    restrictions: parseRestrictions(restrictionRow),
    remark: fields["备注"] ?? "",
    raw,
  }
  return detail
}
