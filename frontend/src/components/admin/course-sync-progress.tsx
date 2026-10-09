import { useEffect, useState } from "react"

import {
  progressPercent,
  type SyncProgress,
} from "./course-sync-progress-utils"

const labels: Record<SyncProgress["stage"], string> = {
  listing: "读取课程列表",
  details: "获取课程详情",
  validating: "检查数据完整性",
  importing: "校验并保存学期数据",
  completed: "本学期已保存",
}

function duration(seconds: number) {
  const value = Math.max(0, Math.floor(seconds))
  return value < 60
    ? `${value} 秒`
    : `${Math.floor(value / 60)} 分 ${value % 60} 秒`
}

export function CourseSyncProgress({
  progress,
  startedAt,
  active,
}: {
  progress?: SyncProgress
  startedAt?: string
  active: boolean
}) {
  const [now, setNow] = useState(() => Date.now())
  useEffect(() => {
    if (!active) return
    const timer = window.setInterval(() => setNow(Date.now()), 1000)
    return () => window.clearInterval(timer)
  }, [active])
  const percent = progress ? progressPercent(progress) : undefined
  const counting = progress && ["listing", "details"].includes(progress.stage)
  const elapsedClock =
    !active && progress ? Date.parse(progress.updated_at) : now
  return (
    <div className="space-y-3" data-sync-progress>
      <div className="flex flex-wrap items-center justify-between gap-2 text-sm">
        <span className="font-medium">
          {progress ? labels[progress.stage] : "等待采集进度"}
        </span>
        {percent !== undefined && (
          <span className="tabular-nums">{percent}%</span>
        )}
      </div>
      {counting && (
        <>
          <progress
            aria-label={`${labels[progress.stage]}进度`}
            max={100}
            value={percent}
            className="h-2 w-full accent-primary"
          />
          <div className="flex flex-wrap gap-x-6 gap-y-1 text-sm tabular-nums">
            <span>
              已读取 {progress.completed.toLocaleString()} /{" "}
              {progress.total?.toLocaleString() ?? "待确认"} 条教学班
            </span>
            {!!progress.pages && (
              <span>
                列表页数 {progress.page ?? 0} / {progress.pages}
              </span>
            )}
            {progress.stage === "details" && (
              <>
                <span>读取失败 {progress.failed} 条</span>
                <span>已重试 {progress.retries} 次</span>
                {!!progress.cached && (
                  <span>从缓存恢复 {progress.cached} 条</span>
                )}
              </>
            )}
          </div>
        </>
      )}
      <div className="flex flex-wrap gap-x-6 gap-y-1 text-xs text-muted-foreground tabular-nums">
        {startedAt && (
          <span>
            已用时 {duration((elapsedClock - Date.parse(startedAt)) / 1000)}
          </span>
        )}
        {progress && (
          <span>
            进度更新于{" "}
            {duration((now - Date.parse(progress.updated_at)) / 1000)}前
          </span>
        )}
      </div>
      {active && progress?.retry_at && (
        <p className="text-sm text-muted-foreground" role="status">
          {Date.parse(progress.retry_at) > now
            ? `教务响应暂时异常，正在等待自动重试（约 ${duration((Date.parse(progress.retry_at) - now) / 1000)}）。`
            : "正在重新请求教务系统。"}
          已读取的数据保留，可继续等待。
        </p>
      )}
      {active && progress && now - Date.parse(progress.updated_at) > 30000 && (
        <p className="text-xs text-muted-foreground">
          正在等待教务系统响应，进度暂未更新。可继续等待或取消任务。
        </p>
      )}
    </div>
  )
}
