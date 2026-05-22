import { Link } from "react-router-dom"
import { TitleBadge } from "@/components/ui/title-badge"
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
      <div className="space-y-1">
        <div className="flex items-center gap-1.5">
          <span className="shrink-0 font-mono text-sm text-muted-foreground">
            {teacher.code}
          </span>
          {teacher.title && <TitleBadge>{teacher.title}</TitleBadge>}
        </div>
        <div className="font-semibold leading-tight">{teacher.name}</div>
        <div className="text-sm text-muted-foreground">{teacher.department}</div>
      </div>
    </Link>
  )
}
