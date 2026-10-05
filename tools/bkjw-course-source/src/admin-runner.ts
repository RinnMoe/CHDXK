import { createInterface } from "node:readline"
import { mkdir, mkdtemp, readFile, rm } from "node:fs/promises"
import { fileURLToPath } from "node:url"
import { join, resolve } from "node:path"
import { chromium } from "playwright"
import { BkjwClient } from "./client.js"
import { fetchSemester } from "./orchestrator.js"
import { semesterDirectory } from "./storage.js"
import { qrLogin } from "./qr-login.js"
import { BkjwError } from "./types.js"

const emit = (event: object) => process.stdout.write(`${JSON.stringify(event)}\n`)
const abort = new AbortController()
const input = createInterface({input: process.stdin})
let choose: ((semesters: string[]) => void) | undefined
const selection = new Promise<string[]>((resolve) => { choose = resolve })
input.on("line", (line) => {
  try {
    const cmd = JSON.parse(line) as {type: string; semesters?: string[]}
    if (cmd.type === "cancel") { abort.abort(); choose?.([]) }
    if (cmd.type === "sync" && Array.isArray(cmd.semesters)) choose?.(cmd.semesters)
  } catch { abort.abort(); choose?.([]) }
})
input.on("close", () => { abort.abort(); choose?.([]) })
process.on("SIGTERM", () => { abort.abort(); choose?.([]) })

let browser: Awaited<ReturnType<typeof chromium.launch>> | undefined
let output: string | undefined
let stage = "BROWSER_START_FAILED"
try {
  browser = await chromium.launch({headless: true})
  const context = await browser.newContext({locale: "zh-CN"})
  stage = "QR_LOGIN_FAILED"
  await qrLogin(context, emit, abort.signal)
  if (abort.signal.aborted) throw new Error("CANCELLED")
  const client = new BkjwClient(context)
  const page = await client.newPage()
  stage = "SEMESTER_DISCOVERY_FAILED"
  const semesters = await client.discoverSemesters(page)
  await client.closePage(page)
  emit({type: "ready", semesters: semesters.map((item) => item.normalized)})
  const selected = await selection
  if (abort.signal.aborted) throw new Error("CANCELLED")
  const runtimeRoot = resolve(fileURLToPath(new URL("../../../.local/course-sync/", import.meta.url)))
  stage = "COURSE_COLLECTION_FAILED"
  await mkdir(runtimeRoot, {recursive: true})
  output = await mkdtemp(join(runtimeRoot, "run-"))
  for (const semester of selected) {
    if (abort.signal.aborted) throw new Error("CANCELLED")
    const option = semesters.find((item) => item.normalized === semester)
    if (!option) throw new Error("INVALID_SEMESTER")
    emit({type: "collecting", semester})
    const report = await fetchSemester(client, option, {outputRoot: output, pageSize: 100, detailConcurrency: 2, resume: false, signal: abort.signal,
      log: () => emit({type: "progress", semester})})
    const directory = semesterDirectory(output, semester)
    // The server's existing import validator is the final authority on completeness.
    emit({type: "archive", semester, document: JSON.parse(await readFile(join(directory, "semester.json"), "utf8")), report: JSON.parse(await readFile(join(directory, "report.json"), "utf8")), complete: report.completed})
  }
  emit({type: "done"})
} catch (error) {
  const code = error instanceof BkjwError ? error.code : error instanceof Error ? error.message : "UNKNOWN"
  const known = ["QR_PAGE_CHANGED", "QR_ENTRY_CHANGED", "QR_CALLBACK_INVALID", "QR_IMAGE_INVALID", "QR_ENTRY_INVALID", "QR_IMAGE_UNAVAILABLE", "QR_EXPIRED", "WECHAT_UNBOUND", "CANCELLED"]
  // Never expose redirected URLs, codes, cookies, or collector internals in events.
  const policyCodes = ["QR_LOGIN_FAILED", "AUTH_REQUIRED", "PAGE_STRUCTURE", "PERMISSION_DENIED"]
  emit({type: "error", error: [...known, ...policyCodes].includes(code) ? code : stage})
} finally {
  await browser?.close()
  if (output) await rm(output, {recursive: true, force: true})
  input.close()
}
