import { copy } from "@/lib/copy"
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
  feedbackMailto?: string
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
            <span>
              {copy(
                "copy.user.components.course.course-detail-overview.6322fa286117"
              )}
            </span>
            <span className="font-medium">{course.department}</span>
          </div>

          {course.target_years && course.target_years.length > 0 && (
            <section className="flex flex-wrap items-center gap-2 text-sm">
              <h2 className="text-sm text-muted-foreground">
                {copy(
                  "copy.user.components.course.course-detail-overview.3ca3a8a1fb00"
                )}
              </h2>
              {course.target_years.map((targetYear) => (
                <CourseBadge key={targetYear} kind="targetYear">
                  {targetYear}
                </CourseBadge>
              ))}
            </section>
          )}

          {teacherGroup.length > 1 && (
            <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1 text-sm text-muted-foreground">
              <span>
                {copy(
                  "copy.user.components.course.course-detail-overview.a182f1486ada"
                )}
              </span>
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
              <h2 className="text-sm text-muted-foreground">
                {copy(
                  "copy.user.components.course.course-detail-overview.32090f3541c7"
                )}
              </h2>
              {courseSemesters.map((semester) => (
                <CourseSemesterBadge key={semester} semester={semester} />
              ))}
            </section>
          )}

          {course.offered_courses.length > 0 && (
            <section className="space-y-3">
              <h2 className="text-sm font-medium">
                {copy(
                  "copy.user.components.course.course-detail-overview.7ff8048129b6"
                )}
              </h2>
              {course.offered_courses.map((offering) => (
                <article
                  key={`${offering.id}-${offering.course_id}`}
                  className="space-y-3 rounded-lg border p-3"
                >
                  <div className="flex flex-wrap items-center gap-2">
                    <CourseSemesterBadge semester={offering.semester} />
                    <span className="text-xs text-muted-foreground">
                      {offering.sections.length}
                      {copy(
                        "copy.user.components.course.course-detail-overview.8a76809069a1"
                      )}
                    </span>
                  </div>
                  {offering.sections.map((section) => (
                    <div
                      key={`${offering.id}-${section.id}`}
                      className="space-y-2 border-t pt-3 first:border-t-0 first:pt-0"
                    >
                      <div className="flex flex-wrap items-baseline gap-x-2 gap-y-1">
                        <span className="font-medium">
                          {section.course_name}
                        </span>
                        {section.lesson_no && (
                          <span className="font-mono text-xs text-muted-foreground">
                            {section.lesson_no}
                          </span>
                        )}
                        {section.teach_class && (
                          <span className="text-xs text-muted-foreground">
                            {section.teach_class}
                          </span>
                        )}
                      </div>
                      <div className="flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
                        <span>
                          {copy(
                            "copy.user.components.course.course-detail-overview.808e510c384c"
                          )}
                          {section.department}
                        </span>
                        <span>
                          {copy(
                            "copy.user.components.course.course-detail-overview.f64cba59f8f6"
                          )}
                          {section.credit}
                        </span>
                        {section.course_type && (
                          <span>
                            {copy(
                              "copy.user.components.course.course-detail-overview.8784ebab6637"
                            )}
                            {section.course_type}
                          </span>
                        )}
                        {section.language && (
                          <span>
                            {copy(
                              "copy.user.components.course.course-detail-overview.a540dbede76d"
                            )}
                            {section.language}
                          </span>
                        )}
                        {section.campus && (
                          <span>
                            {copy(
                              "copy.user.components.course.course-detail-overview.2254162264c1"
                            )}
                            {section.campus}
                          </span>
                        )}
                        {section.assessment && (
                          <span>
                            {copy(
                              "copy.user.components.course.course-detail-overview.8818c263d9b4"
                            )}
                            {section.assessment}
                          </span>
                        )}
                      </div>
                      <div className="flex flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
                        <span className="inline-flex flex-wrap items-center gap-1">
                          <span>
                            {copy(
                              "copy.user.components.course.course-detail-overview.8810575d0e7b"
                            )}
                          </span>
                          {section.teacher_profiles?.length
                            ? section.teacher_profiles.map((teacher, index) => (
                                <span key={teacher.id}>
                                  {index > 0 && <span className="mr-1">/</span>}
                                  <Link
                                    to="/teacher/$teacherID"
                                    params={{ teacherID: String(teacher.id) }}
                                    className="hover:text-primary hover:underline"
                                  >
                                    {teacher.name}
                                  </Link>
                                  <span className="ml-1">
                                    {teacher.role === "primary"
                                      ? copy(
                                          "copy.user.components.course.course-detail-overview.430ea4d0e737"
                                        )
                                      : copy(
                                          "copy.user.components.course.course-detail-overview.549cbbadcd8c"
                                        )}
                                  </span>
                                </span>
                              ))
                            : section.teachers.length > 0
                              ? section.teachers.join(" / ")
                              : copy(
                                  "copy.user.components.course.course-detail-overview.521ea9788171"
                                )}
                        </span>
                        <span>
                          {copy(
                            "copy.user.components.course.course-detail-overview.5a218b9ac440"
                          )}
                          {section.actual_count ?? "—"} /{" "}
                          {section.limit_count ?? "—"}
                        </span>
                        {section.course_period && (
                          <span>
                            {copy(
                              "copy.user.components.course.course-detail-overview.c3924adbfb99"
                            )}
                            {section.course_period}
                          </span>
                        )}
                        {section.total_hours != null && (
                          <span>
                            {copy(
                              "copy.user.components.course.course-detail-overview.879cbad53cbe"
                            )}
                            {section.total_hours}
                          </span>
                        )}
                        {section.weekly_hours != null && (
                          <span>
                            {copy(
                              "copy.user.components.course.course-detail-overview.0fd5935a1a32"
                            )}
                            {section.weekly_hours}
                          </span>
                        )}
                        {section.week_range && (
                          <span>
                            {copy(
                              "copy.user.components.course.course-detail-overview.7de8c3e515b1"
                            )}
                            {section.week_range}
                          </span>
                        )}
                      </div>
                      {(section.remark || section.syllabus_url) && (
                        <div className="flex flex-wrap items-center gap-3 text-xs">
                          {section.remark && (
                            <span className="text-muted-foreground">
                              {copy(
                                "copy.user.components.course.course-detail-overview.73226759270d"
                              )}
                              {section.remark}
                            </span>
                          )}
                          {section.syllabus_url && (
                            <a
                              href={section.syllabus_url}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="text-primary hover:underline"
                            >
                              {copy(
                                "copy.user.components.course.course-detail-overview.11f49fad3a76"
                              )}
                            </a>
                          )}
                        </div>
                      )}
                    </div>
                  ))}
                </article>
              ))}
            </section>
          )}

          <CourseModeratorRemark course={course} />

          {selectedSemesters.length > 0 && (
            <section className="flex flex-wrap items-center gap-2 text-sm">
              <h2 className="text-sm text-muted-foreground">
                {copy(
                  "copy.user.components.course.course-detail-overview.653fa47404b1"
                )}
              </h2>
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
                  {copy(
                    "copy.user.components.course.course-detail-overview.fa2ff3c379a7"
                  )}
                </Link>
              </Button>
            </section>
          )}

          <div className="flex flex-wrap items-center gap-2">
            <CourseNotificationControl
              courseID={course.id}
              level={course.notification_level}
            />
            {feedbackMailto && (
              <Button
                asChild
                size="sm"
                variant="outline"
                className="h-8 px-2 text-muted-foreground hover:text-foreground"
              >
                <a href={feedbackMailto}>
                  <RiMailLine data-icon="inline-start" />
                  {copy(
                    "copy.user.components.course.course-detail-overview.8b2106ca1371"
                  )}
                </a>
              </Button>
            )}
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
