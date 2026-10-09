import { useImportSettings } from "@/hooks/use-import-settings"
import { copy } from "@/lib/copy"

import { useEffect, useRef, useState } from "react"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import { invalidateCourseDataQueries } from "@/hooks/use-admin-course"
import { apiClient } from "@/api/client"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { SchoolLoginFeedback } from "@/components/course/school-login-feedback"
import { getErrorMessage } from "./admin-utils"
import { CourseSyncProgress } from "./course-sync-progress"
import type { SyncProgress } from "./course-sync-progress-utils"

interface SyncStatus {
  id: string
  state:
    | "queued"
    | "connecting"
    | "waiting"
    | "scanned"
    | "authenticating"
    | "discovering"
    | "ready"
    | "running"
    | "done"
    | "failed"
    | "cancelled"
  qr?: string
  queue_position?: number
  semesters: string[]
  selected: string[]
  current?: string
  message?: string
  started_at?: string
  progress?: SyncProgress
  results: { semester: string; course_count: number; section_count: number }[]
}
const storageKey = "chdxk.course-sync-job"
const terminal = (state?: string) =>
  ["done", "failed", "cancelled"].includes(state ?? "")
const states: Record<SyncStatus["state"], string> = {
  queued: "正在排队，轮到后会显示登录二维码",
  authenticating: "手机已确认，正在连接本科教务",
  discovering: "教务登录成功，正在读取学期列表",
  connecting: copy(
    "copy.admin.components.admin.course-sync-panel.12dc5b9fe571"
  ),
  waiting: copy("copy.admin.components.admin.course-sync-panel.472fd78a1ff5"),
  scanned: copy("copy.admin.components.admin.course-sync-panel.6247b1d462a9"),
  ready: copy("copy.admin.components.admin.course-sync-panel.e652a73dc7e2"),
  running: copy("copy.admin.components.admin.course-sync-panel.e9ed02e8e728"),
  done: copy("copy.admin.components.admin.course-sync-panel.678b02618ee7"),
  failed: copy("copy.admin.components.admin.course-sync-panel.762ac5ed92a2"),
  cancelled: copy("copy.admin.components.admin.course-sync-panel.6f96c2add214"),
}

export function CourseSyncPanel() {
  const methods = useImportSettings()
  const [job, setJob] = useState(() => sessionStorage.getItem(storageKey) ?? "")
  const [excluded, setExcluded] = useState<string[]>([])
  const [fresh, setFresh] = useState(false)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState("")
  const client = useQueryClient()
  const query = useQuery({
    queryKey: ["course-sync", job],
    queryFn: () => apiClient<SyncStatus>(`/api/admin/course-sync/${job}`),
    enabled: !!job,
    refetchInterval: (q) => (terminal(q.state.data?.state) ? false : 2000),
    retry: 3,
  })
  const status = query.data
  const refreshed = useRef("")
  const savedCount = status?.results.length ?? 0
  const statusID = status?.id
  useEffect(() => {
    if (!statusID || savedCount === 0) return
    const marker = `${statusID}:${savedCount}`
    if (refreshed.current === marker) return
    refreshed.current = marker
    void invalidateCourseDataQueries(client)
  }, [client, statusID, savedCount])
  const selected =
    status?.semesters.filter((semester) => !excluded.includes(semester)) ?? []
  async function action(kind: "start" | "run" | "cancel") {
    setBusy(true)
    setError("")
    try {
      const value = await apiClient<SyncStatus>(
        kind === "start"
          ? "/api/admin/course-sync"
          : `/api/admin/course-sync/${job}/${kind}`,
        {
          method: "POST",
          body:
            kind === "run"
              ? JSON.stringify({ semesters: selected, fresh })
              : undefined,
        }
      )
      client.setQueryData(["course-sync", value.id], value)
      setJob(value.id)
      sessionStorage.setItem(storageKey, value.id)
      if (kind === "start") setExcluded([])
    } catch (err) {
      setError(getErrorMessage(err))
    } finally {
      setBusy(false)
    }
  }
  return (
    <section className="space-y-4">
      <p className="text-sm text-muted-foreground">
        {copy("copy.admin.components.admin.course-sync-panel.2cda2b1766b5")}
      </p>
      {methods.admin && (!job || terminal(status?.state) || query.isError) && (
        <Button
          data-action="connect"
          disabled={busy}
          onClick={() => void action("start")}
        >
          {busy
            ? copy("copy.admin.components.admin.course-sync-panel.a898478be181")
            : copy(
                "copy.admin.components.admin.course-sync-panel.49364f276b5d"
              )}
        </Button>
      )}
      {!methods.admin && (
        <p className="text-sm text-muted-foreground">
          教务同步已关闭，已提交的任务可继续完成。
        </p>
      )}
      {status && (
        <div className="space-y-4 rounded-lg border p-4" aria-live="polite">
          <h4 className="font-medium">{states[status.state]}</h4>
          {status.state === "queued" && (
            <p className="text-sm text-muted-foreground">
              前面还有 {status.queue_position ?? 0}{" "}
              个任务。可取消排队；二维码显示后再扫码。
            </p>
          )}
          <SchoolLoginFeedback state={status.state} />
          {status.qr && (
            <img
              src={status.qr}
              alt={copy(
                "copy.admin.components.admin.course-sync-panel.6dddbe2fc74d"
              )}
              className="size-56 rounded-md bg-white p-2"
            />
          )}
          {status.state === "waiting" && (
            <p className="text-sm text-muted-foreground">
              {copy(
                "copy.admin.components.admin.course-sync-panel.e62f52e25742"
              )}
            </p>
          )}
          {status.state === "ready" && (
            <>
              <label className="flex items-center gap-2 text-sm">
                <Input
                  type="checkbox"
                  className="size-4"
                  checked={fresh}
                  onChange={(event) => setFresh(event.target.checked)}
                />
                重新采集所选学期（不使用已有缓存）
              </label>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setExcluded([])}
                >
                  {copy(
                    "copy.admin.components.admin.course-sync-panel.dcac5f13ba0c"
                  )}
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setExcluded(status.semesters)}
                >
                  {copy(
                    "copy.admin.components.admin.course-sync-panel.a3c974892464"
                  )}
                </Button>
              </div>
              <div className="grid max-h-80 gap-2 overflow-y-auto sm:grid-cols-3">
                {status.semesters.map((semester) => (
                  <label
                    key={semester}
                    className="flex items-center gap-2 rounded-md border p-3"
                  >
                    <Input
                      type="checkbox"
                      className="size-4"
                      checked={!excluded.includes(semester)}
                      onChange={(event) =>
                        setExcluded(
                          event.target.checked
                            ? excluded.filter((item) => item !== semester)
                            : [...excluded, semester]
                        )
                      }
                    />
                    {semester}
                  </label>
                ))}
              </div>
              <Button
                data-action="sync"
                disabled={busy || !selected.length}
                onClick={() => void action("run")}
              >
                {copy(
                  "copy.admin.components.admin.course-sync-panel.bfe85df36865"
                )}
                {selected.length}
                {copy(
                  "copy.admin.components.admin.course-sync-panel.a0e3b90a2506"
                )}
              </Button>
            </>
          )}
          {status.state === "running" && (
            <p className="text-sm">
              当前学期 {status.current || "准备中"} · 已保存{" "}
              {status.results.length} / {status.selected.length} 个学期
            </p>
          )}
          {status.message && (
            <p role="alert" className="text-sm text-destructive">
              {status.message}
            </p>
          )}
          {(status.state === "running" || status.progress) && (
            <CourseSyncProgress
              progress={status.progress}
              startedAt={status.started_at}
              active={status.state === "running"}
            />
          )}
          {status.results.length > 0 && (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-sm">
                <thead>
                  <tr>
                    <th className="p-2">
                      {copy(
                        "copy.admin.components.admin.course-sync-panel.9b3c6793d920"
                      )}
                    </th>
                    <th className="p-2">
                      {copy(
                        "copy.admin.components.admin.course-sync-panel.c4d02889fa04"
                      )}
                    </th>
                    <th className="p-2">
                      {copy(
                        "copy.admin.components.admin.course-sync-panel.46a9b00fe9a9"
                      )}
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {status.results.map((item) => (
                    <tr key={item.semester} className="border-t">
                      <td className="p-2">{item.semester}</td>
                      <td className="p-2">{item.course_count}</td>
                      <td className="p-2">{item.section_count}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
          {!terminal(status.state) && (
            <Button
              data-action="cancel"
              variant="outline"
              disabled={busy}
              onClick={() => void action("cancel")}
            >
              {copy(
                "copy.admin.components.admin.course-sync-panel.537d17f1c531"
              )}
            </Button>
          )}
        </div>
      )}
      {query.isError && (
        <p role="alert" className="text-sm text-destructive">
          {copy("copy.admin.components.admin.course-sync-panel.462fae248797")}
        </p>
      )}
      {error && (
        <p role="alert" className="text-sm text-destructive">
          {error}
        </p>
      )}
      <p className="text-sm text-muted-foreground">
        采集进度自动缓存 24
        小时。中断后重新扫码、选择相同学期即可续采；完整学期通过校验后才保存，已发布的点评保留。缓存不包含学校登录态。
      </p>
    </section>
  )
}
