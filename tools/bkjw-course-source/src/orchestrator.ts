import { BkjwClient, BJW_HOST } from "./client.js"
import { detailWarnings } from "./parse.js"
import {
  initialCheckpoint,
  loadCheckpoint,
  loadDocument,
  loadReport,
  saveSnapshot,
} from "./storage.js"
import {
  AuthenticationRequiredError,
  BkjwError,
  InterruptedError,
  SCHEMA_VERSION,
  type Checkpoint,
  type FailedLesson,
  type LessonRecord,
  type Report,
  type SemesterDocument,
  type SemesterOption,
} from "./types.js"

const MAX_ATTEMPTS_PER_RUN = 3
const SAVE_EVERY_DETAILS = 10

export interface FetchSemesterOptions {
  outputRoot: string
  pageSize: number
  detailConcurrency: number
  resume: boolean
  signal?: AbortSignal
  log?: (message: string) => void
}

export async function fetchSemester(
  client: BkjwClient,
  option: SemesterOption,
  options: FetchSemesterOptions,
): Promise<Report> {
  const log = options.log ?? (() => undefined)
  const existingCheckpoint = options.resume ? await loadCheckpoint(options.outputRoot, option.normalized) : null
  const existingDocument = options.resume ? await loadDocument(options.outputRoot, option.normalized) : null
  const existingReport = options.resume ? await loadReport(options.outputRoot, option.normalized) : null

  const checkpoint = existingCheckpoint ?? initialCheckpoint(option.normalized, option.id, options.pageSize)
  const document = existingDocument ?? createDocument(option)
  const report = existingReport ?? createReport(option)
  const persistedSemesterIds = [
    checkpoint.semesterId,
    document.source.semesterId,
    report.semesterId,
  ]
  if (options.resume && persistedSemesterIds.some((semesterId) => semesterId !== option.id)) {
    throw new BkjwError(
      `断点中的学期 ID 与当前站点不一致（当前 ${option.id}），为避免混写请使用新的 --output 目录`,
      "SEMESTER_ID_MISMATCH",
    )
  }
  const failed = new Map<string, FailedLesson>(checkpoint.failed.map((item) => [item.lessonId, item]))
  if (checkpoint.listComplete && !report.listComplete) report.listComplete = true
  if (document.expectedCount === null && checkpoint.expectedCount !== null) {
    document.expectedCount = checkpoint.expectedCount
  }
  if (report.expectedCount === null && checkpoint.expectedCount !== null) {
    report.expectedCount = checkpoint.expectedCount
  }
  const lessonMap = new Map<string, LessonRecord>()
  const duplicateIds = new Set(report.duplicateLessonIds ?? [])
  for (const lesson of document.lessons) {
    if (lessonMap.has(lesson.id)) duplicateIds.add(lesson.id)
    else lessonMap.set(lesson.id, lesson)
  }
  for (const lessonId of failed.keys()) {
    if (!lessonMap.has(lessonId)) failed.delete(lessonId)
  }
  report.duplicateLessonIds = Array.from(duplicateIds).sort()
  checkpoint.completedLessonIds = unique(checkpoint.completedLessonIds).filter((id) => {
    const lesson = lessonMap.get(id)
    return lesson !== undefined && lesson.detail !== null
  })
  for (const lesson of lessonMap.values()) {
    if (lesson.detail !== null) {
      if (!checkpoint.completedLessonIds.includes(lesson.id)) checkpoint.completedLessonIds.push(lesson.id)
      failed.delete(lesson.id)
    }
  }
  if (checkpoint.listComplete && checkpoint.expectedCount !== null && lessonMap.size !== checkpoint.expectedCount) {
    checkpoint.listComplete = false
    report.listComplete = false
  }
  const persistQueue = new PersistQueue()

  const persist = async (): Promise<void> => {
    checkpoint.updatedAt = new Date().toISOString()
    report.updatedAt = checkpoint.updatedAt
    document.lessons = Array.from(lessonMap.values())
    report.lessonCount = document.lessons.length
    report.detailSuccessCount = document.lessons.filter((lesson) => lesson.detail !== null).length
    report.missingDetailCount = document.lessons.filter((lesson) => lesson.detail === null).length
    report.failedLessonIds = Array.from(failed.keys()).sort()
    checkpoint.failed = Array.from(failed.values()).sort((a, b) => a.lessonId.localeCompare(b.lessonId))
    await persistQueue.run(() => saveSnapshot(options.outputRoot, { document, checkpoint, report }))
  }

  try {
    if (isComplete(checkpoint, report, document)) {
      report.completed = true
      await persist()
      log(`学期 ${option.normalized} 已完成，跳过重复采集`)
      return report
    }

    throwIfAborted(options.signal)
    if (!checkpoint.listComplete) {
      log(`读取 ${option.normalized} 的全校开课列表`)
      const page = await client.newPage()
      try {
        const listing = await client.collectListings(page, option.id, options.pageSize)
        const nextLessons = new Map<string, LessonRecord>()
        const nextDuplicateIds = new Set<string>()
        for (const lesson of listing.lessons) {
          if (nextLessons.has(lesson.id)) nextDuplicateIds.add(lesson.id)
          else nextLessons.set(lesson.id, lesson)
        }
        checkpoint.expectedCount = listing.expectedCount
        checkpoint.pageSize = listing.pageSize
        checkpoint.lastPage = listing.pageCount
        document.expectedCount = listing.expectedCount
        report.expectedCount = listing.expectedCount
        report.listComplete = true
        report.sourcePage.gridSignature = listing.gridSignature
        lessonMap.clear()
        for (const [lessonId, lesson] of nextLessons) lessonMap.set(lessonId, lesson)
        duplicateIds.clear()
        for (const lessonId of nextDuplicateIds) duplicateIds.add(lessonId)
        report.duplicateLessonIds = Array.from(nextDuplicateIds).sort()
        for (const lessonId of failed.keys()) {
          if (!lessonMap.has(lessonId)) failed.delete(lessonId)
        }
        checkpoint.completedLessonIds = checkpoint.completedLessonIds.filter((lessonId) => {
          const lesson = lessonMap.get(lessonId)
          return lesson !== undefined && lesson.detail !== null
        })
        checkpoint.listComplete = true
        await persist()
        log(`列表读取完成：${listing.lessons.length} 条教学任务，${listing.pageCount} 页`)
      } finally {
        await client.closePage(page)
      }
    }

    throwIfAborted(options.signal)
    const pending = Array.from(lessonMap.values()).filter((lesson) => lesson.detail === null)
    if (pending.length > 0) {
      log(`读取 ${pending.length} 条课程详情，并发数 ${options.detailConcurrency}`)
      await fetchDetails(client, pending, failed, checkpoint, persist, options)
    }

    checkpoint.failed = Array.from(failed.values())
    report.completed = isComplete(checkpoint, report, document)
    if (!report.completed) {
      report.warnings = unique([...report.warnings, "仍有课程详情未成功读取，未标记为完成"])
    }
    await persist()
    if (!report.completed) {
      throw new BkjwError(`学期 ${option.normalized} 采集未完成，请检查 report.json 后使用 --resume 重试`, "INCOMPLETE")
    }
    log(`学期 ${option.normalized} 采集完成`)
    return report
  } catch (error) {
    report.completed = false
    report.warnings = unique([...report.warnings, describeError(error)])
    await persist()
    throw error
  }
}

async function fetchDetails(
  client: BkjwClient,
  pending: LessonRecord[],
  failed: Map<string, FailedLesson>,
  checkpoint: Checkpoint,
  persist: () => Promise<void>,
  options: FetchSemesterOptions,
): Promise<void> {
  let nextIndex = 0
  let completedSinceSave = 0
  let fatalError: Error | null = null
  const worker = async (): Promise<void> => {
    while (true) {
      if (fatalError) return
      throwIfAborted(options.signal)
      const index = nextIndex
      nextIndex += 1
      const lesson = pending[index]
      if (!lesson) return
      let success = false
      let lastError = "未知详情错误"
      for (let attempt = 0; attempt < MAX_ATTEMPTS_PER_RUN; attempt += 1) {
        throwIfAborted(options.signal)
        const totalAttempts = (checkpoint.retries[lesson.id] ?? 0) + 1
        checkpoint.retries[lesson.id] = totalAttempts
        try {
          const detail = await client.fetchDetail(lesson.links.detail)
          lesson.detail = detail
          lesson.warnings = unique([
            ...lesson.warnings,
            ...detailWarnings(detail),
            ...(detail.basic.courseCode && detail.basic.courseCode !== lesson.summary.courseCode
              ? ["详情页课程代码与列表页不一致"]
              : []),
          ])
          if (!checkpoint.completedLessonIds.includes(lesson.id)) checkpoint.completedLessonIds.push(lesson.id)
          failed.delete(lesson.id)
          success = true
          completedSinceSave += 1
          break
        } catch (error) {
          if (error instanceof AuthenticationRequiredError) {
            fatalError = error
            return
          }
          lastError = describeError(error)
          if (attempt + 1 < MAX_ATTEMPTS_PER_RUN) await delay(400 * 2 ** attempt, options.signal)
        }
      }
      if (!success) {
        lesson.warnings = unique([...lesson.warnings, `详情页读取失败：${lastError}`])
        failed.set(lesson.id, {
          lessonId: lesson.id,
          error: lastError,
          attempts: checkpoint.retries[lesson.id] ?? MAX_ATTEMPTS_PER_RUN,
          updatedAt: new Date().toISOString(),
        })
      }
      if (completedSinceSave >= SAVE_EVERY_DETAILS || index === pending.length - 1) {
        completedSinceSave = 0
        await persist()
      }
      await delay(200, options.signal)
    }
  }

  const workerCount = Math.min(Math.max(1, options.detailConcurrency), 8, pending.length)
  const results = await Promise.allSettled(Array.from({ length: workerCount }, () => worker()))
  const rejected = results.find((result): result is PromiseRejectedResult => result.status === "rejected")
  if (fatalError) throw fatalError
  if (rejected) throw rejected.reason
}

function createDocument(option: SemesterOption): SemesterDocument {
  return {
    schemaVersion: SCHEMA_VERSION,
    source: {
      host: BJW_HOST,
      path: "/eams/stdSyllabus!search.action",
      projectId: "1",
      projectLabel: "本科",
      semesterId: option.id,
      label: option.label,
    },
    semester: option.normalized,
    fetchedAt: new Date().toISOString(),
    expectedCount: null,
    lessons: [],
  }
}

function createReport(option: SemesterOption): Report {
  const now = new Date().toISOString()
  return {
    schemaVersion: SCHEMA_VERSION,
    semester: option.normalized,
    semesterId: option.id,
    startedAt: now,
    updatedAt: now,
    completed: false,
    listComplete: false,
    expectedCount: null,
    lessonCount: 0,
    detailSuccessCount: 0,
    missingDetailCount: 0,
    duplicateLessonIds: [],
    failedLessonIds: [],
    warnings: [],
    sourcePage: {
      host: BJW_HOST,
      path: "/eams/stdSyllabus!search.action",
      gridSignature: null,
    },
  }
}

function isComplete(checkpoint: Checkpoint, report: Report, document: SemesterDocument): boolean {
  return (
    checkpoint.listComplete &&
    report.listComplete &&
    checkpoint.expectedCount !== null &&
    document.expectedCount === checkpoint.expectedCount &&
    document.lessons.length === checkpoint.expectedCount &&
    document.lessons.every((lesson) => lesson.detail !== null) &&
    checkpoint.failed.length === 0 &&
    (report.duplicateLessonIds?.length ?? 0) === 0
  )
}

function throwIfAborted(signal: AbortSignal | undefined): void {
  if (signal?.aborted) throw new InterruptedError()
}

function delay(milliseconds: number, signal: AbortSignal | undefined): Promise<void> {
  return new Promise((resolve, reject) => {
    if (signal?.aborted) {
      reject(new InterruptedError())
      return
    }
    let timer: NodeJS.Timeout | undefined
    const onAbort = () => {
      if (timer) clearTimeout(timer)
      signal?.removeEventListener("abort", onAbort)
      reject(new InterruptedError())
    }
    const onDone = () => {
      signal?.removeEventListener("abort", onAbort)
      resolve()
    }
    timer = setTimeout(onDone, milliseconds)
    signal?.addEventListener("abort", onAbort, { once: true })
  })
}

function unique(values: string[]): string[] {
  return values.filter((value, index) => values.indexOf(value) === index)
}

function describeError(error: unknown): string {
  if (error instanceof Error) return error.message
  return String(error)
}

class PersistQueue {
  private tail: Promise<void> = Promise.resolve()

  async run(task: () => Promise<void>): Promise<void> {
    const next = this.tail.then(task)
    this.tail = next.catch(() => undefined)
    await next
  }
}
