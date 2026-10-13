import { useEffect, useRef, useState } from "react"
import { useQuery, useQueryClient } from "@tanstack/react-query"
import { apiClient } from "@/api/client"
import { useImportSettings } from "@/hooks/use-import-settings"
import { useAuth } from "@/contexts/auth-context"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { SchoolLoginFeedback } from "./school-login-feedback"
import { EnrollmentEditor } from "./enrollment-editor-dialog"
import { recentEnrollmentSemesters } from "./enrollment-semesters"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
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
  queue_position?: number
  qr?: string
  semesters: string[]
  selected: string[]
  current?: string
  cached?: number
  warning?: string
  message?: string
  progress?: { retries: number; cached?: number; retry_at?: string }
  results: {
    semester: string
    enrollment: {
      matched: number
      added: number
      unmatched: { lesson_no: string; name: string }[]
      unmatched_count: number
    }
  }[]
}
const terminal = (state?: string) =>
  ["done", "failed", "cancelled"].includes(state ?? "")
const states: Record<string, string> = {
  queued: "正在排队，轮到后会显示登录二维码",
  connecting: "正在连接教务",
  waiting: "请用微信扫码并确认学校账号登录",
  scanned: "已扫码，请在手机上确认登录",
  authenticating: "手机已确认，正在连接本科教务",
  discovering: "教务登录成功，正在读取学期列表",
  ready: "已连接，请选择要导入的学期",
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
  const methods = useImportSettings()
  const [method, setMethod] = useState("school")
  const [open, setOpen] = useState(false)
  const storageKey = `chdxk.enrollment-import.${userID}`
  const [job, setJob] = useState(() => sessionStorage.getItem(storageKey) ?? "")
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState("")
  const [fresh, setFresh] = useState(false)
  const [selection, setSelection] = useState<{
    job: string
    semesters: string[]
  }>()
  const client = useQueryClient()
  const query = useQuery({
    queryKey: ["enrollment-import", userID, job],
    queryFn: () =>
      apiClient<EnrollmentSyncStatus>(`/api/course/enrollment-sync/${job}`),
    enabled: !!job,
    refetchInterval: (q) => (terminal(q.state.data?.state) ? false : 2000),
    retry: 3,
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
    void client.invalidateQueries({ queryKey: ["course-review-prompts"] })
    void client.invalidateQueries({ queryKey: ["course"] })
  }, [client, marker])
  const selected =
    selection?.job === status?.id
      ? (selection?.semesters ?? [])
      : status?.selected.length
        ? status.selected.filter((semester) =>
            status.semesters.includes(semester)
          )
        : recentEnrollmentSemesters(status?.semesters ?? [])
  function selectSemesters(semesters: string[]) {
    if (status) setSelection({ job: status.id, semesters })
  }
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
              ? JSON.stringify({ semesters: selected, fresh })
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
  const schoolVisible =
    methods.school || (!!job && !terminal(status?.state) && !query.isError)
  const available = [
    schoolVisible && "school",
    methods.manual && "manual",
    methods.mychu && "mychu",
  ].filter(Boolean) as string[]
  const activeMethod = available.includes(method)
    ? method
    : (available[0] ?? "")
  if (!available.length) return null
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button variant="outline" data-action="import-my-courses">
          导入课程
        </Button>
      </DialogTrigger>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>导入课程</DialogTitle>
          <DialogDescription>
            选择可用的导入方式。已有选课记录不会重复添加。
          </DialogDescription>
        </DialogHeader>
        <Tabs value={activeMethod} onValueChange={setMethod}>
          <TabsList aria-label="课程导入方式">
            {schoolVisible && (
              <TabsTrigger value="school" data-method="school">
                教务扫码
              </TabsTrigger>
            )}
            {methods.manual && (
              <TabsTrigger value="manual" data-method="manual">
                手动添加
              </TabsTrigger>
            )}
            {methods.mychu && (
              <TabsTrigger value="mychu" data-method="mychu">
                MyCHU
              </TabsTrigger>
            )}
          </TabsList>
          <TabsContent value="school">
            <section className="space-y-4" aria-label="教务扫码导入">
              <h3 className="font-medium">教务扫码导入</h3>
              {methods.school &&
                (!job || terminal(status?.state) || query.isError) && (
                  <Button
                    disabled={busy}
                    onClick={() => void action("start")}
                    data-action="connect"
                  >
                    扫码连接教务
                  </Button>
                )}
              {status && (
                <div
                  className="space-y-3 rounded-lg border p-4"
                  aria-live="polite"
                >
                  <p className="font-medium">
                    {states[status.state] ?? "正在处理"}
                  </p>
                  <SchoolLoginFeedback state={status.state} />
                  {status.state === "queued" && (
                    <p className="text-sm text-muted-foreground">
                      前面还有 {status.queue_position ?? 0}{" "}
                      个任务。可取消排队；二维码显示后再扫码。
                    </p>
                  )}
                  {status.qr && (
                    <img
                      src={status.qr}
                      alt="学校登录二维码"
                      className="mx-auto size-56"
                    />
                  )}
                  {status.state === "ready" && (
                    <>
                      <label className="flex items-center gap-2 text-sm">
                        <Input
                          type="checkbox"
                          className="size-4"
                          checked={fresh}
                          onChange={(e) => setFresh(e.target.checked)}
                        />
                        重新读取所选学期（不使用已有缓存）
                      </label>
                      <p className="text-sm text-muted-foreground">
                        默认选择最近 5
                        个学年，减少空学期查询。可按实际就读时间调整；需要更早的课程时选择全部学期。
                      </p>
                      <div className="flex flex-wrap gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() =>
                            selectSemesters(
                              recentEnrollmentSemesters(status.semesters)
                            )
                          }
                        >
                          最近 5 个学年
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => selectSemesters(status.semesters)}
                        >
                          全部学期
                        </Button>
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => selectSemesters([])}
                        >
                          清空选择
                        </Button>
                      </div>
                      <div className="grid grid-cols-2 gap-2">
                        {status.semesters.map((semester) => (
                          <label
                            key={semester}
                            className="flex items-center gap-2 text-sm"
                          >
                            <Input
                              type="checkbox"
                              className="size-4"
                              checked={selected.includes(semester)}
                              onChange={(e) =>
                                selectSemesters(
                                  e.target.checked
                                    ? status.semesters.filter(
                                        (s) =>
                                          s === semester || selected.includes(s)
                                      )
                                    : selected.filter((s) => s !== semester)
                                )
                              }
                            />
                            {semester}
                          </label>
                        ))}
                      </div>
                      <Button
                        disabled={busy || !selected.length}
                        onClick={() => void action("run")}
                        data-action="run"
                      >
                        导入所选学期（{selected.length}）
                      </Button>
                    </>
                  )}
                  {status.state === "running" && (
                    <p>
                      正在读取 {status.current || "课表"}，已完成{" "}
                      {status.results.length} / {status.selected.length}{" "}
                      个学期。
                    </p>
                  )}
                  {status.state === "running" && !!status.progress?.retries && (
                    <p className="text-sm text-muted-foreground">
                      已自动重试 {status.progress.retries}{" "}
                      次，暂时的响应异常会持续等待重试，无需重新扫码。
                    </p>
                  )}
                  {!!status.cached && (
                    <p className="text-sm text-muted-foreground">
                      已从缓存恢复 {status.cached} 门课程。
                    </p>
                  )}
                  {status.warning && (
                    <p role="status" className="text-sm text-warning">
                      {status.warning}
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
                        {result.semester}：匹配 {result.enrollment.matched}{" "}
                        门，新增 {result.enrollment.added} 条记录。
                      </p>
                      {!!result.enrollment.unmatched_count && (
                        <details>
                          <summary className="cursor-pointer">
                            {result.enrollment.unmatched_count} 门课程暂未匹配
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
                          {result.enrollment.unmatched_count >
                            result.enrollment.unmatched.length && (
                            <p className="mt-2 text-muted-foreground">
                              这里只展示前 20 门未匹配课程。
                            </p>
                          )}
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
                仅导入个人课程，不修改教务数据。已读取的学期缓存 24
                小时，再次扫码同一学校账号可继续导入。完成、取消或超时后关闭本次连接，下次仍需扫码。关闭弹窗后导入继续进行。
              </p>
            </section>
          </TabsContent>
          <TabsContent value="manual" className="space-y-4">
            <p className="text-sm text-muted-foreground">
              选择本站已收录的课程、学期和教学班，添加到你的选课记录。
            </p>
            <EnrollmentEditor onSaved={() => setOpen(false)} />
          </TabsContent>
          <TabsContent value="mychu">
            <section
              className="space-y-2 border-t pt-4"
              aria-label="MyCHU 导入"
            >
              <h3 className="font-medium">通过 MyCHU 应用导入</h3>
              <p className="text-sm text-muted-foreground">
                暂未开放。后续可从 MyCHU 将课程导入本站。
              </p>
              <Button variant="outline" disabled>
                MyCHU 导入（暂未开放）
              </Button>
            </section>
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  )
}
