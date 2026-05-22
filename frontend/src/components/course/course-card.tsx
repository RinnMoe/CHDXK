import { Link } from "react-router-dom"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Separator } from "@/components/ui/separator"
import { RatingDisplay } from "./rating-display"
import type { CourseListItemDTO } from "@/api/course"

interface CourseCardProps {
  course: CourseListItemDTO
}

export function CourseCard({ course }: CourseCardProps) {
  return (
    <Link to={`/courses/${course.id}`} className="block">
      <Card className="transition-shadow hover:shadow-md hover:border-primary/30">
        <CardHeader className="pb-2">
          <div className="flex items-start justify-between gap-2">
            <div className="min-w-0">
              <div className="text-xs text-muted-foreground font-mono">
                {course.code}
              </div>
              <div className="font-semibold leading-tight truncate">
                {course.name}
              </div>
            </div>
            <Badge variant="secondary" className="shrink-0 text-xs">
              {course.credit} 学分
            </Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-2">
          <div className="text-sm text-muted-foreground">
            {course.main_teacher.name}
            {course.main_teacher.title && (
              <span className="ml-1">({course.main_teacher.title})</span>
            )}
          </div>

          <div className="flex flex-wrap gap-1">
            <Badge variant="outline" className="text-xs">
              {course.language}
            </Badge>
            {course.categories.slice(0, 2).map((cat) => (
              <Badge key={cat} variant="outline" className="text-xs">
                {cat}
              </Badge>
            ))}
            {course.target_years.slice(0, 2).map((y) => (
              <Badge key={y} variant="outline" className="text-xs">
                {y}
              </Badge>
            ))}
          </div>

          <Separator />

          <RatingDisplay rating={course.rating} size="sm" />
        </CardContent>
      </Card>
    </Link>
  )
}
