import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { apiClient } from "@/api/client"
import {
  getTeacherFilters,
  type TeacherDTO,
  type TeacherProfileDTO,
} from "@/api/teacher"
import type { PaginatedResult } from "@/api/types"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog"

import { teacherProfileForm, fillTeacherProfile } from "./teacher-profile-form"

const fields = ["学位", "联系方式", "办公地点", "研究方向", "个人简介"]
const base = "/api/admin/teacher-profiles"
export function TeacherMaintenance({ disabled }: { disabled: boolean }) {
  const [q, setQ] = useState("")
  const [department, setDepartment] = useState("")
  const [title, setTitle] = useState("")
  const [page, setPage] = useState(1)
  const [editing, setEditing] = useState<TeacherDTO | null>(null)
  const params = new URLSearchParams({
    q,
    department,
    title,
    page: String(page),
    page_size: "20",
  })
  const list = useQuery({
    queryKey: ["teacher-maintenance", q, department, title, page],
    queryFn: () =>
      apiClient<PaginatedResult<TeacherDTO>>(`${base}/teachers?${params}`),
  })
  const filters = useQuery({
    queryKey: ["teacher-filters"],
    queryFn: getTeacherFilters,
  })
  return (
    <div className="space-y-3">
      <p className="text-sm text-muted-foreground">
        维护教师姓名、学院及展示资料。手动保存的资料不会被自动同步覆盖，留空的字段可从详情页查看缺失提示。
      </p>
      <div className="overflow-x-auto rounded-lg border">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b text-left">
              <th className="p-3">姓名</th>
              <th className="p-3">学院</th>
              <th className="p-3">职称</th>
              <th className="p-3">操作</th>
            </tr>
            <tr className="border-b">
              <th className="p-2">
                <Input
                  aria-label="筛选教师姓名"
                  value={q}
                  onChange={(e) => {
                    setQ(e.target.value)
                    setPage(1)
                  }}
                  placeholder="姓名或拼音"
                />
              </th>
              <th className="p-2">
                <select
                  className="h-9 w-full rounded-md border bg-background px-2 font-normal"
                  aria-label="筛选教师学院"
                  value={department}
                  onChange={(e) => {
                    setDepartment(e.target.value)
                    setPage(1)
                  }}
                >
                  <option value="">全部学院</option>
                  {filters.data?.departments?.map((d) => (
                    <option key={d.name} value={d.name}>
                      {d.name}
                    </option>
                  ))}
                </select>
              </th>
              <th className="p-2">
                <select
                  className="h-9 w-full rounded-md border bg-background px-2 font-normal"
                  aria-label="筛选教师职称"
                  value={title}
                  onChange={(e) => {
                    setTitle(e.target.value)
                    setPage(1)
                  }}
                >
                  <option value="">全部职称</option>
                  {filters.data?.titles?.map((d) => (
                    <option key={d.name} value={d.name}>
                      {d.name}
                    </option>
                  ))}
                </select>
              </th>
              <th />
            </tr>
          </thead>
          <tbody>
            {list.data?.items.map((t) => (
              <tr key={t.id} className="border-b last:border-0">
                <td className="p-3 font-medium">{t.name}</td>
                <td className="p-3">{t.department}</td>
                <td className="p-3">{t.title || "—"}</td>
                <td className="p-3">
                  <Button
                    size="sm"
                    variant="outline"
                    data-edit-teacher={t.id}
                    disabled={disabled}
                    onClick={() => setEditing(t)}
                  >
                    编辑
                  </Button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {list.isPending && <p>正在读取</p>}
      {list.isError && (
        <p role="alert" className="text-destructive">
          教师列表读取失败，请重试
        </p>
      )}
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span className="text-sm text-muted-foreground">
          共 {list.data?.total ?? 0} 位教师 · 第 {page} 页
        </span>
        <div className="flex gap-2">
          <Button
            variant="outline"
            disabled={page === 1 || list.isPending}
            onClick={() => setPage(page - 1)}
          >
            上一页
          </Button>
          <Button
            variant="outline"
            disabled={list.isPending || page * 20 >= (list.data?.total ?? 0)}
            onClick={() => setPage(page + 1)}
          >
            下一页
          </Button>
        </div>
      </div>
      <Dialog
        open={!!editing}
        onOpenChange={(open) => {
          if (!open) setEditing(null)
        }}
      >
        {editing && (
          <TeacherEditor
            key={editing.id}
            teacher={editing}
            close={() => setEditing(null)}
          />
        )}
      </Dialog>
    </div>
  )
}
function TeacherEditor({
  teacher,
  close,
}: {
  teacher: TeacherDTO
  close: () => void
}) {
  const [form, setForm] = useState(() => teacherProfileForm(teacher))
  const [feedback, setFeedback] = useState("")
  const [candidates, setCandidates] = useState<TeacherProfileDTO[]>([])
  const fill = useMutation({
    mutationFn: (url: string) =>
      apiClient<{
        candidates: TeacherProfileDTO[]
        profile?: TeacherProfileDTO
      }>(`${base}/${teacher.id}/preview`, {
        method: "POST",
        body: JSON.stringify({
          name: form.name,
          department: form.department,
          url,
        }),
      }),
    onMutate: () => {
      setFeedback("")
      setCandidates([])
    },
    onSuccess: (result) => {
      if (result.profile) {
        const p = result.profile
        setForm((current) => fillTeacherProfile(current, p))
        setCandidates([])
        setFeedback("资料已填入，请核对后保存。来源缺失的字段保留现有内容。")
      } else if (result.candidates.length > 1) {
        setCandidates(result.candidates)
        setFeedback("找到同学院同名的教师，请选择要填入的资料。")
      } else {
        setFeedback("未找到姓名与学院都匹配的教师主页，现有内容保留。")
      }
    },
  })
  const client = useQueryClient()
  const save = useMutation({
    mutationFn: () =>
      apiClient(`${base}/${teacher.id}/details`, {
        method: "PUT",
        body: JSON.stringify(form),
      }),
    onSuccess: () => {
      for (const key of [
        "teacher-maintenance",
        "teachers",
        "teacher",
        "teacher-filters",
        "teacher-profile-matches",
        "courses",
        "course",
        "teacher-courses",
        "hot-teachers",
        "hot-courses",
        "course-enrollments",
        "followed-courses",
        "ignored-courses",
      ])
        void client.invalidateQueries({ queryKey: [key] })
      close()
    },
  })
  return (
    <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-2xl">
      <DialogHeader>
        <DialogTitle>编辑教师资料</DialogTitle>
        <DialogDescription>
          保存后用于本站展示，自动同步不会覆盖手动维护的资料。
        </DialogDescription>
      </DialogHeader>
      <div className="space-y-3 rounded-lg border bg-muted/30 p-3">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <span className="text-sm font-medium">从学校教师主页填充</span>
          <Button
            type="button"
            variant="outline"
            size="sm"
            data-action="fill-teacher-profile"
            disabled={
              fill.isPending ||
              save.isPending ||
              !form.name.trim() ||
              !form.department.trim()
            }
            onClick={() => fill.mutate("")}
          >
            {fill.isPending ? "正在获取资料…" : "获取并填入资料"}
          </Button>
        </div>
        <p className="text-xs text-muted-foreground">
          根据当前填写的姓名与学院查询，仅填入这位教师的资料，核对后点击保存。
        </p>
        {feedback && (
          <p className="text-sm" role="status">
            {feedback}
          </p>
        )}
        {fill.isError && (
          <p role="alert" className="text-sm text-destructive">
            {fill.error.message}
          </p>
        )}
        {candidates.map((p) => (
          <div
            key={p.url}
            className="flex flex-wrap items-center justify-between gap-2 rounded-md border p-3 text-sm"
          >
            <div>
              <p>
                {p.name} · {p.department}
                {p.title ? ` · ${p.title}` : ""}
              </p>
              <a
                className="text-primary underline underline-offset-4"
                href={p.url}
                target="_blank"
                rel="noreferrer"
              >
                查看学校主页
              </a>
            </div>
            <Button
              type="button"
              size="sm"
              variant="outline"
              data-fill-candidate={p.url}
              disabled={fill.isPending || save.isPending}
              onClick={() => fill.mutate(p.url)}
            >
              填入此教师资料
            </Button>
          </div>
        ))}
      </div>
      <form
        onSubmit={(e) => {
          e.preventDefault()
          save.mutate()
        }}
        className="space-y-4"
      >
        <fieldset
          disabled={fill.isPending || save.isPending}
          className="space-y-4"
        >
          <div className="grid gap-4 sm:grid-cols-2">
            {(["name", "department", "title", "url"] as const).map((key) => (
              <label key={key} className="space-y-2 text-sm">
                <span>
                  {
                    {
                      name: "姓名",
                      department: "学院",
                      title: "职称",
                      url: "教师主页",
                    }[key]
                  }
                </span>
                <Input
                  name={key}
                  required={key === "name" || key === "department"}
                  type={key === "url" ? "url" : "text"}
                  value={form[key]}
                  onChange={(e) => {
                    setForm({ ...form, [key]: e.target.value })
                    if (key === "name" || key === "department") {
                      setCandidates([])
                      setFeedback("")
                      fill.reset()
                    }
                  }}
                  maxLength={key === "url" ? 2000 : 100}
                />
              </label>
            ))}
          </div>
          {fields.map((label) => (
            <label key={label} className="block space-y-2 text-sm">
              <span>{label}</span>
              <Textarea
                name={label}
                value={form.fields[label]}
                rows={label === "个人简介" ? 5 : 2}
                maxLength={10000}
                onChange={(e) =>
                  setForm({
                    ...form,
                    fields: { ...form.fields, [label]: e.target.value },
                  })
                }
              />
            </label>
          ))}
        </fieldset>
        {save.isError && (
          <p role="alert" className="text-sm text-destructive">
            {save.error.message}
          </p>
        )}
        <div className="flex justify-end gap-2">
          <Button
            type="button"
            variant="outline"
            disabled={save.isPending}
            onClick={close}
          >
            取消
          </Button>
          <Button type="submit" disabled={save.isPending || fill.isPending}>
            保存资料
          </Button>
        </div>
      </form>
    </DialogContent>
  )
}
