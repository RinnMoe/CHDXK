import type { Page } from "playwright"
import { BkjwClient, BJW_ORIGIN, discoverPageSemesters } from "./client.js"
import type { SemesterOption } from "./types.js"

export interface PersonalCourse {
  lesson_no: string
  source_lesson_id: string
  name: string
}

// A class can appear several times in table0.activities (different days/weeks).
// Keep its complete teaching-section identity instead of picking a teacher.
export function personalCourses(value: unknown): PersonalCourse[] {
  if (!Array.isArray(value) || value.length > 10000) throw new Error("PAGE_STRUCTURE")
  const courses = new Map<string, PersonalCourse>()
  for (const item of value) {
    if (!item || typeof item !== "object") throw new Error("PAGE_STRUCTURE")
    const raw = item as Record<string, unknown>
    const name = String(raw.courseName ?? "").trim()
    const lesson = String(raw.lessonNo ?? "").trim()
    const id = /^(\d+)(?:\([^)]*\))?$/.exec(String(raw.courseId ?? ""))?.[1] ?? ""
    if (!name || (!lesson && !id)) throw new Error("PAGE_STRUCTURE")
    courses.set(lesson || id, {lesson_no: lesson, source_lesson_id: id, name})
  }
  return [...courses.values()]
}

export async function openPersonalSchedule(page: Page): Promise<SemesterOption[]> {
  await page.goto(`${BJW_ORIGIN}/eams/home.action`, {waitUntil: "load", timeout: 30000})
  await new BkjwClient(page.context()).ensureAuthenticated(page)
  // EAMS loads the query fragment inside its home page. Opening the fragment
  // alone loses the shared bg/jQuery scripts needed by searchTable().
  const menu = page.locator('a[href*="/eams/courseTableForStd.action"]').first()
  if (!await menu.count()) throw new Error("PERSONAL_MENU_MISSING")
  // The entry may live in a collapsed navigation group. EAMS binds its AJAX
  // loader to the anchor; the same DOM click is used by the MyCHU importer.
  await menu.evaluate(element => (element as HTMLElement).click())
  try {
    await page.waitForFunction(() => {
      const win = window as unknown as {searchTable?: unknown}
      return !!document.querySelector('#courseTableType') && typeof win.searchTable === 'function'
    }, undefined, {timeout: 20000})
  } catch { throw new Error("PERSONAL_QUERY_FAILED") }
  await page.evaluate(() => {
    // Semester-calendar initialization can invoke a native form submission.
    // Keep the EAMS fragment in the home document, as in the MyCHU importer.
    HTMLFormElement.prototype.submit = function () {}
    const win = window as unknown as {bg?: {form?: {submit?: (form: unknown, action: unknown, target?: unknown) => unknown}}}
    const form = win.bg?.form
    const submit = form?.submit
    if (form && submit) form.submit = (f, action, target) => target ? submit.call(form, f, action, target) : undefined
  })
  try { return await discoverPageSemesters(page) }
  catch { throw new Error("PERSONAL_CALENDAR_FAILED") }
}

export async function readPersonalCourses(page: Page, semester: SemesterOption, signal: AbortSignal): Promise<PersonalCourse[]> {
  if (signal.aborted) throw new Error("CANCELLED")
  // Fill the hidden target through the query page's existing form and submit
  // exactly its own read-only timetable request.
  const [result] = await Promise.all([page.waitForResponse(r => r.url().includes('/eams/courseTableForStd!courseTable.action') && new URLSearchParams(r.request().postData() ?? new URL(r.url()).search).get("semester.id") === semester.id, {timeout: 30000}),
  page.evaluate(id => {
    const target = document.querySelector<HTMLInputElement>('#semesterCalendar_target, input[name="semester.id"]')
    if (!target) throw new Error("PAGE_STRUCTURE")
    const kind = document.querySelector<HTMLSelectElement>("#courseTableType")
    if (!kind) throw new Error("PAGE_STRUCTURE")
    kind.value = "std"
    if (kind.value !== "std") throw new Error("PERMISSION_DENIED")
    target.value = id
    const win = window as unknown as {table0?: unknown; searchTable?: () => void}
    win.table0 = undefined
    if (typeof win.searchTable !== 'function') throw new Error("PAGE_STRUCTURE")
    win.searchTable()
  }, semester.id)])
  if (!result.ok()) throw new Error("AUTH_REQUIRED")
  await page.waitForFunction(() => Array.isArray((window as unknown as {table0?: {activities?: unknown}}).table0?.activities), undefined, {timeout: 15000})
  await new BkjwClient(page.context()).ensureAuthenticated(page)
  if (signal.aborted) throw new Error("CANCELLED")
  return personalCourses(await page.evaluate(() => (window as unknown as {table0: {activities: unknown}}).table0.activities))
}
