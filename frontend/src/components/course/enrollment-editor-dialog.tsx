import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { RiEditLine } from "@remixicon/react"
import {
  getCourseDetail,
  listCourses,
  type CourseEnrollmentDTO,
  type CourseListItemDTO,
} from "@/api/course"
import { apiClient } from "@/api/client"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"

export function EnrollmentEditorDialog({
  enrollment,
}: {
  enrollment: CourseEnrollmentDTO
}) {
  const [open, setOpen] = useState(false)
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button
          variant="ghost"
          size="icon-sm"
          aria-label="编辑选课记录"
          data-action="edit-enrollment"
        >
          <RiEditLine />
        </Button>
      </DialogTrigger>
      <DialogContent className="max-h-[85vh] overflow-y-auto sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>编辑选课记录</DialogTitle>
          <DialogDescription>
            选择本站已收录的课程、学期和教学班。只修改你的选课记录。
          </DialogDescription>
        </DialogHeader>
        {open && (
          <EnrollmentEditor
            enrollment={enrollment}
            onSaved={() => setOpen(false)}
          />
        )}
      </DialogContent>
    </Dialog>
  )
}

export function EnrollmentEditor({
  enrollment,
  onSaved,
}: {
  enrollment?: CourseEnrollmentDTO
  onSaved: () => void
}) {
  const client = useQueryClient()
  const [search, setSearch] = useState("")
  const [term, setTerm] = useState("")
  const [course, setCourse] = useState<CourseListItemDTO | undefined>(
    enrollment?.course
  )
  const [semester, setSemester] = useState(enrollment?.semester ?? "")
  const [section, setSection] = useState("")
  const results = useQuery({
    queryKey: ["enrollment-course-search", term],
    queryFn: () => listCourses({ q: term, page_size: 20 }),
    enabled: !!term,
  })
  const detail = useQuery({
    queryKey: ["course", course?.id],
    queryFn: () => getCourseDetail(course!.id),
    enabled: !!course,
  })
  const offerings =
    detail.data?.offered_courses.filter(
      (offering) => offering.sections.length > 0
    ) ?? []
  const semesters = [...new Set(offerings.map((offering) => offering.semester))]
    .sort()
    .reverse()
  const sections = offerings
    .filter((offering) => offering.semester === semester)
    .flatMap((offering) => offering.sections)
  const save = useMutation({
    mutationFn: () =>
      apiClient(
        enrollment
          ? `/api/course/enrollment/${enrollment.id}`
          : "/api/course/enrollment",
        {
          method: enrollment ? "PUT" : "POST",
          body: JSON.stringify({ section_id: Number(section) }),
        }
      ),
    onSuccess: async () => {
      await Promise.all([
        client.invalidateQueries({ queryKey: ["course-enrollments"] }),
        client.invalidateQueries({ queryKey: ["course"] }),
      ])
      onSaved()
    },
  })
  return (
    <form
      className="space-y-4"
      onSubmit={(e) => {
        e.preventDefault()
        if (section && sections.some((item) => item.id === Number(section)))
          save.mutate()
      }}
    >
      <div className="space-y-2">
        <label
          htmlFor="enrollment-course-search"
          className="text-sm font-medium"
        >
          查找课程
        </label>
        <div className="flex gap-2">
          <Input
            id="enrollment-course-search"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="输入课程名称或课程代码"
          />
          <Button
            type="button"
            variant="outline"
            onClick={() => {
              if (search.trim() === term) void results.refetch()
              else setTerm(search.trim())
            }}
            disabled={!search.trim() || results.isFetching}
          >
            搜索
          </Button>
        </div>
        {results.isError && (
          <p role="alert" className="text-sm text-destructive">
            课程搜索失败，请重试
          </p>
        )}
        {results.data && (
          <div
            className="max-h-52 overflow-auto rounded-md border"
            aria-label="课程搜索结果"
          >
            {!results.data.items.length && (
              <p className="p-3 text-sm text-muted-foreground">
                没有找到课程，请更换名称或代码
              </p>
            )}
            {results.data.items.map((item) => (
              <button
                key={item.id}
                type="button"
                className="block w-full border-b p-3 text-left text-sm last:border-0 hover:bg-muted"
                data-course-id={item.id}
                onClick={() => {
                  setCourse(item)
                  setTerm("")
                  setSemester("")
                  setSection("")
                  save.reset()
                }}
              >
                <span className="font-medium">{item.name}</span>
                <span className="ml-2 text-muted-foreground">
                  {item.code} · {item.main_teacher.name} · {item.department}
                </span>
              </button>
            ))}
          </div>
        )}
      </div>
      {course && (
        <>
          <p className="rounded-md bg-muted p-3 text-sm">
            已选择：{course.name} · {course.main_teacher.name}
          </p>
          {detail.isPending && (
            <p className="text-sm text-muted-foreground">正在读取教学班</p>
          )}
          {detail.isError && (
            <p role="alert" className="text-sm text-destructive">
              教学班读取失败，请重新选择课程
            </p>
          )}
          {detail.data && !offerings.length && (
            <p className="text-sm text-muted-foreground">
              此课程暂无可选教学班，请选择其他课程
            </p>
          )}
          <label className="block space-y-2 text-sm font-medium">
            学期
            <select
              className="block h-9 w-full rounded-md border bg-background px-3 font-normal"
              value={semester}
              onChange={(e) => {
                setSemester(e.target.value)
                setSection("")
                save.reset()
              }}
              required
              data-field="enrollment-semester"
            >
              <option value="">请选择学期</option>
              {semesters.map((value) => (
                <option key={value} value={value}>
                  {value}
                </option>
              ))}
            </select>
          </label>
          <label className="block space-y-2 text-sm font-medium">
            教学班
            <select
              className="block h-9 w-full rounded-md border bg-background px-3 font-normal"
              value={section}
              onChange={(e) => {
                setSection(e.target.value)
                save.reset()
              }}
              required
              disabled={!semester}
              data-field="enrollment-section"
            >
              <option value="">请选择教学班</option>
              {sections.map((item) => (
                <option key={item.id} value={item.id}>
                  {[item.lesson_no, item.teach_class, item.teachers.join("、")]
                    .filter(Boolean)
                    .join(" · ") || "教学班"}
                </option>
              ))}
            </select>
          </label>
        </>
      )}
      {save.isError && (
        <p role="alert" className="text-sm text-destructive">
          {save.error.message}
        </p>
      )}
      <Button
        type="submit"
        disabled={!section || save.isPending}
        data-action="save-enrollment"
      >
        {save.isPending ? "正在保存" : "保存"}
      </Button>
    </form>
  )
}
