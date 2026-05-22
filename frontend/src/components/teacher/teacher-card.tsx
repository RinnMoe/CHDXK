import { Link } from "react-router-dom"
import { Card, CardContent } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import type { TeacherDTO } from "@/api/teacher"

interface TeacherCardProps {
  teacher: TeacherDTO
}

export function TeacherCard({ teacher }: TeacherCardProps) {
  return (
    <Link to={`/teachers/${teacher.id}`} className="block">
      <Card className="transition-shadow hover:shadow-md hover:border-primary/30">
        <CardContent className="pt-4 space-y-2">
          <div className="flex items-start justify-between gap-2">
            <div className="min-w-0">
              <div className="font-semibold">{teacher.name}</div>
              {teacher.title && (
                <div className="text-sm text-muted-foreground">
                  {teacher.title}
                </div>
              )}
            </div>
          </div>
          <div className="text-sm text-muted-foreground font-mono">
            {teacher.code}
          </div>
          <Badge variant="outline" className="text-xs">
            {teacher.department}
          </Badge>
        </CardContent>
      </Card>
    </Link>
  )
}
