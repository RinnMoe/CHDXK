import { TeacherCard } from "./teacher-card"
import { Skeleton } from "@/components/ui/skeleton"
import type { TeacherDTO } from "@/api/teacher"

interface TeacherListProps {
  teachers: TeacherDTO[]
  isLoading?: boolean
}

function TeacherCardSkeleton() {
  return (
    <div className="rounded-lg border p-4 space-y-2">
      <Skeleton className="h-5 w-24" />
      <Skeleton className="h-4 w-16" />
      <Skeleton className="h-4 w-full" />
    </div>
  )
}

export function TeacherList({ teachers, isLoading }: TeacherListProps) {
  if (isLoading) {
    return (
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 6 }).map((_, i) => (
          <TeacherCardSkeleton key={i} />
        ))}
      </div>
    )
  }

  if (teachers.length === 0) {
    return (
      <div className="py-12 text-center">
        <p className="text-sm text-muted-foreground">暂无教师</p>
      </div>
    )
  }

  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
      {teachers.map((teacher) => (
        <TeacherCard key={teacher.id} teacher={teacher} />
      ))}
    </div>
  )
}
