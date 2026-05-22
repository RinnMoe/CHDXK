import { Link } from "react-router-dom"
import type { TeacherDTO } from "@/api/teacher"

interface TeacherCardProps {
  teacher: TeacherDTO
}

export function TeacherCard({ teacher }: TeacherCardProps) {
  return (
    <Link
      to={`/teachers/${teacher.id}`}
      className="block border-b px-4 py-3 transition-colors hover:bg-muted/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
    >
      <div className="space-y-2">
        <div className="space-y-1">
          <div className="flex min-w-0 items-center justify-between gap-2 text-xs text-muted-foreground">
            <span className="shrink-0 font-mono">
              {teacher.code}
            </span>
            {teacher.title && (
              <span className="min-w-0 truncate text-right text-sm">
                {teacher.title}
              </span>
            )}
          </div>
          <div className="font-semibold leading-tight">
            {teacher.name}
          </div>
        </div>
        <div className="text-sm text-muted-foreground">
          {teacher.department}
        </div>
      </div>
    </Link>
  )
}
