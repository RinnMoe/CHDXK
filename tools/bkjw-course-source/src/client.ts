import { chromium, type BrowserContext, type Page } from "playwright"

import { normalizeTerm, parseDetailPage, parseListingPage, parseSemesterOptions } from "./parse.js"
import {
  AuthenticationRequiredError,
  BkjwError,
  PageStructureError,
  type LessonDetail,
  type LessonRecord,
  type SemesterOption,
} from "./types.js"

export const BJW_ORIGIN = "http://bkjw.chd.edu.cn"
export const BJW_HOST = "bkjw.chd.edu.cn"
const SSO_HOST = "ids.chd.edu.cn"
export const DEFAULT_PROJECT_ID = "1"
export const FALLBACK_PAGE_SIZES = [10, 20, 30, 50, 70, 100, 200, 500, 1000]
export const ALLOWED_PAGE_SIZES = FALLBACK_PAGE_SIZES
const PERSISTED_SESSION_MAX_AGE_SECONDS = 7 * 24 * 60 * 60

export interface LaunchOptions {
  profileDir: string
  headless: boolean
}

export interface ListingCollection {
  lessons: LessonRecord[]
  expectedCount: number
  gridSignature: string
  pageCount: number
  pageSize: number
}

export async function launchBkjwContext(options: LaunchOptions): Promise<BrowserContext> {
  const context = await chromium.launchPersistentContext(options.profileDir, {
    headless: options.headless,
    locale: "zh-CN",
    viewport: null,
  })
  await context.route("**/*", async (route) => {
    const request = route.request()
    const frame = request.frame()
    if (!request.isNavigationRequest() || frame !== frame.page().mainFrame()) {
      await route.continue()
      return
    }
    const url = request.url()
    if (url === "about:blank" || isAllowedUrl(url) || isAllowedSsoLoginUrl(url)) {
      await route.continue()
      return
    }
    await route.abort("blockedbyclient")
  })
  return context
}

export class BkjwClient {
  constructor(private readonly context: BrowserContext) {}

  async newPage(): Promise<Page> {
    return this.context.newPage()
  }

  async closePage(page: Page): Promise<void> {
    if (!page.isClosed()) await page.close()
  }

  /**
   * Chromium drops session cookies when a persistent context closes. Promote
   * only cookies belonging to the two allowed CHD hosts in memory so a
   * successful login can be reused by the next CLI process without exporting
   * or logging cookie values.
   */
  async persistSessionCookies(): Promise<void> {
    const expires = Math.floor(Date.now() / 1000) + PERSISTED_SESSION_MAX_AGE_SECONDS
    const sessionCookies = (await this.context.cookies())
      .filter((cookie) => cookie.expires < 0 && isAllowedCookieDomain(cookie.domain))
      .map((cookie) => ({ ...cookie, expires }))
    if (sessionCookies.length > 0) await this.context.addCookies(sessionCookies)
  }

  async ensureAuthenticated(page: Page): Promise<void> {
    const url = page.url()
    if (!isAllowedUrl(url)) {
      throw new AuthenticationRequiredError(`教务页面被重定向到不允许的地址：${displayUrl(url)}`)
    }
    const content = await page.content()
    if (looksLikeLoginPage(url, content)) throw new AuthenticationRequiredError()
  }

  async navigateLogin(page: Page): Promise<void> {
    const response = await page.goto(toAllowedUrl("/eams/login.action"), {
      waitUntil: "domcontentloaded",
      timeout: 30_000,
    })
    await page.waitForTimeout(100)
    const finalUrl = page.url()
    if (isAllowedUrl(finalUrl)) return
    if (isAllowedSsoLoginUrl(finalUrl)) return
    throw new BkjwError(`登录页面被重定向到不允许的地址：${displayUrl(finalUrl || response?.url() || "")}`, "URL_NOT_ALLOWED")
  }

  async openSearch(page: Page, semesterId: string): Promise<void> {
    await this.goto(page, searchUrl(semesterId))
    try {
      await page.locator("table.gridtable").waitFor({ state: "attached", timeout: 15_000 })
    } catch {
      throw new PageStructureError("列表页缺少课程网格表格（table.gridtable）")
    }
  }

  async navigateEntry(page: Page, requireAuthentication = true): Promise<void> {
    const response = await page.goto(toAllowedUrl("/eams/stdSyllabus.action"), {
      waitUntil: "domcontentloaded",
      timeout: 30_000,
    })
    await page.waitForTimeout(100)
    if (isSsoLoginUrl(page.url())) throw new AuthenticationRequiredError()
    if (!isAllowedUrl(page.url())) {
      throw new BkjwError(`教务页面被重定向到不允许的地址：${displayUrl(page.url())}`, "URL_NOT_ALLOWED")
    }
    const content = await page.content()
    if (isAuthenticationErrorResponse(response?.status() ?? 0, content)) throw new AuthenticationRequiredError()
    if (looksLikePermissionError(content)) throw new BkjwError("教务系统拒绝访问该页面，可能没有权限", "PERMISSION_DENIED")
    if (requireAuthentication) await this.ensureAuthenticated(page)
  }

  async discoverSemesters(page: Page): Promise<SemesterOption[]> {
    await this.navigateEntry(page)
    const semesterInput = page.locator('input[id$="Semester"]').first()
    try {
      await semesterInput.waitFor({ state: "visible", timeout: 15_000 })
    } catch {
      throw new PageStructureError("入口页缺少学期选择器")
    }
    await semesterInput.click()
    await page.waitForTimeout(100)

    const years = unique((await page.locator("#semesterCalendar_yearTb td").allTextContents()).map((year) => year.trim()).filter(Boolean))
    const discovered: SemesterOption[] = []
    for (const year of years) {
      const academicYear = year.match(/\d{4}-\d{4}/)?.[0] ?? year
      const yearCell = page.locator("#semesterCalendar_yearTb td").filter({ hasText: year }).first()
      if (!(await yearCell.isVisible())) {
        await semesterInput.click()
        await page.waitForTimeout(50)
      }
      await yearCell.click()
      await page.waitForTimeout(120)
      const termCells = await page.locator("#semesterCalendar_termTb td").all()
      for (const termCell of termCells) {
        const id = (await termCell.getAttribute("val"))?.trim() || (await termCell.getAttribute("data-value"))?.trim() || ""
        const termLabel = (await termCell.innerText()).trim()
        const term = normalizeTerm(termLabel)
        if (!id || !/^\d+$/.test(term)) continue
        discovered.push({
          id,
          academicYear,
          term,
          label: `${academicYear}学年${term}学期`,
          normalized: `${academicYear}-${term}`,
        })
      }
    }

    if (discovered.length > 0) return dedupeSemesters(discovered)
    return parseSemesterOptions(await page.content())
  }

  async collectListings(page: Page, semesterId: string, requestedPageSize: number): Promise<ListingCollection> {
    await this.openSearch(page, semesterId)
    let current = parseListingPage(await page.content())
    const desiredPageSize = choosePageSize(requestedPageSize, current.availablePageSizes)
    const selectedPageSize = await this.selectedPageSize(page)
    if (selectedPageSize !== desiredPageSize) {
      const oldGridState = await this.gridState(page)
      await page.locator('span[title="点击改变每页数据量"]').first().click()
      const select = page.locator("select.pgbar-selbox:visible").first()
      await select.selectOption(String(desiredPageSize))
      await page.locator("input.pgbar-go:visible").first().click()
      await this.waitForGridChange(page, oldGridState)
      current = parseListingPage(await page.content(), desiredPageSize)
    } else if (selectedPageSize > 0) {
      current = parseListingPage(await page.content(), selectedPageSize)
    }

    const lessons: LessonRecord[] = []
    const signatures = new Set<string>()
    let expectedCount = current.expectedCount
    let pageCount = 0
    let observed = current
    while (true) {
      if (observed.expectedCount !== expectedCount) {
        throw new PageStructureError(`分页总数发生变化：${expectedCount} → ${observed.expectedCount}`)
      }
      if (observed.pageNo !== pageCount + 1) {
        throw new PageStructureError(`分页页码不连续：收到第 ${observed.pageNo} 页，期望第 ${pageCount + 1} 页`)
      }
      pageCount += 1
      lessons.push(...observed.lessons)
      signatures.add(observed.gridSignature)
      if (observed.pageNo >= observed.maxPageNo) break

      const oldGridState = await this.gridState(page)
      const next = page.getByRole("link", { name: "后页›", exact: true }).first()
      if ((await next.count()) === 0) throw new PageStructureError(`第 ${observed.pageNo} 页缺少后页按钮`)
      await page.waitForTimeout(200)
      await next.click()
      await this.waitForGridChange(page, oldGridState)
      observed = parseListingPage(await page.content(), desiredPageSize)
    }

    if (lessons.length !== expectedCount) {
      throw new PageStructureError(`分页条目数不一致：页面声明 ${expectedCount} 条，实际读取 ${lessons.length} 条`)
    }
    if (signatures.size !== 1) {
      throw new PageStructureError("不同分页的表头结构不一致")
    }
    return {
      lessons,
      expectedCount,
      gridSignature: observed.gridSignature,
      pageCount,
      pageSize: desiredPageSize,
    }
  }

  async fetchDetail(path: string): Promise<LessonDetail> {
    const url = toAllowedUrl(path)
    const response = await this.context.request.get(url, {
      failOnStatusCode: false,
      maxRedirects: 0,
      timeout: 30_000,
    })
    const content = await response.text()
    const status = response.status()
    const location = response.headers().location
    if (status >= 300 && status < 400) {
      const redirect = location ? new URL(location, url).toString() : ""
      if (isSsoLoginUrl(redirect) || /\/eams\/login(?:\.action|$)/i.test(redirect)) {
        throw new AuthenticationRequiredError()
      }
      throw new BkjwError(`详情页重定向到非预期地址：${displayUrl(redirect || response.url())}`, "HTTP_REDIRECT")
    }
    if (status >= 400) throw new BkjwError(`详情页返回 HTTP ${status}`, "HTTP_ERROR")
    if (isAuthenticationErrorResponse(status, content) || isSsoLoginUrl(response.url())) {
      throw new AuthenticationRequiredError()
    }
    if (!isAllowedUrl(response.url())) {
      throw new BkjwError(`详情页返回了不允许的地址：${displayUrl(response.url())}`, "URL_NOT_ALLOWED")
    }
    if (looksLikePermissionError(content)) throw new BkjwError("教务系统拒绝访问该页面，可能没有权限", "PERMISSION_DENIED")
    if (looksLikeLoginPage(response.url(), content)) throw new AuthenticationRequiredError()
    return parseDetailPage(content)
  }

  private async goto(page: Page, pathOrUrl: string) {
    const url = toAllowedUrl(pathOrUrl)
    const response = await page.goto(url, { waitUntil: "domcontentloaded", timeout: 30_000 })
    await page.waitForTimeout(100)
    const content = await page.content()
    if (isAuthenticationErrorResponse(response?.status() ?? 0, content)) throw new AuthenticationRequiredError()
    if (isSsoLoginUrl(page.url())) throw new AuthenticationRequiredError()
    if (looksLikePermissionError(content)) throw new BkjwError("教务系统拒绝访问该页面，可能没有权限", "PERMISSION_DENIED")
    await this.ensureAuthenticated(page)
    return response
  }

  private async selectedPageSize(page: Page): Promise<number> {
    const select = page.locator('select[id$="_page_select"]').first()
    const value = await select.inputValue()
    return Number.parseInt(value, 10) || 0
  }

  private async gridState(page: Page): Promise<string> {
    const id = (await page.locator("table.gridtable").first().getAttribute("id")) ?? ""
    const range = await page.locator(".girdbar-pgbar, .gridbar-pgbar").first().textContent()
    return `${id}|${range?.trim() ?? ""}`
  }

  private async waitForGridChange(page: Page, oldGridState: string): Promise<void> {
    for (let attempt = 0; attempt < 80; attempt += 1) {
      const grid = page.locator("table.gridtable").first()
      if ((await grid.count()) > 0) {
        const currentState = await this.gridState(page)
        if (currentState !== oldGridState) {
          await page.waitForTimeout(100)
          return
        }
      }
      await page.waitForTimeout(150)
    }
    throw new PageStructureError("分页请求超时，课程表格未更新")
  }
}

export function searchUrl(semesterId: string, pageNo?: number, pageSize?: number): string {
  const params = new URLSearchParams({
    "lesson.project.id": DEFAULT_PROJECT_ID,
    "lesson.semester.id": semesterId,
  })
  if (pageNo !== undefined) params.set("pageNo", String(pageNo))
  if (pageSize !== undefined) params.set("pageSize", String(pageSize))
  return `/eams/stdSyllabus!search.action?${params.toString()}`
}

export function choosePageSize(requested: number, available: number[]): number {
  const candidates = available
    .filter((size) => Number.isFinite(size) && size > 0)
    .filter((size, index, values) => values.indexOf(size) === index)
    .sort((a, b) => a - b)
  const sizes = candidates.length > 0 ? candidates : FALLBACK_PAGE_SIZES
  const atMost = sizes.filter((size) => size <= requested)
  return atMost.at(-1) ?? sizes[0] ?? 10
}

function toAllowedUrl(pathOrUrl: string): string {
  const url = new URL(pathOrUrl, BJW_ORIGIN)
  if (!isAllowedUrl(url.toString())) {
    throw new BkjwError(`拒绝访问非教务系统地址：${displayUrl(url.toString())}`, "URL_NOT_ALLOWED")
  }
  return url.toString()
}

function displayUrl(value: string): string {
  if (!value) return "未知"
  try {
    const url = new URL(value)
    return `${url.origin}${url.pathname}`
  } catch {
    return "未知"
  }
}

function isAllowedUrl(value: string): boolean {
  try {
    const url = new URL(value)
    return (url.protocol === "http:" || url.protocol === "https:") && url.hostname === BJW_HOST && url.pathname.startsWith("/eams/")
  } catch {
    return false
  }
}

function isAllowedCookieDomain(value: string): boolean {
  const domain = value.replace(/^\./, "").toLowerCase()
  return domain === BJW_HOST || domain === SSO_HOST || domain === "chd.edu.cn"
}

function looksLikeLoginPage(url: string, content: string): boolean {
  try {
    const pathname = new URL(url).pathname
    if (/\/login(?:[!./]|$)/i.test(pathname)) return true
  } catch {
    return true
  }
  const lower = content.toLowerCase()
  return /type\s*=\s*["']password/i.test(lower) && !lower.includes("退出")
}

function isAuthenticationErrorResponse(status: number, content: string): boolean {
  return status === 401 || status === 403 || /AuthenticationException|security\.Authentication|未登录|登录超时/i.test(content)
}

function looksLikePermissionError(content: string): boolean {
  return /无权限|没有权限|权限不足|禁止访问|Access\s+Denied/i.test(content)
}

function isAllowedSsoLoginUrl(value: string): boolean {
  if (!isSsoLoginUrl(value)) return false
  try {
    const url = new URL(value)
    const service = url.searchParams.get("service")
    if (!service) return false
    const serviceUrl = new URL(service)
    return serviceUrl.hostname === BJW_HOST && serviceUrl.pathname === "/eams/login.action"
  } catch {
    return false
  }
}

function isSsoLoginUrl(value: string): boolean {
  try {
    const url = new URL(value)
    return url.protocol === "https:" && url.hostname === SSO_HOST && url.pathname.startsWith("/authserver/login")
  } catch {
    return false
  }
}

function unique(values: string[]): string[] {
  return values.filter((value, index) => values.indexOf(value) === index)
}

function dedupeSemesters(options: SemesterOption[]): SemesterOption[] {
  const seen = new Set<string>()
  return options.filter((option) => {
    const key = `${option.id}|${option.normalized}`
    if (seen.has(key)) return false
    seen.add(key)
    return true
  })
}
