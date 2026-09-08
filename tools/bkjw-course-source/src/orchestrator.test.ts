import { mkdtemp, readFile, rm } from "node:fs/promises"
import { tmpdir } from "node:os"
import { join } from "node:path"

import { afterEach, describe, expect, it } from "vitest"

import { fetchSemester } from "./orchestrator.js"
import { readJsonIfExists } from "./storage.js"
import { BkjwError, type LessonDetail, type LessonRecord, type SemesterOption } from "./types.js"
import type { BkjwClient, ListingCollection } from "./client.js"

const roots: string[] = []
const semester: SemesterOption = {
  id: "222",
  academicYear: "2025-2026",
  term: "1",
  label: "2025-2026学年1学期",
  normalized: "2025-2026-1",
}

const detail: LessonDetail = {
  basic: {
    semester: semester.label,
    courseCode: "CS101",
    courseName: "程序设计",
    credit: 3,
    courseType: "通识课程",
    department: "信息学院",
    campus: "本部",
    language: "中文",
    assessment: "考试",
    teachers: ["张三"],
  },
  arrangement: {
    totalHours: 48,
    weeklyHours: 4,
    weekRange: "1-12",
    startWeek: 1,
    endWeek: 12,
    weekCount: 12,
  },
  restrictions: [],
  remark: "",
  raw: {},
}

function lesson(id: string): LessonRecord {
  return {
    id,
    summary: {
      lessonNo: `${id}.01`,
      courseCode: "CS101",
      courseName: "程序设计",
      courseType: "通识课程",
      teachClass: "班级:2025A",
      teachers: ["张三"],
      actualCount: 1,
      limitCount: 40,
      credit: 3,
      coursePeriod: "48/4",
      weekRange: "1-12",
      raw: {},
    },
    detail: null,
    links: { detail: `/eams/stdSyllabus!info.action?lesson.id=${id}`, syllabus: "" },
    warnings: [],
  }
}

afterEach(async () => {
  await Promise.all(roots.splice(0).map((root) => rm(root, { recursive: true, force: true })))
})

describe("semester checkpoint orchestration", () => {
  it("writes a complete snapshot and resumes completed work without refetching", async () => {
    const root = await mkdtemp(join(tmpdir(), "bkjw-source-"))
    roots.push(root)
    const listing: ListingCollection = {
      lessons: [lesson("101"), lesson("102")],
      expectedCount: 2,
      gridSignature: "headers",
      pageCount: 1,
      pageSize: 1000,
    }
    const fake = {
      newPage: async () => ({}),
      closePage: async () => undefined,
      collectListings: async () => listing,
      fetchDetail: async () => detail,
    } as unknown as BkjwClient
    const first = await fetchSemester(fake, semester, {
      outputRoot: root,
      pageSize: 1000,
      detailConcurrency: 2,
      resume: false,
    })
    expect(first.completed).toBe(true)
    expect(first.detailSuccessCount).toBe(2)
    const saved = JSON.parse(await readFile(join(root, semester.normalized, "semester.json"), "utf8"))
    expect(saved).toMatchObject({ schema_version: 1, expected_count: 2 })
    expect(saved.lessons).toHaveLength(2)
    expect(saved.lessons[0].summary.course_code).toBe("CS101")

    const noRefetch = {
      newPage: async () => {
        throw new Error("should not open a listing page")
      },
    } as unknown as BkjwClient
    const resumed = await fetchSemester(noRefetch, semester, {
      outputRoot: root,
      pageSize: 1000,
      detailConcurrency: 2,
      resume: true,
    })
    expect(resumed.completed).toBe(true)
    expect(await readJsonIfExists(join(root, semester.normalized, "report.json"))).toMatchObject({ completed: true })
  })

  it("records detail failures and retries them on resume", async () => {
    const root = await mkdtemp(join(tmpdir(), "bkjw-source-failure-"))
    roots.push(root)
    const listing: ListingCollection = {
      lessons: [lesson("101")],
      expectedCount: 1,
      gridSignature: "headers",
      pageCount: 1,
      pageSize: 1000,
    }
    const failing = {
      newPage: async () => ({}),
      closePage: async () => undefined,
      collectListings: async () => listing,
      fetchDetail: async () => {
        throw new BkjwError("详情暂时不可用", "HTTP_ERROR")
      },
    } as unknown as BkjwClient
    await expect(
      fetchSemester(failing, semester, {
        outputRoot: root,
        pageSize: 1000,
        detailConcurrency: 1,
        resume: false,
      }),
    ).rejects.toThrow("采集未完成")
    expect(await readJsonIfExists(join(root, semester.normalized, "checkpoint.json"))).toMatchObject({
      list_complete: true,
      failed: [{ lesson_id: "101", attempts: 3 }],
    })

    const recovered = {
      fetchDetail: async () => detail,
    } as unknown as BkjwClient
    const report = await fetchSemester(recovered, semester, {
      outputRoot: root,
      pageSize: 1000,
      detailConcurrency: 1,
      resume: true,
    })
    expect(report.completed).toBe(true)
    expect(report.failedLessonIds).toEqual([])
  })
})
