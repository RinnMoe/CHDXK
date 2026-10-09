import { useEffect, useRef, useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { apiClient } from "@/api/client"
import type { TeacherProfileDTO } from "@/api/teacher"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Link } from "@tanstack/react-router"
import { TeacherMaintenance } from "./teacher-maintenance"

interface Status {
  state: string
  stage: string
  completed: number
  total: number
  failed: number
  message: string
  cached: number
  retries: number
  waiting: number
}
interface Match {
  teacher_id: number
  name: string
  department: string
  status: string
  selected_url: string
  candidates: TeacherProfileDTO[]
}
const base = "/api/admin/teacher-profiles"
export function TeacherProfilesPanel() {
  const client = useQueryClient()
  const [search, setSearch] = useState("")
  const [fresh, setFresh] = useState(false)
  const status = useQuery({
    queryKey: ["teacher-profile-sync"],
    queryFn: () => apiClient<Status>(base),
    refetchInterval: (q) => (q.state.data?.state === "running" ? 2000 : false),
  })
  const matches = useQuery({
    queryKey: ["teacher-profile-matches"],
    queryFn: () => apiClient<Match[]>(`${base}/matches`),
  })
  const action = useMutation({
    mutationFn: (kind: "sync" | "cancel") =>
      apiClient<Status>(`${base}/${kind}`, {
        method: "POST",
        body: kind === "sync" ? JSON.stringify({ fresh }) : undefined,
      }),
    onSuccess: (value) => client.setQueryData(["teacher-profile-sync"], value),
  })
  const previous = useRef("")
  useEffect(() => {
    if (previous.current === "running" && status.data?.state !== "running") {
      void client.invalidateQueries({ queryKey: ["teacher-profile-matches"] })
      void client.invalidateQueries({ queryKey: ["teacher-maintenance"] })
      void client.invalidateQueries({ queryKey: ["teachers"] })
      void client.invalidateQueries({ queryKey: ["teacher"] })
      void client.invalidateQueries({ queryKey: ["teacher-filters"] })
    }
    previous.current = status.data?.state ?? ""
  }, [status.data?.state, client])
  const running = status.data?.state === "running"
  const items = (matches.data ?? []).filter((item) =>
    `${item.name} ${item.department}`.includes(search.trim())
  )
  return (
    <section className="space-y-4">
      <h2 className="text-lg font-semibold">教师资料</h2>
      <TeacherMaintenance disabled={running} />
      <details className="rounded-lg border p-4">
        <summary className="cursor-pointer font-medium">
          自动填充与重名处理
        </summary>
        <div className="mt-4 space-y-4">
          <p className="text-sm text-muted-foreground">
            从学校教师主页补充职称、研究方向和个人简介等资料。按姓名与学院匹配，同学院同名的教师需要在下方确认。列表隐藏空项，详情页缺失字段显示“暂无数据”。
          </p>
          <div className="flex flex-wrap gap-2">
            <Button
              disabled={running || action.isPending}
              onClick={() => action.mutate("sync")}
              data-action="sync-teacher-profiles"
            >
              批量同步教师资料
            </Button>
            {running && (
              <Button
                variant="outline"
                disabled={action.isPending}
                onClick={() => action.mutate("cancel")}
              >
                取消同步
              </Button>
            )}
          </div>
          <label className="flex items-center gap-2 text-sm">
            <Input
              type="checkbox"
              className="size-4"
              disabled={running}
              checked={fresh}
              onChange={(e) => setFresh(e.target.checked)}
            />
            重新采集全部教师资料（不使用已有缓存）
          </label>
          <p className="text-sm text-muted-foreground">
            目录页和教师主页自动缓存 24
            小时，中断后再次同步会补读未完成的资料。手工修改的资料保留。
          </p>
          {status.data && status.data.state !== "idle" && (
            <div className="rounded-lg border p-4 text-sm" aria-live="polite">
              {running
                ? `${status.data.stage === "catalog" ? "读取教师目录" : status.data.stage === "details" ? "读取详细资料" : "保存资料"}：${status.data.completed} / ${status.data.total}`
                : status.data.message}
              <p className="mt-2 text-muted-foreground">
                从缓存恢复 {status.data.cached ?? 0} 个页面，已重试{" "}
                {status.data.retries ?? 0} 次
              </p>
              {running && !!status.data.waiting && (
                <p className="mt-2 text-muted-foreground" role="status">
                  {status.data.waiting}{" "}
                  个请求正在等待自动重试，已读取的资料保留，可继续等待。
                </p>
              )}
              {!!status.data.failed && (
                <p className="mt-2 text-muted-foreground">
                  {status.data.failed}{" "}
                  个主页的详细资料未能读取，目录中的姓名和职称仍可使用，原有资料保留，详情页缺失字段显示“暂无数据”。
                </p>
              )}
            </div>
          )}
          {(status.isError || matches.isError || action.isError) && (
            <p role="alert" className="text-sm text-destructive">
              {action.error?.message ?? "教师资料读取失败，请重试"}
            </p>
          )}
          <div className="flex flex-wrap items-center justify-between gap-2">
            <h3 className="font-medium">
              重名特别处理 ·{" "}
              {matches.data?.filter((item) => item.status === "ambiguous")
                .length ?? 0}{" "}
              项待确认
            </h3>
            <Input
              className="max-w-sm"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="搜索教师姓名或学院"
              aria-label="搜索重名教师"
            />
          </div>
          {matches.isPending ? (
            <p className="text-sm text-muted-foreground">正在读取</p>
          ) : !items.length ? (
            <p className="py-8 text-center text-muted-foreground">
              暂无待处理的重名教师
            </p>
          ) : (
            items.map((item) => (
              <MatchRow key={item.teacher_id} item={item} disabled={running} />
            ))
          )}
        </div>
      </details>
    </section>
  )
}
function MatchRow({ item, disabled }: { item: Match; disabled: boolean }) {
  const client = useQueryClient()
  const [chosen, setChosen] = useState(item.selected_url)
  const save = useMutation({
    mutationFn: () =>
      apiClient(`${base}/${item.teacher_id}`, {
        method: "PUT",
        body: JSON.stringify({ url: chosen }),
      }),
    onSuccess: () => {
      void client.invalidateQueries({ queryKey: ["teacher-profile-matches"] })
      void client.invalidateQueries({ queryKey: ["teacher-maintenance"] })
      void client.invalidateQueries({ queryKey: ["teachers"] })
      void client.invalidateQueries({ queryKey: ["teacher"] })
      void client.invalidateQueries({ queryKey: ["teacher-filters"] })
    },
  })
  return (
    <article
      className="space-y-3 rounded-lg border p-4"
      data-teacher-id={item.teacher_id}
    >
      <div className="flex flex-wrap items-center justify-between gap-2">
        <Link
          to="/teacher/$teacherID"
          params={{ teacherID: String(item.teacher_id) }}
          className="font-medium text-primary"
        >
          {item.name} · {item.department}
        </Link>
        <span className="text-sm text-muted-foreground">
          {item.status === "manual" ? "已人工确认" : "待确认"}
        </span>
      </div>
      <p className="text-sm text-muted-foreground">
        请比较学校主页的履历、研究方向与本站任课记录。同名教师的资料不会自动合并；无法确认时保持待处理。
      </p>
      <div className="space-y-2">
        {item.candidates.map((p) => (
          <label
            key={p.url}
            className="flex items-start gap-3 rounded-md border p-3"
          >
            <input
              type="radio"
              name={`teacher-${item.teacher_id}`}
              value={p.url}
              checked={chosen === p.url}
              onChange={() => setChosen(p.url)}
              disabled={disabled || save.isPending}
            />
            <div className="min-w-0 space-y-1 text-sm">
              <p>
                {p.name} · {p.department} · {p.title || "暂无数据"}
              </p>
              <p className="text-muted-foreground">
                {(
                  p.fields["研究领域"] ||
                  p.fields["研究方向"] ||
                  "研究方向：暂无数据"
                ).slice(0, 300)}
              </p>
              <a
                href={p.url}
                target="_blank"
                rel="noreferrer"
                className="text-primary underline"
              >
                查看学校主页
              </a>
            </div>
          </label>
        ))}
      </div>
      <Button
        disabled={
          disabled || !chosen || save.isPending || chosen === item.selected_url
        }
        onClick={() => save.mutate()}
      >
        确认关联
      </Button>
      {save.isError && (
        <p role="alert" className="text-sm text-destructive">
          {save.error.message}
        </p>
      )}
    </article>
  )
}
