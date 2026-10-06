import { useEffect, useRef, useState } from "react"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import { apiClient } from "@/api/client"
import { useAuth } from "@/contexts/auth-context"
import { Button } from "@/components/ui/button"
import { SchoolLoginFeedback } from "./school-login-feedback"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"

export interface EnrollmentSyncStatus {
  id: string
  state: string
  qr?: string
  semesters: string[]
  selected: string[]
  current?: string
  message?: string
  progress?: { retries: number }
  results: {
    semester: string
    enrollment: {
      matched: number
      added: number
      unmatched: { lesson_no: string; name: string }[]
    }
  }[]
}
const terminal = (state?: string) =>
  ["done", "failed", "cancelled"].includes(state ?? "")
const states: Record<string, string> = {
  connecting: "正在连接教务",
  waiting: "请用微信扫码并确认学校账号登录",
  scanned: "已扫码，请在手机上确认登录",
  authenticating: "手机已确认，正在连接本科教务",
  discovering: "教务登录成功，正在读取学期列表",
  ready: "已连接，可导入全部学期的课程",
  running: "正在读取并导入课程",
  done: "课程导入完成",
  failed: "课程导入已停止",
  cancelled: "已取消课程导入",
}

export function EnrollmentImportDialog() {
  const { user } = useAuth()
  return user ? <ImportDialog key={user.id} userID={user.id} /> : null
}

function ImportDialog({ userID }: { userID: number }) {
  const storageKey = `chdxk.enrollment-import.${userID}`
  const [job, setJob] = useState(() => sessionStorage.getItem(storageKey) ?? "")
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState("")
  const client = useQueryClient()
  const query = useQuery({
    queryKey: ["enrollment-import", userID, job],
    queryFn: () =>
      apiClient<EnrollmentSyncStatus>(`/api/course/enrollment-sync/${job}`),
    enabled: !!job,
    refetchInterval: (q) =>
      terminal(q.state.data?.state) || q.state.error ? false : 2000,
    retry: false,
  })
  const status = query.data
  const refreshed = useRef("")
  const marker = status?.results.length
    ? `${status.id}:${status.results.length}`
    : ""
  useEffect(() => {
    if (!marker || refreshed.current === marker) return
    refreshed.current = marker
    void client.invalidateQueries({ queryKey: ["course-enrollments"] })
    void client.invalidateQueries({ queryKey: ["course"] })
  }, [client, marker])
  const selected = status?.semesters ?? []
  async function action(kind: "start" | "run" | "cancel") {
    setBusy(true)
    setError("")
    try {
      const value = await apiClient<EnrollmentSyncStatus>(
        kind === "start"
          ? "/api/course/enrollment-sync"
          : `/api/course/enrollment-sync/${job}/${kind}`,
        {
          method: "POST",
          body:
            kind === "run"
              ? JSON.stringify({ semesters: selected })
              : undefined,
        }
      )
      client.setQueryData(["enrollment-import", userID, value.id], value)
      setJob(value.id)
      sessionStorage.setItem(storageKey, value.id)
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "导入暂时无法完成，请稍后重试"
      )
    } finally {
      setBusy(false)
    }
  }
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button variant="outline" data-action="import-my-courses">
          导入我的课程
        </Button>
      </DialogTrigger>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>导入我的课程</DialogTitle>
          <DialogDescription>
            扫码登录本科教务，将全部可用学期的个人课程添加到选课记录。已有记录不会重复添加。
          </DialogDescription>
        </DialogHeader>
        <section className="space-y-4" aria-label="教务扫码导入">
          <h3 className="font-medium">教务扫码导入</h3>
          {(!job || terminal(status?.state) || query.isError) && (
            <Button
              disabled={busy}
              onClick={() => void action("start")}
              data-action="connect"
            >
              扫码连接教务
            </Button>
          )}
          {status && (
            <div className="space-y-3 rounded-lg border p-4" aria-live="polite">
              <p className="font-medium">
                {states[status.state] ?? "正在处理"}
              </p>
              <SchoolLoginFeedback state={status.state} />
              {status.qr && (
                <img
                  src={status.qr}
                  alt="学校登录二维码"
                  className="mx-auto size-56"
                />
              )}
              {status.state === "ready" && (
                <>
                  <div className="grid grid-cols-2 gap-2">
                    {status.semesters.map((semester) => (
                      <span
                        key={semester}
                        className="text-sm text-muted-foreground"
                      >
                        {semester}
                      </span>
                    ))}
                  </div>
                  <Button
                    disabled={busy || !selected.length}
                    onClick={() => void action("run")}
                    data-action="run"
                  >
                    导入全部学期（{selected.length}）
                  </Button>
                </>
              )}
              {status.state === "running" && (
                <p>
                  正在读取 {status.current || "课表"}，已完成{" "}
                  {status.results.length} / {status.selected.length} 个学期。
                </p>
              )}
              {status.state === "running" && !!status.progress?.retries && (
                <p className="text-sm text-muted-foreground">
                  教务响应暂时异常，正在自动重试（{status.progress.retries} /
                  2），无需重新扫码。
                </p>
              )}
              {status.message && (
                <p className="text-destructive" role="alert">
                  {status.message}
                </p>
              )}
              {status.results.map((result) => (
                <div key={result.semester} className="space-y-2 text-sm">
                  <p>
                    {result.semester}：匹配 {result.enrollment.matched} 门，新增{" "}
                    {result.enrollment.added} 条记录。
                  </p>
                  {!!result.enrollment.unmatched.length && (
                    <details>
                      <summary className="cursor-pointer">
                        {result.enrollment.unmatched.length} 门课程暂未匹配
                      </summary>
                      <p className="my-2 text-muted-foreground">
                        这些课程尚未加入选课记录，请联系管理员补充对应学期的开课记录后再次导入。
                      </p>
                      <ul className="space-y-1">
                        {result.enrollment.unmatched.map((item, i) => (
                          <li key={`${item.lesson_no}:${i}`}>
                            {item.name}
                            {item.lesson_no ? `（${item.lesson_no}）` : ""}
                          </li>
                        ))}
                      </ul>
                    </details>
                  )}
                </div>
              ))}
              {!terminal(status.state) && (
                <Button
                  variant="outline"
                  disabled={busy}
                  onClick={() => void action("cancel")}
                  data-action="cancel"
                >
                  取消导入
                </Button>
              )}
            </div>
          )}
          {(error || query.isError) && (
            <p role="alert" className="text-sm text-destructive">
              {error || "原导入连接已失效，请重新扫码连接教务"}
            </p>
          )}
          <p className="text-sm text-muted-foreground">
            仅导入个人课程，不修改教务数据。完成、取消或超时后关闭本次教务连接，下次导入需要重新扫码。关闭弹窗后本次导入继续进行。
          </p>
        </section>
        <section className="space-y-2 border-t pt-4" aria-label="MyCHU 导入">
          <h3 className="font-medium">通过 MyCHU 应用导入</h3>
          <p className="text-sm text-muted-foreground">
            暂未开放。后续可从 MyCHU 将课程导入本站。
          </p>
          <Button variant="outline" disabled>
            MyCHU 导入（暂未开放）
          </Button>
        </section>
      </DialogContent>
    </Dialog>
  )
}
