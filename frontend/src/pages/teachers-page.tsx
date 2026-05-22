import { useSearchParams } from "react-router-dom"
import { PageShell } from "@/components/layout/page-shell"
import { TeacherList } from "@/components/teacher/teacher-list"
import { PaginationComponent } from "@/components/common/pagination"
import { Input } from "@/components/ui/input"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { useTeacherFilters, useTeachers } from "@/hooks/use-teacher"

const ALL = "__all__"

export function TeachersPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const page = Number(searchParams.get("page") ?? "1")
  const department = searchParams.get("department") ?? ""
  const title = searchParams.get("title") ?? ""
  const pinyin = searchParams.get("pinyin") ?? ""

  const filter = { department: department || undefined, title: title || undefined, pinyin: pinyin || undefined, page, page_size: 20 }
  const { data: filters, isLoading: filtersLoading } = useTeacherFilters()
  const { data, isLoading } = useTeachers(filter)

  function update(key: string, value: string | null) {
    const next = new URLSearchParams(searchParams)
    if (!value || value === ALL) next.delete(key)
    else next.set(key, value)
    next.delete("page")
    setSearchParams(next)
  }

  function handlePageChange(p: number) {
    const next = new URLSearchParams(searchParams)
    next.set("page", String(p))
    setSearchParams(next)
  }

  return (
    <PageShell>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-bold">教师</h1>
          <p className="text-sm text-muted-foreground mt-1">
            搜索和浏览所有教师
          </p>
        </div>

        <div className="flex flex-col sm:flex-row gap-3">
          <div className="flex-1">
            <Input
              placeholder="拼音首字母或姓名搜索..."
              value={pinyin}
              onChange={(e) => update("pinyin", e.target.value)}
            />
          </div>
          <div className="w-full sm:w-48">
            <Select
              value={department || ALL}
              onValueChange={(v) => update("department", v)}
            >
              <SelectTrigger>
                <SelectValue placeholder="全部学院" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={ALL}>全部学院</SelectItem>
                {filters?.departments.map((d) => (
                  <SelectItem key={d.name} value={d.name}>
                    {d.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="w-full sm:w-40">
            <Select
              value={title || ALL}
              onValueChange={(v) => update("title", v)}
            >
              <SelectTrigger>
                <SelectValue placeholder="全部职称" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value={ALL}>全部职称</SelectItem>
                {filters?.titles.map((t) => (
                  <SelectItem key={t.name} value={t.name}>
                    {t.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>

        {data && (
          <p className="text-sm text-muted-foreground">
            共 {data.total} 位教师
          </p>
        )}

        <TeacherList teachers={data?.items ?? []} isLoading={isLoading || filtersLoading} />

        {data && data.total > 0 && (
          <div className="flex justify-center pt-4">
            <PaginationComponent
              page={data.page}
              pageSize={data.page_size}
              total={data.total}
              onPageChange={handlePageChange}
            />
          </div>
        )}
      </div>
    </PageShell>
  )
}
