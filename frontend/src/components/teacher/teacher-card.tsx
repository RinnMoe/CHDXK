import { Link } from "@tanstack/react-router"
import { TitleBadge } from "@/components/ui/title-badge"
import { displayTeacherTitle } from "@/lib/utils"
import type { TeacherDTO } from "@/api/teacher"

interface TeacherCardProps {
  teacher: TeacherDTO
}

export function TeacherCard({ teacher }: TeacherCardProps) {
  const teacherTitle = displayTeacherTitle(teacher.title)

  return (
    <Link
      to="/teacher/$teacherID"
      params={{ teacherID: String(teacher.id) }}
      className="block border-b p-2 transition-colors hover:bg-muted/40 focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none"
    >
      <div className="space-y-1">
        <div className="flex flex-wrap items-center gap-1.5">
          <span className="leading-tight font-semibold">{teacher.name}</span>
          {teacherTitle && <TitleBadge>{teacherTitle}</TitleBadge>}
        </div>
        <div className="text-sm text-muted-foreground">
          {teacher.department}
        </div>
      </div>
    </Link>
  )
}
