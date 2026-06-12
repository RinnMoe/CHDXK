import { useState } from "react"
import { RiEditLine } from "@remixicon/react"
import { Button } from "@/components/ui/button"
import { CourseModeratorRemarkDialog } from "./course-moderator-remark-dialog"
import type { CourseDetailDTO } from "@/api/course"

export function CourseModeratorRemark({ course }: { course: CourseDetailDTO }) {
  const trimmed = (course.moderator_remark ?? "").trim()
  if (!trimmed) return null

  return (
    <section className="rounded-md border border-primary/20 bg-primary/5 px-3 py-2 text-sm">
      <div className="whitespace-pre-wrap text-foreground/90">{trimmed}</div>
    </section>
  )
}

export function CourseModeratorRemarkButton({
  course,
}: {
  course: CourseDetailDTO
}) {
  const [open, setOpen] = useState(false)

  return (
    <>
      <Button
        type="button"
        variant="outline"
        size="sm"
        className="h-8 px-2 text-muted-foreground hover:text-foreground"
        onClick={() => setOpen(true)}
        aria-label="修改管理员备注"
        title="修改管理员备注"
      >
        <RiEditLine data-icon="inline-start" />
      </Button>
      <CourseModeratorRemarkDialog
        course={course}
        open={open}
        onOpenChange={setOpen}
      />
    </>
  )
}
