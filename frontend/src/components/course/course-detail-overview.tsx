import { Link } from "@tanstack/react-router"
import { RiMailLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { CourseHeaderMeta } from "./course-header-meta"
import { CourseNotificationControl } from "./course-notification-control"
import { RatingDistribution } from "./rating-distribution"
import { CourseBadge, CourseBadges, CourseSemesterBadge } from "./course-badges"
import {
  CourseModeratorRemark,
  CourseModeratorRemarkButton,
} from "./course-detail-moderator-remark"
import type { CourseDetailDTO } from "@/api/course"

interface CourseDetailOverviewProps {
  course: CourseDetailDTO
  courseSemesters: string[]
  selectedSemesters: string[]
  feedbackMailto: string
  isAdmin: boolean
}

export function CourseDetailOverview({
  course,
  courseSemesters,
  selectedSemesters,
  feedbackMailto,
  isAdmin,
}: CourseDetailOverviewProps) {
  const teacherGroup = course.teacher_group ?? []

  return (
    <div className="space-y-4">
      <CourseHeaderMeta course={course} />

      <div className="space-y-6 md:grid md:grid-cols-[minmax(0,1fr)_minmax(18rem,min(24rem,50%))] md:items-start md:gap-6 md:space-y-0">
        <div className="ml-2 space-y-4">
          <CourseBadges
            credit={course.credit}
            language={course.language}
            categories={course.categories}
          />

          <div className="flex flex-wrap items-center gap-2 text-sm text-muted-foreground">
            <span>开课单位</span>
            <span className="font-medium">{course.department}</span>
          </div>

          {course.target_years && course.target_years.length > 0 && (
            <section className="flex flex-wrap items-center gap-2 text-sm">
              <h2 className="text-sm text-muted-foreground">面向年级</h2>
              {course.target_years.map((targetYear) => (
                <CourseBadge key={targetYear} kind="targetYear">
                  {targetYear}
                </CourseBadge>
              ))}
            </section>
          )}

          {teacherGroup.length > 1 && (
            <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1 text-sm text-muted-foreground">
              <span>合上教师</span>
              {teacherGroup.map((teacher, index) => (
                <span key={teacher.id} className="inline-flex items-baseline">
                  {index > 0 && <span className="mr-2">/</span>}
                  <Link
                    to="/teacher/$teacherID"
                    params={{ teacherID: String(teacher.id) }}
                    className="font-medium hover:text-primary hover:underline"
                  >
                    {teacher.name}
                  </Link>
                </span>
              ))}
            </div>
          )}

          {courseSemesters.length > 0 && (
            <section className="flex flex-wrap items-center gap-2 text-sm">
              <h2 className="text-sm text-muted-foreground">开课学期</h2>
              {courseSemesters.map((semester) => (
                <CourseSemesterBadge key={semester} semester={semester} />
              ))}
            </section>
          )}

          <CourseModeratorRemark course={course} />

          {selectedSemesters.length > 0 && (
            <section className="flex flex-wrap items-center gap-2 text-sm">
              <h2 className="text-sm text-muted-foreground">已选学期</h2>
              {selectedSemesters.map((semester) => (
                <CourseSemesterBadge key={semester} semester={semester} />
              ))}
              <Button
                asChild
                variant="ghost"
                size="sm"
                className="h-7 px-2 text-muted-foreground hover:text-foreground"
              >
                <Link to="/course/mine" search={{ type: "enrolled" }}>
                  查看全部
                </Link>
              </Button>
            </section>
          )}

          <div className="flex flex-wrap items-center gap-2">
            <CourseNotificationControl
              courseID={course.id}
              level={course.notification_level}
            />
            <Button
              asChild
              size="sm"
              variant="outline"
              className="h-8 px-2 text-muted-foreground hover:text-foreground"
            >
              <a href={feedbackMailto}>
                <RiMailLine data-icon="inline-start" />
                反馈
              </a>
            </Button>
            {isAdmin && <CourseModeratorRemarkButton course={course} />}
          </div>
        </div>

        <div className="md:self-start md:border-l md:pl-6">
          <Card className="shadow-none ring-0">
            <CardContent className="py-0">
              <RatingDistribution rating={course.rating} />
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}
