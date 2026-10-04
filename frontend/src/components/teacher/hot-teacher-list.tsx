import { copy } from "@/lib/copy"
import { Link } from "@tanstack/react-router"
import { useHotTeachers } from "@/hooks/use-teacher"
import { Skeleton } from "@/components/ui/skeleton"

export function HotTeacherList({ period }: { period: "week" | "month" }) {
  const { data, isLoading, isError } = useHotTeachers(period)
  if (isLoading)
    return (
      <div className="space-y-3">
        {Array.from({ length: 5 }, (_, i) => (
          <Skeleton key={i} className="h-16 w-full" />
        ))}
      </div>
    )
  if (isError)
    return (
      <p className="py-12 text-center text-muted-foreground">
        {copy("copy.user.components.teacher.hot-teacher-list.9890ce184f1a")}
      </p>
    )
  if (!data?.items.length)
    return (
      <p className="py-12 text-center text-muted-foreground">
        {copy("copy.user.components.teacher.hot-teacher-list.c6cdb3ae74ad")}
      </p>
    )
  return (
    <div>
      <div className="flex gap-3 border-b py-2 text-sm text-muted-foreground">
        <span>
          {copy("copy.user.components.teacher.hot-teacher-list.6570a98f2850")}
        </span>
        <span className="flex-1">
          {copy("copy.user.components.teacher.hot-teacher-list.a6654120cbec")}
        </span>
        <span>
          {copy("copy.user.components.teacher.hot-teacher-list.095f31ef0d81")}
        </span>
      </div>
      {data.items.map((item, i) => (
        <Link
          key={item.teacher.id}
          to="/teacher/$teacherID"
          params={{ teacherID: String(item.teacher.id) }}
          className="flex items-center gap-3 border-b py-4 hover:bg-muted/50"
        >
          <span
            className={`w-7 text-center font-bold ${i < 3 ? "text-amber-500" : "text-muted-foreground"}`}
          >
            {i + 1}
          </span>
          <div className="min-w-0 flex-1">
            <p className="font-medium">{item.teacher.name}</p>
            <p className="text-sm text-muted-foreground">
              {item.teacher.department} · {item.course_count}
              {copy(
                "copy.user.components.teacher.hot-teacher-list.aa79ce5c8b7e"
              )}
            </p>
          </div>
          <span className="text-sm font-medium">{item.score}</span>
        </Link>
      ))}
    </div>
  )
}
