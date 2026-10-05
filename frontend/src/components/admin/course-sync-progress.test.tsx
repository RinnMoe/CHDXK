import { renderToStaticMarkup } from "react-dom/server"
import { describe, expect, it } from "vitest"
import { CourseSyncProgress } from "./course-sync-progress"
import {
  progressPercent,
  type SyncProgress,
} from "./course-sync-progress-utils"

const progress: SyncProgress = {
  stage: "details",
  completed: 120,
  total: 1000,
  failed: 2,
  retries: 5,
  updated_at: new Date().toISOString(),
}

describe("course sync progress", () => {
  it("counts only successful details in the percentage and reports failures separately", () => {
    expect(progressPercent(progress)).toBe(12)
    const html = renderToStaticMarkup(
      <CourseSyncProgress progress={progress} active={false} />
    )
    expect(html).toContain('value="12"')
    expect(html).toContain("读取失败 2 条")
    expect(html).toContain("已重试 5 次")
  })
  it("does not invent a percentage before the total is known or while saving", () => {
    expect(progressPercent({ ...progress, total: undefined })).toBeUndefined()
    expect(progressPercent({ ...progress, stage: "importing" })).toBeUndefined()
    const html = renderToStaticMarkup(
      <CourseSyncProgress
        progress={{ ...progress, stage: "listing", total: undefined }}
        active={false}
      />
    )
    expect(html).toContain("待确认")
    expect(html).not.toContain("value=")
  })

  it("explains a stalled response without declaring the job failed", () => {
    const html = renderToStaticMarkup(
      <CourseSyncProgress
        progress={{
          ...progress,
          updated_at: new Date(Date.now() - 60000).toISOString(),
        }}
        active
      />
    )
    expect(html).toContain("正在等待教务系统响应")
    expect(html).not.toContain("同步已停止")
  })

  it("freezes elapsed time at completion when reopening a finished job", () => {
    const time = Date.now()
    const html = renderToStaticMarkup(
      <CourseSyncProgress
        startedAt={new Date(time - 600000).toISOString()}
        progress={{
          ...progress,
          stage: "completed",
          updated_at: new Date(time - 60000).toISOString(),
        }}
        active={false}
      />
    )
    expect(html).toContain("9 分 0 秒")
  })
})
