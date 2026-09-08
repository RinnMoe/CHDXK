import { mkdir, readFile, rename, unlink, writeFile } from "node:fs/promises"
import { dirname, join, resolve } from "node:path"
import { randomUUID } from "node:crypto"
import { fileURLToPath } from "node:url"

import {
  SCHEMA_VERSION,
  type Checkpoint,
  type Report,
  type SemesterDocument,
} from "./types.js"

const PACKAGE_ROOT = resolve(fileURLToPath(new URL("..", import.meta.url)))
const REPOSITORY_ROOT = resolve(PACKAGE_ROOT, "..", "..")

export const DEFAULT_OUTPUT_ROOT = resolve(REPOSITORY_ROOT, "data", "bkjw")
export const DEFAULT_PROFILE_DIR = resolve(REPOSITORY_ROOT, "data", "bkjw", "browser-profile")

export interface SemesterSnapshot {
  document: SemesterDocument
  checkpoint: Checkpoint
  report: Report
}

export function semesterDirectory(outputRoot: string, semester: string): string {
  if (!/^\d{4}-\d{4}-\d+$/.test(semester)) {
    throw new Error(`非法规范化学期：${semester}`)
  }
  return join(resolve(outputRoot), semester)
}

export async function readJsonIfExists<T>(path: string): Promise<T | null> {
  try {
    const text = await readFile(path, "utf8")
    return JSON.parse(text) as T
  } catch (error) {
    if (isNodeError(error) && error.code === "ENOENT") return null
    throw new Error(`读取 JSON 失败 ${path}: ${error instanceof Error ? error.message : String(error)}`)
  }
}

export async function writeJsonAtomic(path: string, value: unknown): Promise<void> {
  await mkdir(dirname(path), { recursive: true })
  const temporary = `${path}.${process.pid}.${randomUUID()}.tmp`
  await writeFile(temporary, `${JSON.stringify(value, null, 2)}\n`, "utf8")
  try {
    await rename(temporary, path)
    return
  } catch (error) {
    if (!isNodeError(error) || !["EEXIST", "EPERM", "ENOTEMPTY"].includes(error.code ?? "")) {
      await safeUnlink(temporary)
      throw error
    }
  }

  // Windows cannot rename over an existing file. Keep the old file recoverable
  // until the new file is in place, then remove only the temporary backup.
  const backup = `${path}.${process.pid}.${randomUUID()}.previous`
  let movedOld = false
  try {
    try {
      await rename(path, backup)
      movedOld = true
    } catch (error) {
      if (!isNodeError(error) || error.code !== "ENOENT") throw error
    }
    await rename(temporary, path)
    if (movedOld) await safeUnlink(backup)
  } catch (error) {
    await safeUnlink(temporary)
    if (movedOld) {
      try {
        await rename(backup, path)
      } catch {
        // Preserve the backup for manual recovery if restoration also fails.
      }
    }
    throw error
  }
}

export async function saveSnapshot(root: string, snapshot: SemesterSnapshot): Promise<void> {
  const directory = semesterDirectory(root, snapshot.document.semester)
  await mkdir(directory, { recursive: true })
  // Checkpoint first makes an interruption conservative: it may repeat a detail
  // request, but it cannot claim progress that was not recorded in the data file.
  await writeJsonAtomic(join(directory, "checkpoint.json"), mapObjectKeys(snapshot.checkpoint, toSnakeCase))
  await writeJsonAtomic(join(directory, "semester.json"), mapObjectKeys(snapshot.document, toSnakeCase))
  await writeJsonAtomic(join(directory, "report.json"), mapObjectKeys(snapshot.report, toSnakeCase))
}

export async function loadCheckpoint(root: string, semester: string): Promise<Checkpoint | null> {
  const value = await readJsonIfExists<unknown>(join(semesterDirectory(root, semester), "checkpoint.json"))
  return value === null ? null : (mapObjectKeys(value, toCamelCase) as Checkpoint)
}

export async function loadDocument(root: string, semester: string): Promise<SemesterDocument | null> {
  const value = await readJsonIfExists<unknown>(join(semesterDirectory(root, semester), "semester.json"))
  return value === null ? null : (mapObjectKeys(value, toCamelCase) as SemesterDocument)
}

export async function loadReport(root: string, semester: string): Promise<Report | null> {
  const value = await readJsonIfExists<unknown>(join(semesterDirectory(root, semester), "report.json"))
  return value === null ? null : (mapObjectKeys(value, toCamelCase) as Report)
}

export function initialCheckpoint(semester: string, semesterId: string, pageSize: number): Checkpoint {
  return {
    schemaVersion: SCHEMA_VERSION,
    semester,
    semesterId,
    expectedCount: null,
    listComplete: false,
    lastPage: 0,
    pageSize,
    completedLessonIds: [],
    retries: {},
    failed: [],
    updatedAt: new Date().toISOString(),
  }
}

function isNodeError(error: unknown): error is NodeJS.ErrnoException {
  return error instanceof Error && "code" in error
}

async function safeUnlink(path: string): Promise<void> {
  try {
    await unlink(path)
  } catch (error) {
    if (!isNodeError(error) || error.code !== "ENOENT") throw error
  }
}

function mapObjectKeys(value: unknown, convert: (key: string) => string): unknown {
  if (Array.isArray(value)) return value.map((item) => mapObjectKeys(item, convert))
  if (value === null || typeof value !== "object") return value
  return Object.fromEntries(
    Object.entries(value).map(([key, item]) => [convert(key), mapObjectKeys(item, convert)]),
  )
}

function toSnakeCase(value: string): string {
  return value.replace(/[A-Z]/g, (letter) => `_${letter.toLowerCase()}`)
}

function toCamelCase(value: string): string {
  return value.replace(/_([a-z])/g, (_match, letter: string) => letter.toUpperCase())
}
