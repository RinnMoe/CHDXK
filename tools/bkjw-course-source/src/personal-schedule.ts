import type { Page } from "playwright"
import { BkjwClient, BJW_ORIGIN, discoverPageSemesters } from "./client.js"
import type { SemesterOption } from "./types.js"
import { setTimeout as delay } from "node:timers/promises"

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
  // EAMS stores arrays per timetable slot, including empty slots.
  const activities = value.every(Array.isArray) ? value.flat() : value
  if (activities.length > 10000) throw new Error("PAGE_STRUCTURE")
  for (const item of activities) {
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
  try { await menu.waitFor({state: 'attached', timeout: 20000}) }
  catch { throw new Error("PERSONAL_MENU_MISSING") }
  await page.evaluate(() => {
    // Semester-calendar initialization can invoke a native form submission.
    // Keep the EAMS fragment in the home document, as in the MyCHU importer.
    HTMLFormElement.prototype.submit = function () {}
    const win = window as unknown as {bg?: {form?: {submit?: (form: unknown, action: unknown, target?: unknown) => unknown}}}
    const form = win.bg?.form
    const submit = form?.submit
    if (form && submit) form.submit = (f, action, target) => target ? submit.call(form, f, action, target) : undefined
  })
  // The entry may live in a collapsed navigation group. EAMS binds its AJAX
  // loader to the anchor; the same DOM click is used by the MyCHU importer.
  await menu.evaluate(element => (element as HTMLElement).click())
  try {
    await page.waitForFunction(() => {
      const win = window as unknown as {searchTable?: unknown}
      return !!document.querySelector('#courseTableType') && typeof win.searchTable === 'function'
    }, undefined, {timeout: 20000})
  } catch { throw new Error("PERSONAL_QUERY_FAILED") }
  // The personal page loads the complete semester catalog asynchronously.
  // Read that catalog instead of driving a popup which may close itself.
  await page.waitForFunction(() => {
    const jq = (window as unknown as {jQuery?: (e: Element) => {data:(key: string) => unknown}}).jQuery
    if (!jq) return true
    return Array.from(document.querySelectorAll("input.calendar-text, #semesterCalendar_target")).some(input => {
      const data = jq(input).data("semesters")
      return data && typeof data === "object" && Object.values(data).some(value => Array.isArray(value) && value.length)
    })
  }, undefined, {timeout: 15000}).catch(() => { throw new Error("PERSONAL_CALENDAR_FAILED") })
  const stored = await page.evaluate(() => {
    const jq = (window as unknown as {jQuery?: (e: Element) => {data:(key: string) => unknown}}).jQuery
    for (const input of document.querySelectorAll("input.calendar-text, #semesterCalendar_target")) {
      const data = jq?.(input).data("semesters")
      if (data && typeof data === "object") {
        const items = Object.values(data).flat() as {id?: unknown; schoolYear?: unknown; name?: unknown}[]
        return items.filter(item => item && item.id && /^\d{4}-\d{4}$/.test(String(item.schoolYear)) && /^\d+$/.test(String(item.name))).map(item => ({
          id:String(item.id), academicYear:String(item.schoolYear), term:String(item.name),
          label:`${item.schoolYear}学年${item.name}学期`, normalized:`${item.schoolYear}-${item.name}`,
        }))
      }
    }
    return []
  })
  if (stored.length) return stored
  try { return await discoverPageSemesters(page) }
  catch { throw new Error("PERSONAL_CALENDAR_FAILED") }
}

async function readPersonalCoursesOnce(page: Page, semester: SemesterOption, signal: AbortSignal): Promise<PersonalCourse[]> {
  if (signal.aborted) throw new Error("CANCELLED")
  // Fill the hidden target through the query page's existing form and submit
  // exactly its own read-only timetable request.
  let result
  try {
  ;[result] = await Promise.all([page.waitForResponse(r => r.url().includes('/eams/courseTableForStd!courseTable.action') && new URLSearchParams(r.request().postData() ?? new URL(r.url()).search).get("semester.id") === semester.id, {timeout: 30000}),
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
  } catch (error) {
    if (signal.aborted) throw new Error("CANCELLED")
    if (error instanceof Error && error.message.includes("PERMISSION_DENIED")) throw new Error("PERMISSION_DENIED")
    throw new Error("PERSONAL_REQUEST_FAILED")
  }
  if (result.status() === 401) throw new Error("AUTH_REQUIRED")
  if (result.status() === 403) throw new Error("PERMISSION_DENIED")
  if (!result.ok() || await result.finished()) throw new Error("PERSONAL_REQUEST_FAILED")
  try {
    await page.waitForFunction(() => Array.isArray((window as unknown as {table0?: {activities?: unknown}}).table0?.activities), undefined, {timeout: 15000})
  } catch {
    if (signal.aborted) throw new Error("CANCELLED")
    await new BkjwClient(page.context()).ensureAuthenticated(page)
    throw new Error("PERSONAL_RENDER_FAILED")
  }
  await new BkjwClient(page.context()).ensureAuthenticated(page)
  if (signal.aborted) throw new Error("CANCELLED")
  const activities = await page.evaluate(() => (window as unknown as {table0: {activities: unknown}}).table0.activities)
  try { return personalCourses(activities) }
  catch { throw new Error("PERSONAL_PARSE_FAILED") }
}

// Retry only temporary request/render failures. Reloading the same authenticated
// page cancels stale requests and restores query controls before another attempt.
export async function readPersonalCourses(page: Page, semester: SemesterOption, signal: AbortSignal, onRetry?: (attempt: number) => void): Promise<PersonalCourse[]> {
  for (let attempt = 0; attempt < 3; attempt++) {
    try {
      if (signal.aborted) throw new Error("CANCELLED")
      if (attempt) {
        onRetry?.(attempt)
        await delay(attempt * 1000, undefined, {signal}).catch(() => {throw new Error("CANCELLED")})
        await openPersonalSchedule(page)
      }
      return await readPersonalCoursesOnce(page, semester, signal)
    } catch (error) {
      if (signal.aborted) throw new Error("CANCELLED")
      const code = error instanceof Error ? error.message : "UNKNOWN"
      if (attempt === 2 || !["PERSONAL_REQUEST_FAILED", "PERSONAL_RENDER_FAILED", "PERSONAL_MENU_MISSING", "PERSONAL_QUERY_FAILED", "PERSONAL_CALENDAR_FAILED"].includes(code)) throw error
    }
  }
  throw new Error("PERSONAL_REQUEST_FAILED")
}
